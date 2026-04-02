package admin

import (
	"context"
	"errors"

	"ms_tmdb/internal/model"
	"ms_tmdb/internal/svc"
	"ms_tmdb/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteMovieLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteMovieLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteMovieLogic {
	return &DeleteMovieLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteMovieLogic) DeleteMovie(req *types.PathIdReq) error {
	if req.Id == 0 {
		return errors.New("无效电影 ID")
	}

	if err := l.svcCtx.DB.Where("tmdb_id = ?", req.Id).Delete(&model.MovieLangSnapshot{}).Error; err != nil {
		return err
	}

	result := l.svcCtx.DB.Where("tmdb_id = ?", req.Id).Delete(&model.Movie{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("电影不存在或已删除")
	}
	return nil
}
