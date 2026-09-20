package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"ms_tmdb/config"
	"ms_tmdb/internal/model"
	"ms_tmdb/pkg/tmdbpath"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

const (
	defaultRetentionDays    = 7
	defaultBodyLimitBytes   = 64 * 1024
	defaultWriteTimeout     = 3 * time.Second
	defaultCleanupBatchSize = 500
	proxyAccessLogTableName = "proxy_access_logs"
	tmdbRequestLogTableName = "tmdb_request_logs"
	autoSyncLogTableName    = "auto_sync_execution_logs"
)

// BodySnapshot 保存被截断后的正文与原始大小。
type BodySnapshot struct {
	Text      string
	Bytes     int64
	Truncated bool
}

// ProxyAccessEntry 是一次外部代理访问的落库载荷。
type ProxyAccessEntry struct {
	RequestID string
	Method    string

	Path       string
	Query      string
	RequestURI string
	ClientIP   string
	UserAgent  string

	StatusCode   int
	DurationMs   int64
	ErrorMessage string

	RequestBody  BodySnapshot
	ResponseBody BodySnapshot
}

// TmdbRequestEntry 是一次真实 TMDB 上游请求的落库载荷。
type TmdbRequestEntry struct {
	RequestID string
	Method    string

	Path string
	URL  string

	StatusCode   int
	DurationMs   int64
	ErrorMessage string

	RequestBody  BodySnapshot
	ResponseBody BodySnapshot
}

// RequestLogService 负责请求日志的截断、写入和保留期清理。
type RequestLogService struct {
	db             *gorm.DB
	retentionDays  int
	bodyLimitBytes int
	reclaimSpace   bool
}

// NewRequestLogService 创建日志服务，并归一化缺省配置。
func NewRequestLogService(db *gorm.DB, c config.TmdbLogConf) *RequestLogService {
	retentionDays := c.RetentionDays
	if retentionDays <= 0 {
		retentionDays = defaultRetentionDays
	}

	bodyLimitBytes := c.BodyLimitBytes
	if bodyLimitBytes <= 0 {
		bodyLimitBytes = defaultBodyLimitBytes
	}

	return &RequestLogService{
		db:             db,
		retentionDays:  retentionDays,
		bodyLimitBytes: bodyLimitBytes,
		// go-zero conf 的 default=true 已在加载配置时生效；这里直接透传。
		reclaimSpace: c.ReclaimSpace,
	}
}

// BodyLimitBytes 返回单个正文允许保存的最大字节数。
func (s *RequestLogService) BodyLimitBytes() int {
	if s == nil || s.bodyLimitBytes <= 0 {
		return defaultBodyLimitBytes
	}
	return s.bodyLimitBytes
}

// RetentionDays 返回当前日志保留天数。
func (s *RequestLogService) RetentionDays() int {
	if s == nil || s.retentionDays <= 0 {
		return defaultRetentionDays
	}
	return s.retentionDays
}

// SetRetentionDays 更新运行时日志保留天数，非法值回退默认。
func (s *RequestLogService) SetRetentionDays(days int) {
	if s == nil {
		return
	}
	if days <= 0 {
		days = defaultRetentionDays
	}
	s.retentionDays = days
}

// ReclaimSpaceEnabled 返回清理后是否执行磁盘空间回收。
func (s *RequestLogService) ReclaimSpaceEnabled() bool {
	return s != nil && s.reclaimSpace
}

// SetReclaimSpace 更新运行时磁盘回收开关。
func (s *RequestLogService) SetReclaimSpace(enabled bool) {
	if s == nil {
		return
	}
	s.reclaimSpace = enabled
}

// CaptureBody 按配置截断正文，数据库中只保存可读文本。
func (s *RequestLogService) CaptureBody(raw []byte) BodySnapshot {
	limit := defaultBodyLimitBytes
	if s != nil && s.bodyLimitBytes > 0 {
		limit = s.bodyLimitBytes
	}
	return CaptureBody(raw, limit)
}

// WriteProxyAccessAsync 异步写入外部访问日志，避免日志库慢时拖慢代理响应。
func (s *RequestLogService) WriteProxyAccessAsync(ctx context.Context, entry ProxyAccessEntry) {
	if s == nil || s.db == nil {
		return
	}

	go func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				logx.Errorf("写入代理访问日志异常: %v", recovered)
			}
		}()

		writeCtx, cancel := context.WithTimeout(withoutCancel(ctx), defaultWriteTimeout)
		defer cancel()

		if err := s.WriteProxyAccess(writeCtx, entry); err != nil {
			logx.Errorf("写入代理访问日志失败: %v", err)
		}
	}()
}

