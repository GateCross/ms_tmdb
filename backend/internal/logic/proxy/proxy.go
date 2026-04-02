package proxy

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"ms_tmdb/internal/model"
	"ms_tmdb/pkg/tmdbclient"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

const tvSeasonLocalDataKey = "_ms_tv_season_local"

// ProxyService 代理服务，封装 Read-Through 缓存逻辑
type ProxyService struct {
	DB          *gorm.DB
	TmdbClient  *tmdbclient.Client
	DefaultLang string
}

func NewProxyService(db *gorm.DB, client *tmdbclient.Client, defaultLang string) *ProxyService {
	return &ProxyService{
		DB:          db,
		TmdbClient:  client,
		DefaultLang: strings.TrimSpace(defaultLang),
	}
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
	language := s.requestLanguage(opts)
	if s.isNonDefaultLanguage(language) {
		return s.getMovieDetailByLanguage(tmdbID, opts, language)
	}

	var movie model.Movie
	err := s.DB.Where("tmdb_id = ?", tmdbID).First(&movie).Error
	syncTmdbID := tmdbID

	if err == nil {
		syncTmdbID = resolveSyncTmdbID(movie.SyncTmdbID, movie.TmdbID)
		if syncTmdbID == 0 {
			syncTmdbID = tmdbID
		}
		if movie.IsModified || !isExpired(movie.LastSyncedAt, 24*time.Hour) {
			return rewriteTMDBID(json.RawMessage(movie.TmdbData), tmdbID, syncTmdbID)
		}
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
	language := s.requestLanguage(opts)
	if s.isNonDefaultLanguage(language) {
		return s.getTVSeriesDetailByLanguage(tmdbID, opts, language)
	}

	var tv model.TVSeries
	err := s.DB.Where("tmdb_id = ?", tmdbID).First(&tv).Error
	syncTmdbID := tmdbID

	if err == nil {
		syncTmdbID = resolveSyncTmdbID(tv.SyncTmdbID, tv.TmdbID)
		if syncTmdbID == 0 {
			syncTmdbID = tmdbID
		}
		if tv.IsModified || !isExpired(tv.LastSyncedAt, 24*time.Hour) {
			return rewriteTMDBID(json.RawMessage(tv.TmdbData), tmdbID, syncTmdbID)
		}
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
	language := s.requestLanguage(opts)
	if s.isNonDefaultLanguage(language) {
		return s.getPersonDetailByLanguage(tmdbID, opts, language)
	}

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

func (s *ProxyService) requestLanguage(opts *tmdbclient.RequestOption) string {
	if opts != nil {
		if language := strings.TrimSpace(opts.Language); language != "" {
			return language
		}
	}
	return strings.TrimSpace(s.DefaultLang)
}

func (s *ProxyService) isNonDefaultLanguage(language string) bool {
	normalized := normalizeLanguageTag(language)
	if normalized == "" {
		return false
	}
	defaultLanguage := normalizeLanguageTag(s.DefaultLang)
	if defaultLanguage == "" {
		return false
	}
	return normalized != defaultLanguage
}

func normalizeLanguageTag(language string) string {
	return strings.ToLower(strings.TrimSpace(language))
}

func (s *ProxyService) getMovieDetailByLanguage(tmdbID int, opts *tmdbclient.RequestOption, language string) (json.RawMessage, error) {
	normalizedLanguage := normalizeLanguageTag(language)

	var movie model.Movie
	movieErr := s.DB.Where("tmdb_id = ?", tmdbID).First(&movie).Error
	syncTmdbID := tmdbID
	var localData model.RawJSON

	if movieErr == nil {
		syncTmdbID = resolveSyncTmdbID(movie.SyncTmdbID, movie.TmdbID)
		if syncTmdbID == 0 {
			syncTmdbID = tmdbID
		}
		localData = movie.LocalData
	} else if movieErr != nil && !errors.Is(movieErr, gorm.ErrRecordNotFound) {
		return nil, movieErr
	}

	var snapshot model.MovieLangSnapshot
	snapshotErr := s.DB.Where("tmdb_id = ? AND language = ?", tmdbID, normalizedLanguage).First(&snapshot).Error
	if snapshotErr == nil {
		cachedData, err := rewriteTMDBID(json.RawMessage(snapshot.TmdbData), tmdbID, resolveSyncTmdbID(snapshot.SyncTmdbID, syncTmdbID))
		if err != nil {
			return nil, err
		}
		if !isExpired(snapshot.LastSyncedAt, 24*time.Hour) {
			return mergeTMDBWithLocalData(cachedData, localData)
		}
	} else if !errors.Is(snapshotErr, gorm.ErrRecordNotFound) {
		return nil, snapshotErr
	}

	data, fetchErr := s.TmdbClient.GetMovie(syncTmdbID, opts)
	if fetchErr != nil {
		if snapshotErr == nil {
			cachedData, err := rewriteTMDBID(json.RawMessage(snapshot.TmdbData), tmdbID, resolveSyncTmdbID(snapshot.SyncTmdbID, syncTmdbID))
			if err != nil {
				return nil, err
			}
			return mergeTMDBWithLocalData(cachedData, localData)
		}
		if movieErr == nil {
			logx.Infof("TMDB 不可用，返回本地默认语言缓存: movie/%d language=%s", tmdbID, normalizedLanguage)
			return rewriteTMDBID(json.RawMessage(movie.TmdbData), tmdbID, syncTmdbID)
		}
		return nil, fetchErr
	}

	normalizedData, err := rewriteTMDBID(data, tmdbID, syncTmdbID)
	if err != nil {
		return nil, err
	}
	if err := s.upsertMovieLangSnapshot(tmdbID, normalizedLanguage, syncTmdbID, normalizedData); err != nil {
		return nil, err
	}
	if errors.Is(movieErr, gorm.ErrRecordNotFound) {
		s.warmDefaultMovieRecord(tmdbID, opts)
	}
	return mergeTMDBWithLocalData(normalizedData, localData)
}

func (s *ProxyService) getTVSeriesDetailByLanguage(tmdbID int, opts *tmdbclient.RequestOption, language string) (json.RawMessage, error) {
	normalizedLanguage := normalizeLanguageTag(language)

	var tv model.TVSeries
	tvErr := s.DB.Where("tmdb_id = ?", tmdbID).First(&tv).Error
	syncTmdbID := tmdbID
	var localData model.RawJSON

	if tvErr == nil {
		syncTmdbID = resolveSyncTmdbID(tv.SyncTmdbID, tv.TmdbID)
		if syncTmdbID == 0 {
			syncTmdbID = tmdbID
		}
		localData = tv.LocalData
	} else if tvErr != nil && !errors.Is(tvErr, gorm.ErrRecordNotFound) {
		return nil, tvErr
	}

	var snapshot model.TVLangSnapshot
	snapshotErr := s.DB.Where("tmdb_id = ? AND language = ?", tmdbID, normalizedLanguage).First(&snapshot).Error
	if snapshotErr == nil {
		cachedData, err := rewriteTMDBID(json.RawMessage(snapshot.TmdbData), tmdbID, resolveSyncTmdbID(snapshot.SyncTmdbID, syncTmdbID))
		if err != nil {
			return nil, err
		}
		if !isExpired(snapshot.LastSyncedAt, 24*time.Hour) {
			return mergeTMDBWithLocalData(cachedData, localData)
		}
	} else if !errors.Is(snapshotErr, gorm.ErrRecordNotFound) {
		return nil, snapshotErr
	}

	data, fetchErr := s.TmdbClient.GetTVSeries(syncTmdbID, opts)
	if fetchErr != nil {
		if snapshotErr == nil {
			cachedData, err := rewriteTMDBID(json.RawMessage(snapshot.TmdbData), tmdbID, resolveSyncTmdbID(snapshot.SyncTmdbID, syncTmdbID))
			if err != nil {
				return nil, err
			}
			return mergeTMDBWithLocalData(cachedData, localData)
		}
		if tvErr == nil {
			logx.Infof("TMDB 不可用，返回本地默认语言缓存: tv/%d language=%s", tmdbID, normalizedLanguage)
			return rewriteTMDBID(json.RawMessage(tv.TmdbData), tmdbID, syncTmdbID)
		}
		return nil, fetchErr
	}

	normalizedData, err := rewriteTMDBID(data, tmdbID, syncTmdbID)
	if err != nil {
		return nil, err
	}
	if err := s.upsertTVLangSnapshot(tmdbID, normalizedLanguage, syncTmdbID, normalizedData); err != nil {
		return nil, err
	}
	if errors.Is(tvErr, gorm.ErrRecordNotFound) {
		s.warmDefaultTVSeriesRecord(tmdbID, opts)
	}
	return mergeTMDBWithLocalData(normalizedData, localData)
}

func (s *ProxyService) getPersonDetailByLanguage(tmdbID int, opts *tmdbclient.RequestOption, language string) (json.RawMessage, error) {
	normalizedLanguage := normalizeLanguageTag(language)

	var person model.Person
	personErr := s.DB.Where("tmdb_id = ?", tmdbID).First(&person).Error
	var localData model.RawJSON
	if personErr == nil {
		localData = person.LocalData
	}
	if personErr != nil && !errors.Is(personErr, gorm.ErrRecordNotFound) {
		return nil, personErr
	}

	var snapshot model.PersonLangSnapshot
	snapshotErr := s.DB.Where("tmdb_id = ? AND language = ?", tmdbID, normalizedLanguage).First(&snapshot).Error
	if snapshotErr == nil && !isExpired(snapshot.LastSyncedAt, 48*time.Hour) {
		return mergeTMDBWithLocalData(json.RawMessage(snapshot.TmdbData), localData)
	}
	if snapshotErr != nil && !errors.Is(snapshotErr, gorm.ErrRecordNotFound) {
		return nil, snapshotErr
	}

	data, fetchErr := s.TmdbClient.GetPerson(tmdbID, opts)
	if fetchErr != nil {
		if snapshotErr == nil {
			return mergeTMDBWithLocalData(json.RawMessage(snapshot.TmdbData), localData)
		}
		if personErr == nil {
			logx.Infof("TMDB 不可用，返回本地默认语言缓存: person/%d language=%s", tmdbID, normalizedLanguage)
			return mergeTMDBWithLocalData(json.RawMessage(person.TmdbData), localData)
		}
		return nil, fetchErr
	}

	if err := s.upsertPersonLangSnapshot(tmdbID, normalizedLanguage, data); err != nil {
		return nil, err
	}
	if errors.Is(personErr, gorm.ErrRecordNotFound) {
		s.warmDefaultPersonRecord(tmdbID, opts)
	}
	return mergeTMDBWithLocalData(data, localData)
}

func (s *ProxyService) upsertMovieLangSnapshot(tmdbID int, language string, syncTmdbID int, data json.RawMessage) error {
	var snapshot model.MovieLangSnapshot
	now := time.Now()
	resolvedSyncTmdbID := resolveSyncTmdbID(syncTmdbID, tmdbID)

	err := s.DB.Where("tmdb_id = ? AND language = ?", tmdbID, language).First(&snapshot).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return s.DB.Create(&model.MovieLangSnapshot{
			TmdbID:       tmdbID,
			Language:     language,
			SyncTmdbID:   resolvedSyncTmdbID,
			TmdbData:     model.RawJSON(data),
			LastSyncedAt: &now,
		}).Error
	}
	if err != nil {
		return err
	}

	return s.DB.Model(&model.MovieLangSnapshot{}).Where("tmdb_id = ? AND language = ?", tmdbID, language).Updates(map[string]interface{}{
		"sync_tmdb_id":   resolvedSyncTmdbID,
		"tmdb_data":      model.RawJSON(data),
		"last_synced_at": &now,
	}).Error
}

func (s *ProxyService) upsertTVLangSnapshot(tmdbID int, language string, syncTmdbID int, data json.RawMessage) error {
	var snapshot model.TVLangSnapshot
	now := time.Now()
	resolvedSyncTmdbID := resolveSyncTmdbID(syncTmdbID, tmdbID)

	err := s.DB.Where("tmdb_id = ? AND language = ?", tmdbID, language).First(&snapshot).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return s.DB.Create(&model.TVLangSnapshot{
			TmdbID:       tmdbID,
			Language:     language,
			SyncTmdbID:   resolvedSyncTmdbID,
			TmdbData:     model.RawJSON(data),
			LastSyncedAt: &now,
		}).Error
	}
	if err != nil {
		return err
	}

	return s.DB.Model(&model.TVLangSnapshot{}).Where("tmdb_id = ? AND language = ?", tmdbID, language).Updates(map[string]interface{}{
		"sync_tmdb_id":   resolvedSyncTmdbID,
		"tmdb_data":      model.RawJSON(data),
		"last_synced_at": &now,
	}).Error
}

func (s *ProxyService) upsertPersonLangSnapshot(tmdbID int, language string, data json.RawMessage) error {
	var snapshot model.PersonLangSnapshot
	now := time.Now()

	err := s.DB.Where("tmdb_id = ? AND language = ?", tmdbID, language).First(&snapshot).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return s.DB.Create(&model.PersonLangSnapshot{
			TmdbID:       tmdbID,
			Language:     language,
			TmdbData:     model.RawJSON(data),
			LastSyncedAt: &now,
		}).Error
	}
	if err != nil {
		return err
	}

	return s.DB.Model(&model.PersonLangSnapshot{}).Where("tmdb_id = ? AND language = ?", tmdbID, language).Updates(map[string]interface{}{
		"tmdb_data":      model.RawJSON(data),
		"last_synced_at": &now,
	}).Error
}

