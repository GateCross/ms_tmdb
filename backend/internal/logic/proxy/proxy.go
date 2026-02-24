package proxy

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"ms_tmdb/internal/model"
	"ms_tmdb/pkg/tmdbclient"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

const tvSeasonLocalDataKey = "_ms_tv_season_local"

// ProxyService 代理服务，封装 Read-Through 缓存逻辑
type ProxyService struct {
	DB         *gorm.DB
	TmdbClient *tmdbclient.Client
}

func NewProxyService(db *gorm.DB, client *tmdbclient.Client) *ProxyService {
	return &ProxyService{DB: db, TmdbClient: client}
}

// GetMovieDetail Read-Through 获取电影详情
func (s *ProxyService) GetMovieDetail(tmdbID int, opts *tmdbclient.RequestOption) (json.RawMessage, error) {
	var movie model.Movie
	err := s.DB.Where("tmdb_id = ?", tmdbID).First(&movie).Error

	if err == nil {
		if movie.IsModified || !isExpired(movie.LastSyncedAt, 24*time.Hour) {
			return json.RawMessage(movie.TmdbData), nil
		}
	}

	data, fetchErr := s.TmdbClient.GetMovie(tmdbID, opts)
	if fetchErr != nil {
		if err == nil {
			logx.Infof("TMDB 不可用，返回本地缓存: movie/%d", tmdbID)
			return json.RawMessage(movie.TmdbData), nil
		}
		return nil, fetchErr
	}

	s.upsertMovie(tmdbID, data)
	return data, nil
}

// GetTvSeriesDetail Read-Through 获取电视剧详情
func (s *ProxyService) GetTvSeriesDetail(tmdbID int, opts *tmdbclient.RequestOption) (json.RawMessage, error) {
	var tv model.TVSeries
	err := s.DB.Where("tmdb_id = ?", tmdbID).First(&tv).Error

	if err == nil {
		if tv.IsModified || !isExpired(tv.LastSyncedAt, 24*time.Hour) {
			return json.RawMessage(tv.TmdbData), nil
		}
	}

	data, fetchErr := s.TmdbClient.GetTVSeries(tmdbID, opts)
	if fetchErr != nil {
		if err == nil {
			return json.RawMessage(tv.TmdbData), nil
		}
		return nil, fetchErr
	}

	s.upsertTVSeries(tmdbID, data)
	return data, nil
}

// GetTvSeasonDetail 优先返回本地保存的季明细，未保存时透传 TMDB
func (s *ProxyService) GetTvSeasonDetail(seriesID, seasonNumber int, opts *tmdbclient.RequestOption) (json.RawMessage, error) {
	localData, hasLocal, localErr := s.GetLocalTvSeason(seriesID, seasonNumber)
	if localErr != nil {
		return nil, localErr
	}
	if hasLocal {
		raw, err := json.Marshal(localData)
		if err != nil {
			return nil, err
		}
		return raw, nil
	}

	return s.TmdbClient.GetTVSeason(seriesID, seasonNumber, opts)
}

// GetLocalTvSeason 获取本地已保存季明细
func (s *ProxyService) GetLocalTvSeason(seriesID, seasonNumber int) (map[string]interface{}, bool, error) {
	var tv model.TVSeries
	if err := s.DB.Where("tmdb_id = ?", seriesID).First(&tv).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, err
	}

	localPatch, err := rawJSONToMap(tv.LocalData)
	if err != nil {
		return nil, false, err
	}

	seasons, ok := localPatch[tvSeasonLocalDataKey].(map[string]interface{})
	if !ok {
		return nil, false, nil
	}

	seasonData, ok := seasons[strconv.Itoa(seasonNumber)].(map[string]interface{})
	if !ok {
		return nil, false, nil
	}

	normalized, err := normalizeSeasonDetailPayload(seasonData, seasonNumber)
	if err != nil {
		return nil, false, err
	}
	return normalized, true, nil
}

// SaveTvSeasonToLocal 从 TMDB 拉取季明细并写入本地（重复调用即覆盖）
func (s *ProxyService) SaveTvSeasonToLocal(seriesID, seasonNumber int, opts *tmdbclient.RequestOption) (map[string]interface{}, error) {
	if err := s.ensureTVSeriesExists(seriesID, opts); err != nil {
		return nil, err
	}

	raw, err := s.TmdbClient.GetTVSeason(seriesID, seasonNumber, opts)
	if err != nil {
		return nil, err
	}
	payload, err := unmarshalRawToMap(raw)
	if err != nil {
		return nil, err
	}

	normalized, err := normalizeSeasonDetailPayload(payload, seasonNumber)
	if err != nil {
		return nil, err
	}
	if err := s.saveTvSeasonPayload(seriesID, seasonNumber, normalized); err != nil {
		return nil, err
	}
	return normalized, nil
}

