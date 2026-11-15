package model

import (
	"crypto/sha1"
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"errors"
	"reflect"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// RawJSON JSONB 字段类型
type RawJSON json.RawMessage

func (j RawJSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return []byte(j), nil
}

func (j *RawJSON) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("类型断言 []byte 失败")
	}
	*j = append((*j)[0:0], bytes...)
	return nil
}

func (j RawJSON) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return []byte(j), nil
}

func (j *RawJSON) UnmarshalJSON(data []byte) error {
	if j == nil {
		return errors.New("RawJSON: UnmarshalJSON on nil pointer")
	}
	*j = append((*j)[0:0], data...)
	return nil
}

// Movie 电影模型
type Movie struct {
	gorm.Model
	TmdbID           int        `gorm:"uniqueIndex;not null" json:"tmdb_id"`
	SyncTmdbID       int        `gorm:"index;default:0" json:"sync_tmdb_id"`
	Title            string     `gorm:"index" json:"title"`
	OriginalTitle    string     `json:"original_title"`
	Overview         string     `gorm:"type:text" json:"overview"`
	ReleaseDate      string     `json:"release_date"`
	Popularity       float64    `json:"popularity"`
	VoteAverage      float64    `json:"vote_average"`
	VoteCount        int        `json:"vote_count"`
	PosterPath       string     `json:"poster_path"`
	BackdropPath     string     `json:"backdrop_path"`
	OriginalLanguage string     `json:"original_language"`
	Adult            bool       `json:"adult"`
	Status           string     `json:"status"`
	Runtime          int        `json:"runtime"`
	Budget           int64      `json:"budget"`
	Revenue          int64      `json:"revenue"`
	Tagline          string     `json:"tagline"`
	Homepage         string     `json:"homepage"`
	ImdbID           string     `json:"imdb_id"`
	TmdbData         RawJSON    `gorm:"type:jsonb" json:"tmdb_data"`
	LocalData        RawJSON    `gorm:"type:jsonb" json:"local_data"`
	IsModified       bool       `gorm:"default:false" json:"is_modified"`
	LastSyncedAt     *time.Time `json:"last_synced_at"`
}

// MovieLangSnapshot 非默认语言电影详情快照
type MovieLangSnapshot struct {
	gorm.Model
	TmdbID       int        `gorm:"not null;index:idx_movie_lang_snapshot_tmdb_language,unique" json:"tmdb_id"`
	Language     string     `gorm:"size:16;not null;index:idx_movie_lang_snapshot_tmdb_language,unique" json:"language"`
	SyncTmdbID   int        `gorm:"index;default:0" json:"sync_tmdb_id"`
	TmdbData     RawJSON    `gorm:"type:jsonb" json:"tmdb_data"`
	LastSyncedAt *time.Time `json:"last_synced_at"`
}

// TVSeries 电视剧模型
type TVSeries struct {
	gorm.Model
	TmdbID           int        `gorm:"uniqueIndex;not null" json:"tmdb_id"`
	SyncTmdbID       int        `gorm:"index;default:0" json:"sync_tmdb_id"`
	Name             string     `gorm:"index" json:"name"`
	OriginalName     string     `json:"original_name"`
	Overview         string     `gorm:"type:text" json:"overview"`
	FirstAirDate     string     `json:"first_air_date"`
	LastAirDate      string     `json:"last_air_date"`
	Popularity       float64    `json:"popularity"`
	VoteAverage      float64    `json:"vote_average"`
	VoteCount        int        `json:"vote_count"`
	PosterPath       string     `json:"poster_path"`
	BackdropPath     string     `json:"backdrop_path"`
	OriginalLanguage string     `json:"original_language"`
	Status           string     `json:"status"`
	Type             string     `json:"type"`
	NumberOfSeasons  int        `json:"number_of_seasons"`
	NumberOfEpisodes int        `json:"number_of_episodes"`
	Homepage         string     `json:"homepage"`
	InProduction     bool       `json:"in_production"`
	Tagline          string     `json:"tagline"`
	TmdbData         RawJSON    `gorm:"type:jsonb" json:"tmdb_data"`
	LocalData        RawJSON    `gorm:"type:jsonb" json:"local_data"`
	IsModified       bool       `gorm:"default:false" json:"is_modified"`
	SkipSync         bool       `gorm:"default:false" json:"skip_sync"`
	LastSyncedAt     *time.Time `json:"last_synced_at"`
}

