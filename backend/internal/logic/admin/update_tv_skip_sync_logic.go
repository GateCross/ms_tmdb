package admin

import (
	"context"
	"errors"

	"ms_tmdb/internal/model"
	"ms_tmdb/internal/svc"
	"ms_tmdb/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateTvSkipSyncLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateTvSkipSyncLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTvSkipSyncLogic {
	return &UpdateTvSkipSyncLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UpdateTvSkipSync 标记剧集是否跳过自动同步；仅改该开关，不动 is_modified/LastSyncedAt
func (l *UpdateTvSkipSyncLogic) UpdateTvSkipSync(req *types.AdminTvSkipSyncReq) (*types.AdminTvSkipSyncResp, error) {
	if req.Id <= 0 {
		return nil, errors.New("无效剧集 ID")
	}

	result := l.svcCtx.DB.Model(&model.TVSeries{}).
		Where("tmdb_id = ?", req.Id).
		Update("skip_sync", req.SkipSync)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, errors.New("剧集不存在或已删除")
	}
	return &types.AdminTvSkipSyncResp{
		SkipSync: req.SkipSync,
		Message:  "已更新跳过自动同步状态",
	}, nil
}
