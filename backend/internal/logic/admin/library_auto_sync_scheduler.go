package admin

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"ms_tmdb/internal/model"
	"ms_tmdb/internal/svc"
	"ms_tmdb/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	defaultAutoSyncBatchSize      = 50
	maxAutoSyncBatchSize          = 500
	autoSyncLoopTick              = 10 * time.Second
	defaultAutoSyncCronExpression = "*/30 * * * *"
	autoSyncLogStatusSuccess      = "success"
	autoSyncLogStatusPartial      = "partial_failed"
	autoSyncLogStatusPanic        = "panic"
)

var (
	autoSyncSchedulerMu sync.RWMutex
	autoSyncScheduler   *LibraryAutoSyncScheduler
)

type AutoSyncSettings struct {
	Enabled          bool
	CronExpr         string
	Mode             string
	BatchSize        int
	StartDelaySecond int
	Running          bool
}

type LibraryAutoSyncScheduler struct {
	svcCtx *svc.ServiceContext

	mu                sync.RWMutex
	settings          AutoSyncSettings
	running           bool
	started           bool
	eligibleAt        time.Time
	lastCronRunMinute time.Time
	cronMatcher       *cronMatcher
}

type autoSyncStats struct {
	checked int
	synced  int
	failed  int
}

func SetLibraryAutoSyncScheduler(scheduler *LibraryAutoSyncScheduler) {
	autoSyncSchedulerMu.Lock()
	defer autoSyncSchedulerMu.Unlock()
	autoSyncScheduler = scheduler
}

func GetLibraryAutoSyncScheduler() *LibraryAutoSyncScheduler {
	autoSyncSchedulerMu.RLock()
	defer autoSyncSchedulerMu.RUnlock()
	return autoSyncScheduler
}

func NewLibraryAutoSyncScheduler(svcCtx *svc.ServiceContext) *LibraryAutoSyncScheduler {
	cfg := svcCtx.Config.Tmdb.AutoSync
	settings := normalizeAutoSyncSettings(AutoSyncSettings{
		Enabled:          cfg.Enabled,
		CronExpr:         cfg.CronExpr,
		Mode:             cfg.Mode,
		BatchSize:        cfg.BatchSize,
		StartDelaySecond: cfg.StartDelaySecond,
	})

	matcher, err := buildCronMatcher(settings)
	if err != nil {
		logx.Errorf("TMDB 自动同步 cron 配置无效，使用默认值: %v", err)
		settings.CronExpr = defaultAutoSyncCronExpression
		matcher, _ = buildCronMatcher(settings)
	}

	return &LibraryAutoSyncScheduler{
		svcCtx:            svcCtx,
		settings:          settings,
		running:           false,
		started:           false,
		eligibleAt:        time.Time{},
		lastCronRunMinute: time.Time{},
		cronMatcher:       matcher,
	}
}

func (s *LibraryAutoSyncScheduler) Start() {
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return
	}
	s.started = true
	settings := s.settings
	if settings.Enabled {
		s.eligibleAt = time.Now().Add(time.Duration(settings.StartDelaySecond) * time.Second)
	}
	eligibleAt := s.eligibleAt
	s.mu.Unlock()

	logx.Infof(
		"TMDB 自动同步调度器启动: enabled=%t, cron=%s, mode=%s, batch_size=%d, start_delay=%ds, eligible_at=%s",
		settings.Enabled,
		settings.CronExpr,
		settings.Mode,
		settings.BatchSize,
		settings.StartDelaySecond,
		formatTime(eligibleAt),
	)

	go func() {
		ticker := time.NewTicker(autoSyncLoopTick)
		defer ticker.Stop()
		for {
			s.maybeRun()
			<-ticker.C
		}
	}()
}

func (s *LibraryAutoSyncScheduler) GetSettings() AutoSyncSettings {
	s.mu.RLock()
	defer s.mu.RUnlock()

	settings := s.settings
	settings.Running = s.running
	return settings
}