// TVLangSnapshot 非默认语言剧集详情快照
type TVLangSnapshot struct {
	gorm.Model
	TmdbID       int        `gorm:"not null;index:idx_tv_lang_snapshot_tmdb_language,unique" json:"tmdb_id"`
	Language     string     `gorm:"size:16;not null;index:idx_tv_lang_snapshot_tmdb_language,unique" json:"language"`
	SyncTmdbID   int        `gorm:"index;default:0" json:"sync_tmdb_id"`
	TmdbData     RawJSON    `gorm:"type:jsonb" json:"tmdb_data"`
	LastSyncedAt *time.Time `json:"last_synced_at"`
}

// Person 人物模型
type Person struct {
	gorm.Model
	TmdbID             int        `gorm:"uniqueIndex;not null" json:"tmdb_id"`
	Name               string     `gorm:"index" json:"name"`
	Biography          string     `gorm:"type:text" json:"biography"`
	Birthday           string     `json:"birthday"`
	Deathday           string     `json:"deathday"`
	Gender             int        `json:"gender"`
	KnownForDepartment string     `json:"known_for_department"`
	PlaceOfBirth       string     `json:"place_of_birth"`
	Popularity         float64    `json:"popularity"`
	ProfilePath        string     `json:"profile_path"`
	Adult              bool       `json:"adult"`
	ImdbID             string     `json:"imdb_id"`
	Homepage           string     `json:"homepage"`
	TmdbData           RawJSON    `gorm:"type:jsonb" json:"tmdb_data"`
	LocalData          RawJSON    `gorm:"type:jsonb" json:"local_data"`
	IsModified         bool       `gorm:"default:false" json:"is_modified"`
	LastSyncedAt       *time.Time `json:"last_synced_at"`
}

// PersonLangSnapshot 非默认语言人物详情快照
type PersonLangSnapshot struct {
	gorm.Model
	TmdbID       int        `gorm:"not null;index:idx_person_lang_snapshot_tmdb_language,unique" json:"tmdb_id"`
	Language     string     `gorm:"size:16;not null;index:idx_person_lang_snapshot_tmdb_language,unique" json:"language"`
	TmdbData     RawJSON    `gorm:"type:jsonb" json:"tmdb_data"`
	LastSyncedAt *time.Time `json:"last_synced_at"`
}

// AutoSyncExecutionLog 自动同步执行日志
type AutoSyncExecutionLog struct {
	gorm.Model
	TriggeredAt time.Time `gorm:"index" json:"triggered_at"`
	CronExpr    string    `gorm:"size:64;index" json:"cron_expr"`
	Mode        string    `gorm:"size:32;index" json:"mode"`
	BatchSize   int       `json:"batch_size"`

	StartedAt  time.Time `gorm:"index" json:"started_at"`
	FinishedAt time.Time `gorm:"index" json:"finished_at"`
	DurationMs int64     `json:"duration_ms"`

	Status  string  `gorm:"size:32;index" json:"status"`
	Checked int     `json:"checked"`
	Synced  int     `json:"synced"`
	Failed  int     `json:"failed"`
	Message string  `gorm:"type:text" json:"message"`
	Detail  RawJSON `gorm:"type:jsonb" json:"detail"`
}

// ProxyAccessLog 记录外部访问 TMDB 代理入口的请求与响应摘要。
type ProxyAccessLog struct {
	gorm.Model
	RequestID string `gorm:"size:64;index" json:"request_id"`
	Method    string `gorm:"size:16;index" json:"method"`

	MediaType string `gorm:"size:16" json:"media_type"`
	TmdbID    int    `json:"tmdb_id"`

	Path       string `gorm:"type:text" json:"path"`
	Query      string `gorm:"type:text" json:"query"`
	RequestURI string `gorm:"type:text" json:"request_uri"`
	ClientIP   string `gorm:"size:128;index" json:"client_ip"`
	UserAgent  string `gorm:"type:text" json:"user_agent"`

	StatusCode   int    `gorm:"index" json:"status_code"`
	DurationMs   int64  `json:"duration_ms"`
	ErrorMessage string `gorm:"type:text" json:"error_message"`

	RequestBody           string `gorm:"type:text" json:"request_body"`
	RequestBodyBytes      int64  `json:"request_body_bytes"`
	RequestBodyTruncated  bool   `json:"request_body_truncated"`
	ResponseBody          string `gorm:"type:text" json:"response_body"`
	ResponseBodyBytes     int64  `json:"response_body_bytes"`
	ResponseBodyTruncated bool   `json:"response_body_truncated"`
}