// CaptureBody 按指定上限截断正文。
func CaptureBody(raw []byte, limit int) BodySnapshot {
	if limit <= 0 {
		limit = defaultBodyLimitBytes
	}

	size := int64(len(raw))
	textBytes := raw
	truncated := len(raw) > limit
	if truncated {
		textBytes = raw[:limit]
	}

	return BodySnapshot{
		Text:      strings.ToValidUTF8(string(textBytes), "?"),
		Bytes:     size,
		Truncated: truncated,
	}
}

// WriteProxyAccess 写入外部访问日志。
// path 落库为 Canonical（由中间件传入）；media_type/tmdb_id 经 tmdbpath.ParseMediaTarget 解析。
func (s *RequestLogService) WriteProxyAccess(ctx context.Context, entry ProxyAccessEntry) error {
	if s == nil || s.db == nil {
		return nil
	}

	return s.db.WithContext(contextOrBackground(ctx)).Create(buildProxyAccessLog(entry)).Error
}

// buildProxyAccessLog 将访问入口载荷组装为落库模型。
// Path 已是 Canonical 时 ParseMediaTarget 同样成立；若仍带入口前缀也能正确剥离后解析。
// request_uri / query / request_id 原样保留，用于追溯原始入口。
func buildProxyAccessLog(entry ProxyAccessEntry) *model.ProxyAccessLog {
	mediaType, tmdbID := tmdbpath.ParseMediaTarget(entry.Path)
	return &model.ProxyAccessLog{
		RequestID: entry.RequestID,
		Method:    entry.Method,

		MediaType: mediaType,
		TmdbID:    tmdbID,

		Path:       entry.Path,
		Query:      entry.Query,
		RequestURI: entry.RequestURI,
		ClientIP:   entry.ClientIP,
		UserAgent:  entry.UserAgent,

		StatusCode:   entry.StatusCode,
		DurationMs:   entry.DurationMs,
		ErrorMessage: entry.ErrorMessage,

		RequestBody:           entry.RequestBody.Text,
		RequestBodyBytes:      entry.RequestBody.Bytes,
		RequestBodyTruncated:  entry.RequestBody.Truncated,
		ResponseBody:          entry.ResponseBody.Text,
		ResponseBodyBytes:     entry.ResponseBody.Bytes,
		ResponseBodyTruncated: entry.ResponseBody.Truncated,
	}
}

// WriteTmdbRequest 写入真实 TMDB 上游请求日志。
func (s *RequestLogService) WriteTmdbRequest(ctx context.Context, entry TmdbRequestEntry) error {
	if s == nil || s.db == nil {
		return nil
	}

	return s.db.WithContext(contextOrBackground(ctx)).Create(&model.TmdbRequestLog{
		RequestID: entry.RequestID,
		Method:    entry.Method,

		Path: entry.Path,
		URL:  entry.URL,

		StatusCode:   entry.StatusCode,
		DurationMs:   entry.DurationMs,
		ErrorMessage: entry.ErrorMessage,

		RequestBody:           entry.RequestBody.Text,
		RequestBodyBytes:      entry.RequestBody.Bytes,
		RequestBodyTruncated:  entry.RequestBody.Truncated,
		ResponseBody:          entry.ResponseBody.Text,
		ResponseBodyBytes:     entry.ResponseBody.Bytes,
		ResponseBodyTruncated: entry.ResponseBody.Truncated,
	}).Error
}

// CleanupExpired 物理删除超过保留期的日志，并按配置回收磁盘空间。
func (s *RequestLogService) CleanupExpired(ctx context.Context) error {
	if s == nil || s.db == nil {
		return nil
	}

	retentionDays := s.RetentionDays()
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	db := s.db.WithContext(withoutCancel(ctx))
	startedAt := time.Now()

	var totalDeleted int64
	for _, tableName := range cleanupExpiredLogTables() {
		deleted, err := cleanupExpiredRequestLogTable(db, tableName, cutoff)
		if err != nil {
			return fmt.Errorf("清理表 %s 失败: %w", tableName, err)
		}
		totalDeleted += deleted
		if deleted > 0 {
			logx.Infof("请求日志清理: table=%s deleted=%d cutoff=%s", tableName, deleted, cutoff.Format(time.RFC3339))
		}

		// 仅在本表确实删过行且开启回收时做 VACUUM FULL，避免无意义锁表。
		if deleted > 0 && s.reclaimSpace {
			if err := reclaimRequestLogTableSpace(db, tableName); err != nil {
				// 回收失败不回滚已删除数据，只记录错误并继续其他表。
				logx.Errorf("请求日志空间回收失败: table=%s err=%v", tableName, err)
				continue
			}
		}
	}

	logx.Infof(
		"请求日志清理完成: retention_days=%d deleted=%d reclaim=%t duration=%s",
		retentionDays,
		totalDeleted,
		s.reclaimSpace,
		time.Since(startedAt).Round(time.Millisecond),
	)
	return nil
}

