package admin

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ms_tmdb/internal/svc"
	"ms_tmdb/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateProxySettingsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateProxySettingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProxySettingsLogic {
	return &UpdateProxySettingsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateProxySettingsLogic) UpdateProxySettings(req *types.AdminProxyReq) (*types.AdminProxyResp, error) {
	proxyURL := strings.TrimSpace(req.ProxyURL)
	oldProxyURL := l.svcCtx.TmdbClient.GetProxy()
	if err := l.svcCtx.TmdbClient.SetProxy(proxyURL); err != nil {
		return nil, err
	}

	configFile := strings.TrimSpace(l.svcCtx.Config.ConfigFile)
	if configFile == "" {
		configFile = "etc/tmdb.yaml"
	}
	if err := writeProxyURLToConfigFile(configFile, proxyURL); err != nil {
		// 配置写入失败时回滚当前进程代理，避免“显示成功但重启丢失”。
		_ = l.svcCtx.TmdbClient.SetProxy(oldProxyURL)
		return nil, err
	}

	return &types.AdminProxyResp{
		ProxyURL: proxyURL,
		Enabled:  proxyURL != "",
	}, nil
}

func writeProxyURLToConfigFile(configPath string, proxyURL string) error {
	configFilePath := filepath.Clean(strings.TrimSpace(configPath))
	if configFilePath == "" {
		return errors.New("配置文件路径为空")
	}

	content, err := os.ReadFile(configFilePath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	raw := string(content)
	lineSep := "\n"
	if strings.Contains(raw, "\r\n") {
		lineSep = "\r\n"
		raw = strings.ReplaceAll(raw, "\r\n", "\n")
	}

	lines := strings.Split(raw, "\n")
	tmdbFound := false
	inTmdb := false
	tmdbIndent := 0
	proxyLineIndex := -1
	insertIndex := -1

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		indent := leadingIndentLen(line)

		if !inTmdb {
			if strings.HasPrefix(trimmed, "Tmdb:") {
				tmdbFound = true
				inTmdb = true
				tmdbIndent = indent
				insertIndex = i + 1
			}
			continue
		}

		if trimmed != "" && !strings.HasPrefix(trimmed, "#") && indent <= tmdbIndent {
			insertIndex = i
			break
		}
		if trimmed != "" {
			insertIndex = i + 1
		}
		if strings.HasPrefix(trimmed, "ProxyURL:") {
			proxyLineIndex = i
		}
	}

	if !tmdbFound {
		return errors.New("配置文件缺少 Tmdb 段")
	}
	if insertIndex < 0 {
		insertIndex = len(lines)
	}

	proxyLine := strings.Repeat(" ", tmdbIndent+2) + "ProxyURL: " + yamlDoubleQuoted(proxyURL)
	if proxyLineIndex >= 0 {
		lines[proxyLineIndex] = proxyLine
	} else {
		lines = append(lines[:insertIndex], append([]string{proxyLine}, lines[insertIndex:]...)...)
	}

	output := strings.Join(lines, "\n")
	if lineSep == "\r\n" {
		output = strings.ReplaceAll(output, "\n", "\r\n")
	}
	if err := os.WriteFile(configFilePath, []byte(output), 0o644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}
	return nil
}

func leadingIndentLen(line string) int {
	length := 0
	for _, ch := range line {
		if ch != ' ' && ch != '\t' {
			break
		}
		length++
	}
	return length
}

func yamlDoubleQuoted(text string) string {
	escaped := strings.ReplaceAll(text, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	return `"` + escaped + `"`
}