// TmdbRequestLog 记录服务端真实请求 TMDB 上游的结果。
type TmdbRequestLog struct {
	gorm.Model
	RequestID string `gorm:"size:64;index" json:"request_id"`
	Method    string `gorm:"size:16;index" json:"method"`

	Path string `gorm:"type:text" json:"path"`
	URL  string `gorm:"type:text" json:"url"`

	StatusCode   int    `gorm:"index" json:"status_code"`
	DurationMs   int64  `json:"duration_ms"`
	ErrorMessage string `gorm:"type:text" json:"error_message"`

	RequestBody           string `gorm:"type:text" json:"request_body"`
	RequestBodyBytes      int64  `json:"request_body_bytes"`
	RequestBodyTruncated  bool   `json:"request_body_truncated"`
	ResponseBody          string `gorm:"type:text" json:"response_body"`
	ResponseBodyBytes     int64  `json:"response_body_bytes"`
	ResponseBodyTruncated bool   `json:"response_body_truncated"`
}

// AutoMigrate 自动建表迁移
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(autoMigrateModels()...)
}

// RunStartupMigrations 按版本执行启动迁移，避免每次启动都触发 GORM 元数据扫描。
func RunStartupMigrations(db *gorm.DB) error {
	if err := ensureSchemaMigrationTable(db); err != nil {
		return err
	}

	migrations := []startupMigration{
		{
			key:     schemaMigrationKey,
			version: autoMigrateSchemaVersion(),
			run:     AutoMigrate,
		},
		{
			key:     startupCleanupMigrationKey,
			version: startupCleanupSchemaVersion(),
			run:     CleanupSoftDeletedRows,
		},
		{
			key:     queryIndexMigrationKey,
			version: queryIndexSchemaVersion(),
			run:     EnsureQueryIndexes,
		},
	}

	storedVersions, err := readStoredMigrationVersions(db, migrations)
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		if err := runStartupMigrationIfNeeded(db, migration, storedVersions[migration.key]); err != nil {
			return err
		}
	}

	return nil
}

type startupMigration struct {
	key     string
	version string
	run     func(*gorm.DB) error
}

const (
	schemaMigrationKey         = "gorm_auto_migrate"
	startupCleanupMigrationKey = "startup_cleanup_soft_deleted"
	queryIndexMigrationKey     = "query_indexes"
)

var errStartupMigrationDeferred = errors.New("启动迁移部分依赖暂不可用")

func runStartupMigrationIfNeeded(db *gorm.DB, migration startupMigration, storedVersion string) error {
	if storedVersion == migration.version {
		return nil
	}

	logx.Infof("检测到数据库启动迁移 %s 版本变化，开始执行", migration.key)
	if err := migration.run(db); err != nil {
		if errors.Is(err, errStartupMigrationDeferred) {
			logx.Infof("数据库启动迁移 %s 未完全完成，将在下次启动重试: %v", migration.key, err)
			return nil
		}
		return err
	}
	return writeStoredMigrationVersion(db, migration.key, migration.version)
}

func readStoredMigrationVersions(db *gorm.DB, migrations []startupMigration) (map[string]string, error) {
	result := make(map[string]string, len(migrations))
	keys := make([]string, 0, len(migrations))
	for _, migration := range migrations {
		keys = append(keys, migration.key)
	}
	if len(keys) == 0 {
		return result, nil
	}

	var rows []struct {
		Key     string
		Version string
	}
	if err := db.Raw("SELECT key, version FROM schema_migrations WHERE key IN ?", keys).Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.Key] = row.Version
	}
	return result, nil
}

func writeStoredMigrationVersion(db *gorm.DB, key string, version string) error {
	return db.Exec(
		`INSERT INTO schema_migrations (key, version, migrated_at)
VALUES (?, ?, NOW())
ON CONFLICT (key) DO UPDATE SET version = EXCLUDED.version, migrated_at = EXCLUDED.migrated_at`,
		key,
		version,
	).Error
}

func autoMigrateModels() []interface{} {
	return []interface{}{
		&Movie{},
		&MovieLangSnapshot{},
		&TVSeries{},
		&TVLangSnapshot{},
		&Person{},
		&PersonLangSnapshot{},
		&AutoSyncExecutionLog{},
		&ProxyAccessLog{},
		&TmdbRequestLog{},
	}
}

