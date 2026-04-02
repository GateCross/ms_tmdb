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

// ResolveMovieSyncID 将对外 TMDB ID 解析为实际拉取 TMDB 的 ID。
func (s *ProxyService) ResolveMovieSyncID(tmdbID int) int {
	var movie model.Movie
	if err := s.DB.Where("tmdb_id = ?", tmdbID).First(&movie).Error; err != nil {
		return tmdbID
	}
	resolved := resolveSyncTmdbID(movie.SyncTmdbID, movie.TmdbID)
	if resolved <= 0 {
		return tmdbID
	}
	return resolved
}

// ResolveTVSyncID 将对外 TMDB ID 解析为实际拉取 TMDB 的 ID。
func (s *ProxyService) ResolveTVSyncID(tmdbID int) int {
	var tv model.TVSeries
	if err := s.DB.Where("tmdb_id = ?", tmdbID).First(&tv).Error; err != nil {
		return tmdbID
	}
	resolved := resolveSyncTmdbID(tv.SyncTmdbID, tv.TmdbID)
	if resolved <= 0 {
		return tmdbID
	}
	return resolved
}

// GetMovieDetail Read-Through 获取电影详情
func (s *ProxyService) GetMovieDetail(tmdbID int, opts *tmdbclient.RequestOption) (json.RawMessage, error) {
	var movie model.Movie
	err := s.DB.Where("tmdb_id = ?", tmdbID).First(&movie).Error
	syncTmdbID := tmdbID

	// 1. 判断是否带有 append_to_response 等复杂参数
	bypassCache := shouldBypassCache(opts)

	if err == nil {
		syncTmdbID = resolveSyncTmdbID(movie.SyncTmdbID, movie.TmdbID)
		if syncTmdbID == 0 {
			syncTmdbID = tmdbID
		}
		// 2. 只有在没有复杂请求参数时，才允许返回本地基础缓存
		if !bypassCache {
			if movie.IsModified || !isExpired(movie.LastSyncedAt, 24*time.Hour) {
				return rewriteTMDBID(json.RawMessage(movie.TmdbData), tmdbID, syncTmdbID)
			}
		}
	}

	data, fetchErr := s.TmdbClient.GetMovie(syncTmdbID, opts)
	if fetchErr != nil {
		// 3. 降级处理：即使有复杂参数，如果 TMDB 挂了，仍尽力返回本地缓存
		if err == nil {
			logx.Infof("TMDB 不可用，返回本地缓存: movie/%d", tmdbID)
			return rewriteTMDBID(json.RawMessage(movie.TmdbData), tmdbID, syncTmdbID)
		}
		return nil, fetchErr
	}

	normalizedData, normalizeErr := rewriteTMDBID(data, tmdbID, syncTmdbID)
	if normalizeErr != nil {
		return nil, normalizeErr
	}

	// 4. 不要把带 append_to_response 的大 JSON 写入到基础缓存库中
	if !bypassCache {
		if err := s.upsertMovie(tmdbID, syncTmdbID, normalizedData); err != nil {
			return nil, err
		}
	}
	
	return normalizedData, nil
}

	data, fetchErr := s.TmdbClient.GetMovie(syncTmdbID, opts)
	if fetchErr != nil {
		if err == nil {
			logx.Infof("TMDB 不可用，返回本地缓存: movie/%d", tmdbID)
			return rewriteTMDBID(json.RawMessage(movie.TmdbData), tmdbID, syncTmdbID)
		}
		return nil, fetchErr
	}

	normalizedData, normalizeErr := rewriteTMDBID(data, tmdbID, syncTmdbID)
	if normalizeErr != nil {
		return nil, normalizeErr
	}

	if err := s.upsertMovie(tmdbID, syncTmdbID, normalizedData); err != nil {
		return nil, err
	}
	return normalizedData, nil
}