func mergeTMDBWithLocalData(tmdbData json.RawMessage, localData model.RawJSON) (json.RawMessage, error) {
	if len(localData) == 0 {
		return tmdbData, nil
	}

	payload := map[string]interface{}{}
	if err := json.Unmarshal(tmdbData, &payload); err != nil {
		return nil, err
	}

	localPatch := map[string]interface{}{}
	if err := json.Unmarshal(localData, &localPatch); err != nil {
		return nil, err
	}

	for key, value := range localPatch {
		payload[key] = value
	}

	merged, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return merged, nil
}

func (s *ProxyService) warmDefaultMovieRecord(tmdbID int, opts *tmdbclient.RequestOption) {
	defaultOpts := cloneRequestOption(opts)
	defaultOpts.Language = s.DefaultLang
	if _, err := s.GetMovieDetail(tmdbID, defaultOpts); err != nil {
		logx.Errorf("写入默认语言电影缓存失败: movie/%d err=%v", tmdbID, err)
	}
}

func (s *ProxyService) warmDefaultTVSeriesRecord(tmdbID int, opts *tmdbclient.RequestOption) {
	defaultOpts := cloneRequestOption(opts)
	defaultOpts.Language = s.DefaultLang
	if _, err := s.GetTvSeriesDetail(tmdbID, defaultOpts); err != nil {
		logx.Errorf("写入默认语言剧集缓存失败: tv/%d err=%v", tmdbID, err)
	}
}

func (s *ProxyService) warmDefaultPersonRecord(tmdbID int, opts *tmdbclient.RequestOption) {
	defaultOpts := cloneRequestOption(opts)
	defaultOpts.Language = s.DefaultLang
	if _, err := s.GetPersonDetail(tmdbID, defaultOpts); err != nil {
		logx.Errorf("写入默认语言人物缓存失败: person/%d err=%v", tmdbID, err)
	}
}

func cloneRequestOption(opts *tmdbclient.RequestOption) *tmdbclient.RequestOption {
	if opts == nil {
		return &tmdbclient.RequestOption{}
	}
	cloned := *opts
	if opts.ExtraParams != nil {
		extra := make(map[string]string, len(opts.ExtraParams))
		for key, value := range opts.ExtraParams {
			extra[key] = value
		}
		cloned.ExtraParams = extra
	}
	return &cloned
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