// cleanupExpiredLogTables 返回统一保留策略覆盖的日志表。
func cleanupExpiredLogTables() []string {
	return []string{
		proxyAccessLogTableName,
		tmdbRequestLogTableName,
		autoSyncLogTableName,
	}
}

// isAllowedCleanupTable 限制动态 SQL 仅作用于白名单日志表。
func isAllowedCleanupTable(tableName string) bool {
	for _, name := range cleanupExpiredLogTables() {
		if name == tableName {
			return true
		}
	}
	return false
}

func cleanupExpiredRequestLogTable(db *gorm.DB, tableName string, cutoff time.Time) (int64, error) {
	if !isAllowedCleanupTable(tableName) {
		return 0, fmt.Errorf("不允许清理的表: %s", tableName)
	}

	var deleted int64
	for {
		var ids []uint
		if err := db.
			Table(tableName).
			Select("id").
			Where("created_at < ?", cutoff).
			Order("created_at ASC, id ASC").
			Limit(defaultCleanupBatchSize).
			Scan(&ids).Error; err != nil {
			return deleted, err
		}
		if len(ids) == 0 {
			return deleted, nil
		}

		result := db.Exec("DELETE FROM "+tableName+" WHERE id IN ?", ids)
		if result.Error != nil {
			return deleted, result.Error
		}
		deleted += result.RowsAffected
		if len(ids) < defaultCleanupBatchSize {
			return deleted, nil
		}
	}
}

// reclaimRequestLogTableSpace 通过 VACUUM FULL 把删除后的空间归还给操作系统。
// 仅允许白名单日志表；执行期间会锁表，适合低峰或可接受短暂停写的场景。
func reclaimRequestLogTableSpace(db *gorm.DB, tableName string) error {
	if !isAllowedCleanupTable(tableName) {
		return fmt.Errorf("不允许回收的表: %s", tableName)
	}

	beforeBytes, err := relationTotalBytes(db, tableName)
	if err != nil {
		return err
	}

	startedAt := time.Now()
	// PostgreSQL 标识符不能参数化，表名已通过白名单校验。
	if err := db.Exec("VACUUM (FULL, ANALYZE) " + tableName).Error; err != nil {
		return err
	}

	afterBytes, err := relationTotalBytes(db, tableName)
	if err != nil {
		return err
	}

	logx.Infof(
		"请求日志空间已回收: table=%s before=%s after=%s freed=%s duration=%s",
		tableName,
		formatBytes(beforeBytes),
		formatBytes(afterBytes),
		formatBytes(beforeBytes-afterBytes),
		time.Since(startedAt).Round(time.Millisecond),
	)
	return nil
}

func relationTotalBytes(db *gorm.DB, tableName string) (int64, error) {
	var size int64
	// 使用 to_regclass 避免直接拼接未校验标识符；调用方仍需保证白名单。
	err := db.Raw("SELECT pg_total_relation_size(to_regclass(?))", "public."+tableName).Scan(&size).Error
	if err != nil {
		return 0, err
	}
	return size, nil
}

func formatBytes(size int64) string {
	if size < 0 {
		size = 0
	}
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// StartRetentionCleaner 启动后台日志保留期清理（启动后异步执行一次，随后每 24 小时执行一次）。
func (s *RequestLogService) StartRetentionCleaner(ctx context.Context) func() {
	cleanerCtx, cancel := context.WithCancel(withoutCancel(ctx))

	go func() {
		// 启动后稍微延迟（例如 10 秒），等待 HTTP 服务完全就绪并对外提供服务后再开始初次清理
		initialTimer := time.NewTimer(10 * time.Second)
		select {
		case <-cleanerCtx.Done():
			initialTimer.Stop()
			return
		case <-initialTimer.C:
			if err := s.CleanupExpired(cleanerCtx); err != nil {
				logx.Errorf("启动后清理请求日志失败: %v", err)
			}
		}

		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-cleanerCtx.Done():
				return
			case <-ticker.C:
				if err := s.CleanupExpired(cleanerCtx); err != nil {
					logx.Errorf("定时清理请求日志失败: %v", err)
				}
			}
		}
	}()

	return cancel
}

func withoutCancel(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return context.WithoutCancel(ctx)
}

func contextOrBackground(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
