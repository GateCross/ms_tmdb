package admin

import (
	"context"
	"strings"
	"time"

	"ms_tmdb/internal/model"
	"ms_tmdb/internal/svc"
	"ms_tmdb/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAutoSyncLogsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListAutoSyncLogsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAutoSyncLogsLogic {
	return &ListAutoSyncLogsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListAutoSyncLogsLogic) ListAutoSyncLogs(req *types.AdminAutoSyncLogListReq) (*types.AdminAutoSyncLogListResp, error) {
	page, pageSize := normalizePage(req.Page, req.PageSize)

	query := l.svcCtx.DB.Model(&model.AutoSyncExecutionLog{})
	status := strings.TrimSpace(req.Status)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var records []model.AutoSyncExecutionLog
	if err := query.
		Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&records).Error; err != nil {
		return nil, err
	}

	results := make([]types.AdminAutoSyncLogItem, len(records))
	for i, item := range records {
		results[i] = types.AdminAutoSyncLogItem{
			Id:          int64(item.ID),
			TriggeredAt: formatLogTime(item.TriggeredAt),
			CronExpr:    item.CronExpr,
			Mode:        item.Mode,
			BatchSize:   item.BatchSize,
			Status:      item.Status,
			Checked:     item.Checked,
			Synced:      item.Synced,
			Failed:      item.Failed,
			DurationMs:  item.DurationMs,
			Message:     item.Message,
			StartedAt:   formatLogTime(item.StartedAt),
			FinishedAt:  formatLogTime(item.FinishedAt),
			CreatedAt:   formatLogTime(item.CreatedAt),
		}
	}

	return &types.AdminAutoSyncLogListResp{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Results:  results,
	}, nil
}

func formatLogTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}
