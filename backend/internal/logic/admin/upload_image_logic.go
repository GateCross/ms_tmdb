package admin

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ms_tmdb/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	maxUploadImageSize = 10 << 20
	uploadDirName      = "uploads"
)

var allowedImageExt = map[string]struct{}{
	".jpg":  {},
	".jpeg": {},
	".png":  {},
	".webp": {},
	".gif":  {},
}

var allowedImageContentTypes = map[string]struct{}{
	"image/jpeg": {},
	"image/png":  {},
	"image/webp": {},
	"image/gif":  {},
}

type UploadImageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUploadImageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadImageLogic {
	return &UploadImageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UploadImageLogic) UploadImage(file multipart.File, header *multipart.FileHeader) (string, error) {
	if header == nil {
		return "", errors.New("文件不能为空")
	}
	if header.Size > maxUploadImageSize {
		return "", errors.New("图片大小不能超过 10MB")
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if _, ok := allowedImageExt[ext]; !ok {
		return "", errors.New("仅支持 jpg/jpeg/png/webp/gif 图片")
	}

	contentType := strings.TrimSpace(header.Header.Get("Content-Type"))
	if contentType != "" && !strings.HasPrefix(strings.ToLower(contentType), "image/") {
		return "", errors.New("仅支持图片文件上传")
	}
	reader := bufio.NewReader(file)
	peek, err := reader.Peek(512)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, bufio.ErrBufferFull) {
		return "", err
	}
	if detected := http.DetectContentType(peek); !isAllowedImageContentType(detected) {
		return "", errors.New("文件内容不是受支持的图片格式")
	}

	uploadDir := filepath.Join(".", uploadDirName)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", err
	}

	fileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	savePath := filepath.Join(uploadDir, fileName)

	dst, err := os.Create(savePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	written, err := io.Copy(dst, io.LimitReader(reader, maxUploadImageSize+1))
	if err != nil {
		return "", err
	}
	if written > maxUploadImageSize {
		_ = os.Remove(savePath)
		return "", errors.New("图片大小不能超过 10MB")
	}

	return "/uploads/" + fileName, nil
}

func isAllowedImageContentType(contentType string) bool {
	normalized := strings.ToLower(strings.TrimSpace(contentType))
	if idx := strings.Index(normalized, ";"); idx >= 0 {
		normalized = strings.TrimSpace(normalized[:idx])
	}
	_, ok := allowedImageContentTypes[normalized]
	return ok
}