// GetTvSeriesDetail Read-Through 获取电视剧详情
func (s *ProxyService) GetTvSeriesDetail(tmdbID int, opts *tmdbclient.RequestOption) (json.RawMessage, error) {
	var tv model.TVSeries
	err := s.DB.Where("tmdb_id = ?", tmdbID).First(&tv).Error
	syncTmdbID := tmdbID

	bypassCache := shouldBypassCache(opts)

	if err == nil {
		syncTmdbID = resolveSyncTmdbID(tv.SyncTmdbID, tv.TmdbID)
		if syncTmdbID == 0 {
			syncTmdbID = tmdbID
		}
		if !bypassCache {
			if tv.IsModified || !isExpired(tv.LastSyncedAt, 24*time.Hour) {
				return rewriteTMDBID(json.RawMessage(tv.TmdbData), tmdbID, syncTmdbID)
			}
		}
	}

	data, fetchErr := s.TmdbClient.GetTVSeries(syncTmdbID, opts)
	if fetchErr != nil {
		if err == nil {
			logx.Infof("TMDB 不可用，返回本地缓存: tv/%d", tmdbID)
			return rewriteTMDBID(json.RawMessage(tv.TmdbData), tmdbID, syncTmdbID)
		}
		return nil, fetchErr
	}

	normalizedData, normalizeErr := rewriteTMDBID(data, tmdbID, syncTmdbID)
	if normalizeErr != nil {
		return nil, normalizeErr
	}

	if !bypassCache {
		if err := s.upsertTVSeries(tmdbID, syncTmdbID, normalizedData); err != nil {
			return nil, err
		}
	}
	
	return normalizedData, nil
}

	data, fetchErr := s.TmdbClient.GetTVSeries(syncTmdbID, opts)
	if fetchErr != nil {
		if err == nil {
			return rewriteTMDBID(json.RawMessage(tv.TmdbData), tmdbID, syncTmdbID)
		}
		return nil, fetchErr
	}

	normalizedData, normalizeErr := rewriteTMDBID(data, tmdbID, syncTmdbID)
	if normalizeErr != nil {
		return nil, normalizeErr
	}

	if err := s.upsertTVSeries(tmdbID, syncTmdbID, normalizedData); err != nil {
		return nil, err
	}
	return normalizedData, nil
}

// GetTvSeasonDetail 优先返回本地保存的季明细，未保存时透传 TMDB，带本地修改时进行数据合并
func (s *ProxyService) GetTvSeasonDetail(seriesID, seasonNumber int, opts *tmdbclient.RequestOption) (json.RawMessage, error) {
	syncSeriesID := seriesID
	var tv model.TVSeries
	if err := s.DB.Where("tmdb_id = ?", seriesID).First(&tv).Error; err == nil {
		syncSeriesID = resolveSyncTmdbID(tv.SyncTmdbID, tv.TmdbID)
	}

	// 先尝试获取本地覆盖的数据
	localData, hasLocal, localErr := s.GetLocalTvSeason(seriesID, seasonNumber)
	if localErr != nil {
		return nil, localErr
	}

	bypassCache := shouldBypassCache(opts)

	// 情况1：本地有数据，且没有请求 credits/videos 等附加信息，直接返回本地
	if hasLocal && !bypassCache {
		raw, err := json.Marshal(localData)
		if err != nil {
			return nil, err
		}
		return raw, nil
	}

	// 情况2：去 TMDB 拉取最新（或带 append）的数据
	tmdbData, err := s.TmdbClient.GetTVSeason(syncSeriesID, seasonNumber, opts)
	if err != nil {
		// 降级：拉取失败但本地有数据，返回本地
		if hasLocal {
			raw, _ := json.Marshal(localData)
			return raw, nil
		}
		return nil, err
	}

	// 情况3：既有 TMDB 远端返回的最新数据，本地又有修改，执行数据合并 (Merge)
	if hasLocal {
		tmdbMap, err := unmarshalRawToMap(tmdbData)
		if err == nil {
			// 将本地自定义的字段(如集名称、简介)覆盖到 TMDB 官方结果上
			for k, v := range localData {
				tmdbMap[k] = v
			}
			mergedRaw, err := json.Marshal(tmdbMap)
			if err == nil {
				return mergedRaw, nil
			}
		} else {
			logx.Errorf("合并季明细失败，反序列化 TMDB 数据错误: %v", err)
		}
	}

	// 如果没有本地数据，或者合并失败，直接返回原汁原味的 TMDB 数据
	return tmdbData, nil
}

	syncSeriesID := seriesID
	var tv model.TVSeries
	if err := s.DB.Where("tmdb_id = ?", seriesID).First(&tv).Error; err == nil {
		syncSeriesID = resolveSyncTmdbID(tv.SyncTmdbID, tv.TmdbID)
	}

	return s.TmdbClient.GetTVSeason(syncSeriesID, seasonNumber, opts)
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

	syncSeriesID := seriesID
	var tv model.TVSeries
	if err := s.DB.Where("tmdb_id = ?", seriesID).First(&tv).Error; err == nil {
		syncSeriesID = resolveSyncTmdbID(tv.SyncTmdbID, tv.TmdbID)
	}

	raw, err := s.TmdbClient.GetTVSeason(syncSeriesID, seasonNumber, opts)
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

	if err := s.upsertPerson(tmdbID, data); err != nil {
		return nil, err
	}
	return data, nil
}

