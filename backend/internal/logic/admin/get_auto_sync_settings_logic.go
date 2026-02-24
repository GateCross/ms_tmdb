package admin

import (
	"context"

	"ms_tmdb/internal/svc"
	"ms_tmdb/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAutoSyncSettingsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAutoSyncSettingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAutoSyncSettingsLogic {
	return &GetAutoSyncSettingsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAutoSyncSettingsLogic) GetAutoSyncSettings() (*types.AdminAutoSyncResp, error) {
	settings := normalizeAutoSyncSettings(AutoSyncSettings{
		Enabled:          l.svcCtx.Config.Tmdb.AutoSync.Enabled,
		CronExpr:         l.svcCtx.Config.Tmdb.AutoSync.CronExpr,
		Mode:             l.svcCtx.Config.Tmdb.AutoSync.Mode,
		BatchSize:        l.svcCtx.Config.Tmdb.AutoSync.BatchSize,
		StartDelaySecond: l.svcCtx.Config.Tmdb.AutoSync.StartDelaySecond,
	})

	if scheduler := GetLibraryAutoSyncScheduler(); scheduler != nil {
		settings = scheduler.GetSettings()
	}

	return &types.AdminAutoSyncResp{
		Enabled:          settings.Enabled,
		CronExpr:         settings.CronExpr,
		Mode:             settings.Mode,
		BatchSize:        settings.BatchSize,
		StartDelaySecond: settings.StartDelaySecond,
		Running:          settings.Running,
	}, nil
}
