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

type GetTvSkipSyncLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTvSkipSyncLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTvSkipSyncLogic {
	return &GetTvSkipSyncLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetTvSkipSync 查询剧集是否跳过自动同步
func (l *GetTvSkipSyncLogic) GetTvSkipSync(req *types.PathIdReq) (*types.AdminTvSkipSyncResp, error) {
	if req.Id <= 0 {
		return nil, errors.New("无效剧集 ID")
	}

	var tv model.TVSeries
	err := l.svcCtx.DB.Where("tmdb_id = ?", req.Id).First(&tv).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("剧集不存在或已删除")
	}
	if err != nil {
		return nil, err
	}
	return &types.AdminTvSkipSyncResp{
		SkipSync: tv.SkipSync,
	}, nil
}