func (s *LibraryAutoSyncScheduler) UpdateSettings(input AutoSyncSettings) (AutoSyncSettings, error) {
	settings := normalizeAutoSyncSettings(input)
	matcher, err := buildCronMatcher(settings)
	if err != nil {
		return AutoSyncSettings{}, err
	}

	s.mu.Lock()
	old := s.settings
	s.settings = settings
	s.cronMatcher = matcher
	s.lastCronRunMinute = time.Time{}

	if !settings.Enabled {
		s.eligibleAt = time.Time{}
	} else if !old.Enabled {
		s.eligibleAt = time.Now().Add(time.Duration(settings.StartDelaySecond) * time.Second)
	}

	eligibleAt := s.eligibleAt
	running := s.running
	s.mu.Unlock()

	logx.Infof(
		"TMDB 自动同步配置已更新: enabled=%t, cron=%s, mode=%s, batch_size=%d, start_delay=%ds, eligible_at=%s",
		settings.Enabled,
		settings.CronExpr,
		settings.Mode,
		settings.BatchSize,
		settings.StartDelaySecond,
		formatTime(eligibleAt),
	)

	settings.Running = running
	return settings, nil
}

func (s *LibraryAutoSyncScheduler) maybeRun() {
	s.mu.Lock()
	if !s.settings.Enabled || s.running || s.cronMatcher == nil {
		s.mu.Unlock()
		return
	}

	now := time.Now()
	if !s.eligibleAt.IsZero() && now.Before(s.eligibleAt) {
		s.mu.Unlock()
		return
	}

	minuteKey := now.Truncate(time.Minute)
	if s.lastCronRunMinute.Equal(minuteKey) || !s.cronMatcher.Match(now) {
		s.mu.Unlock()
		return
	}

	settings := s.settings
	s.running = true
	s.lastCronRunMinute = minuteKey
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
	}()

	s.runOnce(settings, minuteKey)
}

func (s *LibraryAutoSyncScheduler) runOnce(settings AutoSyncSettings, triggeredAt time.Time) {
	start := time.Now()
	status := autoSyncLogStatusSuccess
	message := "自动同步执行成功"
	totalChecked := 0
	totalSynced := 0
	totalFailed := 0

	defer func() {
		finishedAt := time.Now()
		if recovered := recover(); recovered != nil {
			status = autoSyncLogStatusPanic
			message = fmt.Sprintf("任务执行异常: %v", recovered)
			totalFailed++
			logx.Errorf("TMDB 自动同步任务 panic: %v", recovered)
		} else if totalFailed > 0 {
			status = autoSyncLogStatusPartial
			message = fmt.Sprintf("执行完成，但存在 %d 项失败", totalFailed)
		}

		s.persistExecutionLog(settings, triggeredAt, start, finishedAt, status, totalChecked, totalSynced, totalFailed, message)

		logx.Infof(
			"TMDB 自动同步完成: mode=%s, checked=%d, synced=%d, failed=%d, status=%s, elapsed=%s",
			settings.Mode,
			totalChecked,
			totalSynced,
			totalFailed,
			status,
			finishedAt.Sub(start).String(),
		)
	}()

	ctx := context.Background()

	movieStats := s.syncMovies(ctx, settings)
	tvStats := s.syncTvSeries(ctx, settings)
	personStats := s.syncPeople(ctx, settings)

	totalChecked = movieStats.checked + tvStats.checked + personStats.checked
	totalSynced = movieStats.synced + tvStats.synced + personStats.synced
	totalFailed = movieStats.failed + tvStats.failed + personStats.failed
}