// UpdateLocalTvSeason 更新本地季明细（仅修改本地覆盖数据）
func (s *ProxyService) UpdateLocalTvSeason(seriesID, seasonNumber int, payload map[string]interface{}) (map[string]interface{}, error) {
	if err := s.ensureTVSeriesExists(seriesID, nil); err != nil {
		return nil, err
	}

	normalized, err := normalizeSeasonDetailPayload(payload, seasonNumber)
	if err != nil {
		return nil, err
	}
	if err := s.saveTvSeasonPayload(seriesID, seasonNumber, normalized); err != nil {
		return nil, err
	}
	return normalized, nil
}

// GetPersonDetail Read-Through 获取人物详情
func (s *ProxyService) GetPersonDetail(tmdbID int, opts *tmdbclient.RequestOption) (json.RawMessage, error) {
	var person model.Person
	err := s.DB.Where("tmdb_id = ?", tmdbID).First(&person).Error

	if err == nil {
		if person.IsModified || !isExpired(person.LastSyncedAt, 48*time.Hour) {
			return json.RawMessage(person.TmdbData), nil
		}
	}

	data, fetchErr := s.TmdbClient.GetPerson(tmdbID, opts)
	if fetchErr != nil {
		if err == nil {
			return json.RawMessage(person.TmdbData), nil
		}
		return nil, fetchErr
	}

	s.upsertPerson(tmdbID, data)
	return data, nil
}

func (s *ProxyService) upsertMovie(tmdbID int, data json.RawMessage) {
	var parsed struct {
		Title            string  `json:"title"`
		OriginalTitle    string  `json:"original_title"`
		Overview         string  `json:"overview"`
		ReleaseDate      string  `json:"release_date"`
		Popularity       float64 `json:"popularity"`
		VoteAverage      float64 `json:"vote_average"`
		VoteCount        int     `json:"vote_count"`
		PosterPath       string  `json:"poster_path"`
		BackdropPath     string  `json:"backdrop_path"`
		OriginalLanguage string  `json:"original_language"`
		Adult            bool    `json:"adult"`
		Status           string  `json:"status"`
		Runtime          int     `json:"runtime"`
		Budget           int64   `json:"budget"`
		Revenue          int64   `json:"revenue"`
		Tagline          string  `json:"tagline"`
		Homepage         string  `json:"homepage"`
		ImdbID           string  `json:"imdb_id"`
	}
	json.Unmarshal(data, &parsed)

	now := time.Now()
	result := s.DB.Where("tmdb_id = ?", tmdbID).First(&model.Movie{})
	if result.Error == gorm.ErrRecordNotFound {
		s.DB.Create(&model.Movie{
			TmdbID: tmdbID, Title: parsed.Title, OriginalTitle: parsed.OriginalTitle,
			Overview: parsed.Overview, ReleaseDate: parsed.ReleaseDate,
			Popularity: parsed.Popularity, VoteAverage: parsed.VoteAverage, VoteCount: parsed.VoteCount,
			PosterPath: parsed.PosterPath, BackdropPath: parsed.BackdropPath,
			OriginalLanguage: parsed.OriginalLanguage, Adult: parsed.Adult, Status: parsed.Status,
			Runtime: parsed.Runtime, Budget: parsed.Budget, Revenue: parsed.Revenue,
			Tagline: parsed.Tagline, Homepage: parsed.Homepage, ImdbID: parsed.ImdbID,
			TmdbData: model.RawJSON(data), LastSyncedAt: &now,
		})
	} else {
		s.DB.Model(&model.Movie{}).Where("tmdb_id = ?", tmdbID).Updates(map[string]interface{}{
			"title": parsed.Title, "original_title": parsed.OriginalTitle,
			"overview": parsed.Overview, "popularity": parsed.Popularity,
			"vote_average": parsed.VoteAverage, "poster_path": parsed.PosterPath,
			"tmdb_data": model.RawJSON(data), "last_synced_at": &now,
		})
	}
}

