package admin

import (
	"context"
	"errors"

	"ms_tmdb/internal/model"
	"ms_tmdb/internal/svc"
	"ms_tmdb/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type ClearMovieLocalLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewClearMovieLocalLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ClearMovieLocalLogic {
	return &ClearMovieLocalLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ClearMovieLocalLogic) ClearMovieLocal(req *types.AdminSyncReq) error {
	if req.Id <= 0 {
		return errors.New("无效电影 ID")
	}

	var movie model.Movie
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("tmdb_id = ?", req.Id).First(&movie).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("电影不存在或已删除")
		}
		return err
	}

	// 清理本地覆盖需要回到 TMDB 最新数据，否则仅清空 local_data 会留下已合并的本地字段。
	syncLogic := NewSyncMovieLogic(l.ctx, l.svcCtx)
	_, err := syncLogic.SyncMovie(&types.AdminSyncReq{
		Id:   req.Id,
		Mode: syncModeOverwriteAll,
	})
	return err
}