func (s *LibraryAutoSyncScheduler) syncMovies(ctx context.Context, settings AutoSyncSettings) autoSyncStats {
	var stats autoSyncStats

	compareLogic := NewCompareMovieRemoteLogic(ctx, s.svcCtx)
	syncLogic := NewSyncMovieLogic(ctx, s.svcCtx)

	var lastID uint
	for {
		var records []model.Movie
		query := s.svcCtx.DB.
			Model(&model.Movie{}).
			Select("id", "tmdb_id").
			Where("tmdb_id > 0").
			Order("id ASC").
			Limit(settings.BatchSize)
		if lastID > 0 {
			query = query.Where("id > ?", lastID)
		}
		if err := query.Find(&records).Error; err != nil {
			logx.Errorf("自动同步电影列表失败: %v", err)
			stats.failed++
			return stats
		}
		if len(records) == 0 {
			return stats
		}

		for _, record := range records {
			lastID = record.ID
			stats.checked++
			needSync, err := s.needSyncMovie(record.TmdbID, settings.Mode, compareLogic)
			if err != nil {
				stats.failed++
				continue
			}
			if !needSync {
				continue
			}
			if _, err := syncLogic.SyncMovie(&types.AdminSyncReq{Id: record.TmdbID, Mode: settings.Mode}); err != nil {
				stats.failed++
				logx.Errorf("自动同步电影失败: tmdb_id=%d, err=%v", record.TmdbID, err)
				continue
			}
			stats.synced++
		}
	}
}

func (s *LibraryAutoSyncScheduler) syncTvSeries(ctx context.Context, settings AutoSyncSettings) autoSyncStats {
	var stats autoSyncStats

	compareLogic := NewCompareTvRemoteLogic(ctx, s.svcCtx)
	syncLogic := NewSyncTvSeriesLogic(ctx, s.svcCtx)

	var lastID uint
	for {
		var records []model.TVSeries
		query := s.svcCtx.DB.
			Model(&model.TVSeries{}).
			Select("id", "tmdb_id").
			Where("tmdb_id > 0").
			Order("id ASC").
			Limit(settings.BatchSize)
		if lastID > 0 {
			query = query.Where("id > ?", lastID)
		}
		if err := query.Find(&records).Error; err != nil {
			logx.Errorf("自动同步剧集列表失败: %v", err)
			stats.failed++
			return stats
		}
		if len(records) == 0 {
			return stats
		}

		for _, record := range records {
			lastID = record.ID
			stats.checked++
			needSync, err := s.needSyncTv(record.TmdbID, settings.Mode, compareLogic)
			if err != nil {
				stats.failed++
				continue
			}
			if !needSync {
				continue
			}
			if _, err := syncLogic.SyncTvSeries(&types.AdminSyncReq{Id: record.TmdbID, Mode: settings.Mode}); err != nil {
				stats.failed++
				logx.Errorf("自动同步剧集失败: tmdb_id=%d, err=%v", record.TmdbID, err)
				continue
			}
			stats.synced++
		}
	}
}

func (s *LibraryAutoSyncScheduler) syncPeople(ctx context.Context, settings AutoSyncSettings) autoSyncStats {
	var stats autoSyncStats

	compareLogic := NewComparePersonRemoteLogic(ctx, s.svcCtx)
	syncLogic := NewSyncPersonLogic(ctx, s.svcCtx)

	var lastID uint
	for {
		var records []model.Person
		query := s.svcCtx.DB.
			Model(&model.Person{}).
			Select("id", "tmdb_id").
			Where("tmdb_id > 0").
			Order("id ASC").
			Limit(settings.BatchSize)
		if lastID > 0 {
			query = query.Where("id > ?", lastID)
		}
		if err := query.Find(&records).Error; err != nil {
			logx.Errorf("自动同步人物列表失败: %v", err)
			stats.failed++
			return stats
		}
		if len(records) == 0 {
			return stats
		}

		for _, record := range records {
			lastID = record.ID
			stats.checked++
			needSync, err := s.needSyncPerson(record.TmdbID, compareLogic)
			if err != nil {
				stats.failed++
				continue
			}
			if !needSync {
				continue
			}
			if _, err := syncLogic.SyncPerson(&types.AdminSyncReq{Id: record.TmdbID, Mode: settings.Mode}); err != nil {
				stats.failed++
				logx.Errorf("自动同步人物失败: tmdb_id=%d, err=%v", record.TmdbID, err)
				continue
			}
			stats.synced++
		}
	}
}

