package admin

import (
	"context"
	"errors"
	"strings"

	"ms_tmdb/internal/svc"
	"ms_tmdb/pkg/tmdbclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type TvSeasonLocalLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTvSeasonLocalLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TvSeasonLocalLogic {
	return &TvSeasonLocalLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TvSeasonLocalLogic) GetLocalSeason(seriesID, seasonNumber int) (map[string]interface{}, bool, error) {
	if seriesID <= 0 {
		return nil, false, errors.New("无效剧集 ID")
	}
	if seasonNumber < 0 {
		return nil, false, errors.New("无效季号")
	}
	return l.svcCtx.ProxyService.GetLocalTvSeason(seriesID, seasonNumber)
}

func (l *TvSeasonLocalLogic) SaveSeasonFromTMDB(seriesID, seasonNumber int, language, appendToResponse string) (map[string]interface{}, error) {
	if seriesID <= 0 {
		return nil, errors.New("无效剧集 ID")
	}
	if seasonNumber < 0 {
		return nil, errors.New("无效季号")
	}

	opts := &tmdbclient.RequestOption{
		Context:  l.ctx,
		Language: strings.TrimSpace(language),
	}
	if text := strings.TrimSpace(appendToResponse); text != "" {
		opts.AppendToResponse = text
	}
	return l.svcCtx.ProxyService.SaveTvSeasonToLocal(seriesID, seasonNumber, opts)
}

func (l *TvSeasonLocalLogic) UpdateLocalSeason(seriesID, seasonNumber int, payload map[string]interface{}) (map[string]interface{}, error) {
	if seriesID <= 0 {
		return nil, errors.New("无效剧集 ID")
	}
	if seasonNumber < 0 {
		return nil, errors.New("无效季号")
	}
	if len(payload) == 0 {
		return nil, errors.New("更新内容不能为空")
	}
	return l.svcCtx.ProxyService.UpdateLocalTvSeason(seriesID, seasonNumber, payload)
}
