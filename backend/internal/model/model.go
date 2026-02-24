package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

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

// TVSeries 电视剧模型
type TVSeries struct {
	gorm.Model
	TmdbID           int        `gorm:"uniqueIndex;not null" json:"tmdb_id"`
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
	LastSyncedAt     *time.Time `json:"last_synced_at"`
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

	Status  string `gorm:"size:32;index" json:"status"`
	Checked int    `json:"checked"`
	Synced  int    `json:"synced"`
	Failed  int    `json:"failed"`
	Message string `gorm:"type:text" json:"message"`
}

// AutoMigrate 自动建表迁移
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&Movie{},
		&TVSeries{},
		&Person{},
		&AutoSyncExecutionLog{},
	)
}