func (s *ProxyService) upsertMovie(tmdbID int, syncTmdbID int, data json.RawMessage) error {
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
	if err := json.Unmarshal(data, &parsed); err != nil {
		logx.Errorf("解析电影 TMDB 数据失败: tmdb_id=%d err=%v", tmdbID, err)
		return err
	}

	now := time.Now()
	result := s.DB.Where("tmdb_id = ?", tmdbID).First(&model.Movie{})
	if result.Error == gorm.ErrRecordNotFound {
		if err := s.DB.Create(&model.Movie{
			TmdbID: tmdbID, SyncTmdbID: resolveSyncTmdbID(syncTmdbID, tmdbID), Title: parsed.Title, OriginalTitle: parsed.OriginalTitle,
			Overview: parsed.Overview, ReleaseDate: parsed.ReleaseDate,
			Popularity: parsed.Popularity, VoteAverage: parsed.VoteAverage, VoteCount: parsed.VoteCount,
			PosterPath: parsed.PosterPath, BackdropPath: parsed.BackdropPath,
			OriginalLanguage: parsed.OriginalLanguage, Adult: parsed.Adult, Status: parsed.Status,
			Runtime: parsed.Runtime, Budget: parsed.Budget, Revenue: parsed.Revenue,
			Tagline: parsed.Tagline, Homepage: parsed.Homepage, ImdbID: parsed.ImdbID,
			TmdbData: model.RawJSON(data), LastSyncedAt: &now,
		}).Error; err != nil {
			return err
		}
		return nil
	}
	if result.Error != nil {
		return result.Error
	}

	if err := s.DB.Model(&model.Movie{}).Where("tmdb_id = ?", tmdbID).Updates(map[string]interface{}{
		"title": parsed.Title, "original_title": parsed.OriginalTitle,
		"overview": parsed.Overview, "popularity": parsed.Popularity,
		"vote_average": parsed.VoteAverage, "poster_path": parsed.PosterPath,
		"tmdb_data": model.RawJSON(data), "last_synced_at": &now, "sync_tmdb_id": resolveSyncTmdbID(syncTmdbID, tmdbID),
	}).Error; err != nil {
		return err
	}

	return nil
}