func ensureSchemaMigrationTable(db *gorm.DB) error {
	return db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
	key varchar(64) PRIMARY KEY,
	version varchar(64) NOT NULL,
	migrated_at timestamptz NOT NULL
)`).Error
}

func autoMigrateSchemaVersion() string {
	parts := []string{"auto_migrate_v1"}
	for _, modelValue := range autoMigrateModels() {
		modelType := reflect.TypeOf(modelValue)
		if modelType.Kind() == reflect.Pointer {
			modelType = modelType.Elem()
		}
		parts = append(parts, modelType.PkgPath(), modelType.Name())
		for i := 0; i < modelType.NumField(); i++ {
			field := modelType.Field(i)
			parts = append(parts, field.Name, field.Type.String(), string(field.Tag))
		}
	}
	return hashStrings(parts...)
}

func hashStrings(parts ...string) string {
	h := sha1.New()
	for _, part := range parts {
		_, _ = h.Write([]byte(part))
		_, _ = h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}

// EnsureQueryIndexes 为常见列表查询补充索引
func EnsureQueryIndexes(db *gorm.DB) error {
	for _, stmt := range queryIndexSchemaStatements() {
		if err := db.Exec(stmt).Error; err != nil {
			return err
		}
	}

	for _, stmt := range queryIndexStatements() {
		if err := db.Exec(stmt).Error; err != nil {
			return err
		}
	}

	// 日志列表关键词为 col ILIKE '%x%' OR ...，需 pg_trgm GIN 索引才能避免 10 万级以上全表扫描与超时。
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS pg_trgm").Error; err != nil {
		logx.Infof("无法启用 pg_trgm（缺权限时可忽略），日志关键词查询可能较慢，并将在下次启动重试: %v", err)
		return errors.Join(errStartupMigrationDeferred, err)
	}

	for _, stmt := range queryIndexTrgmStatements() {
		if err := db.Exec(stmt).Error; err != nil {
			return err
		}
	}

	return nil
}

func queryIndexSchemaStatements() []string {
	return []string{
		"ALTER TABLE proxy_access_logs ADD COLUMN IF NOT EXISTS media_type varchar(16)",
		"ALTER TABLE proxy_access_logs ADD COLUMN IF NOT EXISTS tmdb_id bigint",
	}
}

func queryIndexStatements() []string {
	return []string{
		"CREATE INDEX IF NOT EXISTS idx_movies_live_popularity_desc ON movies (popularity DESC) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_tv_series_live_popularity_desc ON tv_series (popularity DESC) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_movies_live_created_tmdb_desc ON movies (created_at DESC, tmdb_id DESC) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_tv_series_live_created_tmdb_desc ON tv_series (created_at DESC, tmdb_id DESC) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_movies_live_id_asc ON movies (id ASC) WHERE deleted_at IS NULL AND tmdb_id > 0",
		"CREATE INDEX IF NOT EXISTS idx_tv_series_live_id_asc ON tv_series (id ASC) WHERE deleted_at IS NULL AND tmdb_id > 0",
		"CREATE INDEX IF NOT EXISTS idx_people_live_id_asc ON people (id ASC) WHERE deleted_at IS NULL AND tmdb_id > 0",
		"CREATE INDEX IF NOT EXISTS idx_proxy_access_logs_created_at_desc ON proxy_access_logs (created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_tmdb_request_logs_created_at_desc ON tmdb_request_logs (created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_proxy_access_logs_cleanup ON proxy_access_logs (created_at ASC, id ASC)",
		"CREATE INDEX IF NOT EXISTS idx_tmdb_request_logs_cleanup ON tmdb_request_logs (created_at ASC, id ASC)",
		"CREATE INDEX IF NOT EXISTS idx_proxy_access_logs_deleted_id_desc ON proxy_access_logs (deleted_at, id DESC)",
		"CREATE INDEX IF NOT EXISTS idx_tmdb_request_logs_deleted_id_desc ON tmdb_request_logs (deleted_at, id DESC)",
		"CREATE INDEX IF NOT EXISTS idx_proxy_access_logs_live_id_desc ON proxy_access_logs (id DESC) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_tmdb_request_logs_live_id_desc ON tmdb_request_logs (id DESC) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_proxy_access_logs_live_success_id_desc ON proxy_access_logs (id DESC) WHERE deleted_at IS NULL AND status_code >= 200 AND status_code < 400",
		"CREATE INDEX IF NOT EXISTS idx_tmdb_request_logs_live_success_id_desc ON tmdb_request_logs (id DESC) WHERE deleted_at IS NULL AND status_code >= 200 AND status_code < 400",
		"CREATE INDEX IF NOT EXISTS idx_proxy_access_logs_live_error_id_desc ON proxy_access_logs (id DESC) WHERE deleted_at IS NULL AND (status_code = 0 OR status_code < 200 OR status_code >= 400)",
		"CREATE INDEX IF NOT EXISTS idx_tmdb_request_logs_live_error_id_desc ON tmdb_request_logs (id DESC) WHERE deleted_at IS NULL AND (status_code = 0 OR status_code < 200 OR status_code >= 400)",
		"CREATE INDEX IF NOT EXISTS idx_proxy_access_logs_live_success_created_desc ON proxy_access_logs (created_at DESC) WHERE deleted_at IS NULL AND status_code >= 200 AND status_code < 400",
		"CREATE INDEX IF NOT EXISTS idx_proxy_access_logs_hot_media ON proxy_access_logs (created_at DESC, media_type, tmdb_id) WHERE deleted_at IS NULL AND status_code >= 200 AND status_code < 400 AND media_type IN ('movie', 'tv') AND tmdb_id <> 0",
		"CREATE INDEX IF NOT EXISTS idx_auto_sync_logs_live_id_desc ON auto_sync_execution_logs (id DESC) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_auto_sync_logs_live_status_id_desc ON auto_sync_execution_logs (status, id DESC) WHERE deleted_at IS NULL",
		"CREATE INDEX IF NOT EXISTS idx_auto_sync_logs_deleted_id_desc ON auto_sync_execution_logs (deleted_at, id DESC)",
		"CREATE INDEX IF NOT EXISTS idx_auto_sync_logs_deleted_status_id_desc ON auto_sync_execution_logs (deleted_at, status, id DESC)",
		"CREATE INDEX IF NOT EXISTS idx_auto_sync_logs_cleanup ON auto_sync_execution_logs (created_at ASC, id ASC)",
	}
}

func queryIndexTrgmStatements() []string {
	return []string{
		"CREATE INDEX IF NOT EXISTS idx_movies_title_trgm ON movies USING gin (title gin_trgm_ops)",
		"CREATE INDEX IF NOT EXISTS idx_movies_original_title_trgm ON movies USING gin (original_title gin_trgm_ops)",
		"CREATE INDEX IF NOT EXISTS idx_tv_series_name_trgm ON tv_series USING gin (name gin_trgm_ops)",
		"CREATE INDEX IF NOT EXISTS idx_tv_series_original_name_trgm ON tv_series USING gin (original_name gin_trgm_ops)",
		"CREATE INDEX IF NOT EXISTS idx_proxy_access_logs_method_trgm ON proxy_access_logs USING gin (method gin_trgm_ops)",
		"CREATE INDEX IF NOT EXISTS idx_proxy_access_logs_path_trgm ON proxy_access_logs USING gin (path gin_trgm_ops)",
		"CREATE INDEX IF NOT EXISTS idx_proxy_access_logs_query_trgm ON proxy_access_logs USING gin (query gin_trgm_ops)",
		"CREATE INDEX IF NOT EXISTS idx_proxy_access_logs_request_uri_trgm ON proxy_access_logs USING gin (request_uri gin_trgm_ops)",
		"CREATE INDEX IF NOT EXISTS idx_proxy_access_logs_client_ip_trgm ON proxy_access_logs USING gin (client_ip gin_trgm_ops)",
		"CREATE INDEX IF NOT EXISTS idx_proxy_access_logs_user_agent_trgm ON proxy_access_logs USING gin (user_agent gin_trgm_ops)",
		"CREATE INDEX IF NOT EXISTS idx_proxy_access_logs_err_trgm ON proxy_access_logs USING gin (error_message gin_trgm_ops)",
		"CREATE INDEX IF NOT EXISTS idx_tmdb_request_logs_method_trgm ON tmdb_request_logs USING gin (method gin_trgm_ops)",
		"CREATE INDEX IF NOT EXISTS idx_tmdb_request_logs_path_trgm ON tmdb_request_logs USING gin (path gin_trgm_ops)",
		"CREATE INDEX IF NOT EXISTS idx_tmdb_request_logs_url_trgm ON tmdb_request_logs USING gin (url gin_trgm_ops)",
		"CREATE INDEX IF NOT EXISTS idx_tmdb_request_logs_err_trgm ON tmdb_request_logs USING gin (error_message gin_trgm_ops)",
	}
}

func queryIndexSchemaVersion() string {
	parts := []string{"query_indexes_v1", "CREATE EXTENSION IF NOT EXISTS pg_trgm"}
	parts = append(parts, queryIndexSchemaStatements()...)
	parts = append(parts, queryIndexStatements()...)
	parts = append(parts, queryIndexTrgmStatements()...)
	return hashStrings(parts...)
}
