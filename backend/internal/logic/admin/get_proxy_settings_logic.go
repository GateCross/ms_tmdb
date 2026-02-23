package admin

import (
	"context"

	"ms_tmdb/internal/svc"
	"ms_tmdb/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetProxySettingsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetProxySettingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProxySettingsLogic {
	return &GetProxySettingsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetProxySettingsLogic) GetProxySettings() (*types.AdminProxyResp, error) {
	current := l.svcCtx.TmdbClient.GetProxy()
	return &types.AdminProxyResp{
		ProxyURL: current,
		Enabled:  current != "",
	}, nil
}