func (s *ProxyService) upsertTVSeries(tmdbID int, syncTmdbID int, data json.RawMessage) error {
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
	if err := json.Unmarshal(data, &parsed); err != nil {
		logx.Errorf("解析剧集 TMDB 数据失败: tmdb_id=%d err=%v", tmdbID, err)
		return err
	}

	now := time.Now()
	result := s.DB.Where("tmdb_id = ?", tmdbID).First(&model.TVSeries{})
	if result.Error == gorm.ErrRecordNotFound {
		if err := s.DB.Create(&model.TVSeries{
			TmdbID: tmdbID, SyncTmdbID: resolveSyncTmdbID(syncTmdbID, tmdbID), Name: parsed.Name, OriginalName: parsed.OriginalName,
			Overview: parsed.Overview, FirstAirDate: parsed.FirstAirDate,
			Popularity: parsed.Popularity, VoteAverage: parsed.VoteAverage,
			PosterPath: parsed.PosterPath, Status: parsed.Status,
			TmdbData: model.RawJSON(data), LastSyncedAt: &now,
		}).Error; err != nil {
			return err
		}
		return nil
	}
	if result.Error != nil {
		return result.Error
	}

	if err := s.DB.Model(&model.TVSeries{}).Where("tmdb_id = ?", tmdbID).Updates(map[string]interface{}{
		"name": parsed.Name, "overview": parsed.Overview,
		"popularity": parsed.Popularity, "vote_average": parsed.VoteAverage,
		"poster_path": parsed.PosterPath, "tmdb_data": model.RawJSON(data), "last_synced_at": &now, "sync_tmdb_id": resolveSyncTmdbID(syncTmdbID, tmdbID),
	}).Error; err != nil {
		return err
	}

	return nil
}

func (s *ProxyService) upsertPerson(tmdbID int, data json.RawMessage) error {
	var parsed struct {
		Name        string  `json:"name"`
		Biography   string  `json:"biography"`
		Popularity  float64 `json:"popularity"`
		ProfilePath string  `json:"profile_path"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		logx.Errorf("解析人物 TMDB 数据失败: tmdb_id=%d err=%v", tmdbID, err)
		return err
	}

	now := time.Now()
	result := s.DB.Where("tmdb_id = ?", tmdbID).First(&model.Person{})
	if result.Error == gorm.ErrRecordNotFound {
		if err := s.DB.Create(&model.Person{
			TmdbID: tmdbID, Name: parsed.Name, Biography: parsed.Biography,
			Popularity: parsed.Popularity, ProfilePath: parsed.ProfilePath,
			TmdbData: model.RawJSON(data), LastSyncedAt: &now,
		}).Error; err != nil {
			return err
		}
		return nil
	}
	if result.Error != nil {
		return result.Error
	}

	if err := s.DB.Model(&model.Person{}).Where("tmdb_id = ?", tmdbID).Updates(map[string]interface{}{
		"name": parsed.Name, "popularity": parsed.Popularity,
		"profile_path": parsed.ProfilePath, "tmdb_data": model.RawJSON(data), "last_synced_at": &now,
	}).Error; err != nil {
		return err
	}

	return nil
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

func resolveSyncTmdbID(syncTmdbID int, currentTmdbID int) int {
	if syncTmdbID > 0 {
		return syncTmdbID
	}
	if currentTmdbID > 0 {
		return currentTmdbID
	}
	return 0
}

func rewriteTMDBID(raw json.RawMessage, tmdbID int, syncTmdbID int) (json.RawMessage, error) {
	payload := map[string]interface{}{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}
	if tmdbID != 0 {
		payload["id"] = tmdbID
	}
	if syncTmdbID == 0 {
		syncTmdbID = tmdbID
	}
	payload["sync_tmdb_id"] = syncTmdbID
	normalized, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return normalized, nil
}
// --- 新增辅助函数 ---

// shouldBypassCache 判断是否需要跳过本地基础缓存
// 如果请求要求附加响应内容(如 credits, videos)，本地数据库通常没有这些完整字段，必须强制透传 TMDB
func shouldBypassCache(opts *tmdbclient.RequestOption) bool {
	if opts == nil {
		return false
	}
	if opts.AppendToResponse != "" {
		return true
	}
	return false
}