func (s *LibraryAutoSyncScheduler) needSyncMovie(tmdbID int, mode string, compareLogic *CompareMovieRemoteLogic) (bool, error) {
	resp, err := compareLogic.CompareMovieRemote(&types.AdminSyncReq{Id: tmdbID})
	if err != nil {
		logx.Errorf("检测电影远程差异失败: tmdb_id=%d, err=%v", tmdbID, err)
		return false, err
	}
	if len(resp.DiffFields) > 0 {
		return true, nil
	}
	return mode == syncModeOverwriteAll && len(resp.LocalOverrideDiffFields) > 0, nil
}

func (s *LibraryAutoSyncScheduler) needSyncTv(tmdbID int, mode string, compareLogic *CompareTvRemoteLogic) (bool, error) {
	resp, err := compareLogic.CompareTvRemote(&types.AdminSyncReq{Id: tmdbID})
	if err != nil {
		logx.Errorf("检测剧集远程差异失败: tmdb_id=%d, err=%v", tmdbID, err)
		return false, err
	}
	if len(resp.DiffFields) > 0 {
		return true, nil
	}
	return mode == syncModeOverwriteAll && len(resp.LocalOverrideDiffFields) > 0, nil
}

func (s *LibraryAutoSyncScheduler) needSyncPerson(tmdbID int, compareLogic *ComparePersonRemoteLogic) (bool, error) {
	resp, err := compareLogic.ComparePersonRemote(&types.AdminSyncReq{Id: tmdbID})
	if err != nil {
		logx.Errorf("检测人物远程差异失败: tmdb_id=%d, err=%v", tmdbID, err)
		return false, err
	}
	return len(resp.DiffFields) > 0, nil
}

func normalizeAutoSyncSettings(settings AutoSyncSettings) AutoSyncSettings {
	settings.CronExpr = strings.TrimSpace(settings.CronExpr)
	if settings.CronExpr == "" {
		settings.CronExpr = defaultAutoSyncCronExpression
	}
	if settings.BatchSize <= 0 {
		settings.BatchSize = defaultAutoSyncBatchSize
	}
	if settings.BatchSize > maxAutoSyncBatchSize {
		settings.BatchSize = maxAutoSyncBatchSize
	}
	if settings.StartDelaySecond < 0 {
		settings.StartDelaySecond = 0
	}

	settings.Mode = normalizeAutoSyncMode(settings.Mode)
	settings.Running = false
	return settings
}

func buildCronMatcher(settings AutoSyncSettings) (*cronMatcher, error) {
	matcher, err := parseCronMatcher(settings.CronExpr)
	if err != nil {
		return nil, fmt.Errorf("cron 表达式无效: %w", err)
	}
	return matcher, nil
}

func normalizeAutoSyncMode(mode string) string {
	switch normalizeSyncMode(mode) {
	case syncModeOverwriteAll:
		return syncModeOverwriteAll
	default:
		return syncModeUpdateUnchanged
	}
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format(time.RFC3339)
}

func (s *LibraryAutoSyncScheduler) persistExecutionLog(
	settings AutoSyncSettings,
	triggeredAt time.Time,
	startedAt time.Time,
	finishedAt time.Time,
	status string,
	checked int,
	synced int,
	failed int,
	message string,
) {
	record := model.AutoSyncExecutionLog{
		TriggeredAt: triggeredAt,
		CronExpr:    settings.CronExpr,
		Mode:        settings.Mode,
		BatchSize:   settings.BatchSize,
		StartedAt:   startedAt,
		FinishedAt:  finishedAt,
		DurationMs:  finishedAt.Sub(startedAt).Milliseconds(),
		Status:      status,
		Checked:     checked,
		Synced:      synced,
		Failed:      failed,
		Message:     message,
	}

	if err := s.svcCtx.DB.Create(&record).Error; err != nil {
		logx.Errorf("写入自动同步日志失败: %v", err)
	}
}
