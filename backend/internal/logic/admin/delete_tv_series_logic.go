package admin

import (
	"context"
	"errors"

	"ms_tmdb/internal/model"
	"ms_tmdb/internal/svc"
	"ms_tmdb/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteTvSeriesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteTvSeriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteTvSeriesLogic {
	return &DeleteTvSeriesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteTvSeriesLogic) DeleteTvSeries(req *types.PathIdReq) error {
	if req.Id == 0 {
		return errors.New("无效剧集 ID")
	}

	if err := l.svcCtx.DB.Where("tmdb_id = ?", req.Id).Delete(&model.TVLangSnapshot{}).Error; err != nil {
		return err
	}

	result := l.svcCtx.DB.Where("tmdb_id = ?", req.Id).Delete(&model.TVSeries{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("剧集不存在或已删除")
	}
	return nil
}