func (s *ProxyService) upsertTVSeries(tmdbID int, data json.RawMessage) {
	var parsed struct {
		Name         string  `json:"name"`
		OriginalName string  `json:"original_name"`
		Overview     string  `json:"overview"`
		FirstAirDate string  `json:"first_air_date"`
		Popularity   float64 `json:"popularity"`
		VoteAverage  float64 `json:"vote_average"`
		PosterPath   string  `json:"poster_path"`
		Status       string  `json:"status"`
	}
	json.Unmarshal(data, &parsed)

	now := time.Now()
	result := s.DB.Where("tmdb_id = ?", tmdbID).First(&model.TVSeries{})
	if result.Error == gorm.ErrRecordNotFound {
		s.DB.Create(&model.TVSeries{
			TmdbID: tmdbID, Name: parsed.Name, OriginalName: parsed.OriginalName,
			Overview: parsed.Overview, FirstAirDate: parsed.FirstAirDate,
			Popularity: parsed.Popularity, VoteAverage: parsed.VoteAverage,
			PosterPath: parsed.PosterPath, Status: parsed.Status,
			TmdbData: model.RawJSON(data), LastSyncedAt: &now,
		})
	} else {
		s.DB.Model(&model.TVSeries{}).Where("tmdb_id = ?", tmdbID).Updates(map[string]interface{}{
			"name": parsed.Name, "overview": parsed.Overview,
			"popularity": parsed.Popularity, "vote_average": parsed.VoteAverage,
			"poster_path": parsed.PosterPath, "tmdb_data": model.RawJSON(data), "last_synced_at": &now,
		})
	}
}

func (s *ProxyService) upsertPerson(tmdbID int, data json.RawMessage) {
	var parsed struct {
		Name        string  `json:"name"`
		Biography   string  `json:"biography"`
		Popularity  float64 `json:"popularity"`
		ProfilePath string  `json:"profile_path"`
	}
	json.Unmarshal(data, &parsed)

	now := time.Now()
	result := s.DB.Where("tmdb_id = ?", tmdbID).First(&model.Person{})
	if result.Error == gorm.ErrRecordNotFound {
		s.DB.Create(&model.Person{
			TmdbID: tmdbID, Name: parsed.Name, Biography: parsed.Biography,
			Popularity: parsed.Popularity, ProfilePath: parsed.ProfilePath,
			TmdbData: model.RawJSON(data), LastSyncedAt: &now,
		})
	} else {
		s.DB.Model(&model.Person{}).Where("tmdb_id = ?", tmdbID).Updates(map[string]interface{}{
			"name": parsed.Name, "popularity": parsed.Popularity,
			"profile_path": parsed.ProfilePath, "tmdb_data": model.RawJSON(data), "last_synced_at": &now,
		})
	}
}

func isExpired(syncedAt *time.Time, ttl time.Duration) bool {
	if syncedAt == nil {
		return true
	}
	return time.Since(*syncedAt) > ttl
}

func (s *ProxyService) ensureTVSeriesExists(seriesID int, opts *tmdbclient.RequestOption) error {
	var tv model.TVSeries
	if err := s.DB.Where("tmdb_id = ?", seriesID).First(&tv).Error; err == nil {
		return nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	_, err := s.GetTvSeriesDetail(seriesID, opts)
	return err
}

func (s *ProxyService) saveTvSeasonPayload(seriesID, seasonNumber int, payload map[string]interface{}) error {
	var tv model.TVSeries
	if err := s.DB.Where("tmdb_id = ?", seriesID).First(&tv).Error; err != nil {
		return err
	}

	localPatch, err := rawJSONToMap(tv.LocalData)
	if err != nil {
		return err
	}
	seasons, _ := localPatch[tvSeasonLocalDataKey].(map[string]interface{})
	if seasons == nil {
		seasons = map[string]interface{}{}
	}
	seasons[strconv.Itoa(seasonNumber)] = payload
	localPatch[tvSeasonLocalDataKey] = seasons

	rawPatch, err := marshalMapToRawJSON(localPatch)
	if err != nil {
		return err
	}

	return s.DB.Model(&model.TVSeries{}).Where("tmdb_id = ?", seriesID).Updates(map[string]interface{}{
		"local_data":  rawPatch,
		"is_modified": true,
	}).Error
}

func rawJSONToMap(raw model.RawJSON) (map[string]interface{}, error) {
	result := map[string]interface{}{}
	if len(raw) == 0 {
		return result, nil
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func marshalMapToRawJSON(payload map[string]interface{}) (model.RawJSON, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return model.RawJSON(raw), nil
}

func unmarshalRawToMap(raw json.RawMessage) (map[string]interface{}, error) {
	result := map[string]interface{}{}
	if len(raw) == 0 {
		return result, nil
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func normalizeSeasonDetailPayload(input map[string]interface{}, seasonNumber int) (map[string]interface{}, error) {
	raw, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	result["season_number"] = seasonNumber
	if _, ok := result["episodes"].([]interface{}); !ok {
		result["episodes"] = []interface{}{}
	}
	return result, nil
}
