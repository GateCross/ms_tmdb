package admin

import (
	"context"
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
	if err := l.svcCtx.TmdbClient.SetProxy(proxyURL); err != nil {
		return nil, err
	}

	return &types.AdminProxyResp{
		ProxyURL: proxyURL,
		Enabled:  proxyURL != "",
	}, nil
}
