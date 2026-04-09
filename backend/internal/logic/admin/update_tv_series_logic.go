package admin

import (
	"context"
	"errors"

	"ms_tmdb/internal/model"
	"ms_tmdb/internal/svc"
	"ms_tmdb/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type UpdateTvSeriesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateTvSeriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTvSeriesLogic {
	return &UpdateTvSeriesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateTvSeriesLogic) UpdateTvSeries(req *types.AdminUpdateReq) error {
	if req.Id == 0 {
		return errors.New("无效剧集 ID")
	}

	var tv model.TVSeries
	if err := l.svcCtx.DB.Where("tmdb_id = ?", req.Id).First(&tv).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if req.Id < 0 {
				return errors.New("本地剧集不存在或已删除，请重新创建")
			}
			if _, syncErr := l.svcCtx.ProxyService.GetTvSeriesDetail(req.Id, nil); syncErr != nil {
				return syncErr
			}
			if retryErr := l.svcCtx.DB.Where("tmdb_id = ?", req.Id).First(&tv).Error; retryErr != nil {
				if errors.Is(retryErr, gorm.ErrRecordNotFound) {
					return errors.New("剧集未同步到本地，请刷新详情页后重试")
				}
				return retryErr
			}
		} else {
			return err
		}
	}

	patch := make(map[string]interface{})
	updates := map[string]interface{}{
		"is_modified": true,
	}

	hasFieldUpdate := false
	if req.TmdbId != nil {
		nextTmdbID := *req.TmdbId
		if nextTmdbID <= 0 {
			return errors.New("tmdb_id 必须大于 0")
		}
		if nextTmdbID != tv.TmdbID {
			var duplicated model.TVSeries
			if err := l.svcCtx.DB.Where("tmdb_id = ?", nextTmdbID).First(&duplicated).Error; err == nil {
				return errors.New("目标 tmdb_id 已存在，请使用其他 ID")
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			updates["tmdb_id"] = nextTmdbID
			updates["sync_tmdb_id"] = effectiveSyncTmdbID(tv.SyncTmdbID, tv.TmdbID)
			patch["id"] = nextTmdbID
			hasFieldUpdate = true
		}
	}

	if req.Name != nil {
		value := trimPtrString(req.Name)
		patch["name"] = value
		updates["name"] = value
		hasFieldUpdate = true
	}
	if req.OriginalName != nil {
		value := trimPtrString(req.OriginalName)
		patch["original_name"] = value
		updates["original_name"] = value
		hasFieldUpdate = true
	}
	if req.Overview != nil {
		value := trimPtrString(req.Overview)
		patch["overview"] = value
		updates["overview"] = value
		hasFieldUpdate = true
	}
	if req.Tagline != nil {
		value := trimPtrString(req.Tagline)
		patch["tagline"] = value
		updates["tagline"] = value
		hasFieldUpdate = true
	}
	if req.FirstAirDate != nil {
		value := trimPtrString(req.FirstAirDate)
		patch["first_air_date"] = value
		updates["first_air_date"] = value
		hasFieldUpdate = true
	}
	if req.Status != nil {
		value := trimPtrString(req.Status)
		patch["status"] = value
		updates["status"] = value
		hasFieldUpdate = true
	}
	if req.NumberOfSeasons != nil {
		patch["number_of_seasons"] = *req.NumberOfSeasons
		updates["number_of_seasons"] = *req.NumberOfSeasons
		hasFieldUpdate = true
	}
	if req.NumberOfEpisodes != nil {
		patch["number_of_episodes"] = *req.NumberOfEpisodes
		updates["number_of_episodes"] = *req.NumberOfEpisodes
		hasFieldUpdate = true
	}
	if req.OriginalLanguage != nil {
		value := trimPtrString(req.OriginalLanguage)
		patch["original_language"] = value
		updates["original_language"] = value
		hasFieldUpdate = true
	}
	if req.Homepage != nil {
		value := trimPtrString(req.Homepage)
		patch["homepage"] = value
		updates["homepage"] = value
		hasFieldUpdate = true
	}
	if req.PosterPath != nil {
		value := trimPtrString(req.PosterPath)
		patch["poster_path"] = value
		updates["poster_path"] = value
		hasFieldUpdate = true
	}
	if req.BackdropPath != nil {
		value := trimPtrString(req.BackdropPath)
		patch["backdrop_path"] = value
		updates["backdrop_path"] = value
		hasFieldUpdate = true
	}
	if req.VoteAverage != nil {
		patch["vote_average"] = *req.VoteAverage
		updates["vote_average"] = *req.VoteAverage
		hasFieldUpdate = true
	}
	if req.Popularity != nil {
		patch["popularity"] = *req.Popularity
		updates["popularity"] = *req.Popularity
		hasFieldUpdate = true
	}
	if req.Type != nil {
		value := trimPtrString(req.Type)
		patch["type"] = value
		updates["type"] = value
		hasFieldUpdate = true
	}
	if req.GenreNames != nil {
		genres := buildGenresFromNames(req.GenreNames)
		patch["genres"] = genres
		hasFieldUpdate = true
	}

	if !hasFieldUpdate {
		return errors.New("没有可更新的剧集字段")
	}

	mergedTMDBData, err := mergeTMDBData(tv.TmdbData, patch)
	if err != nil {
		return err
	}
	localPatch, err := rawJSONToMap(tv.LocalData)
	if err != nil {
		return err
	}
	localData, err := toRawJSON(mergeMap(localPatch, patch))
	if err != nil {
		return err
	}
	updates["tmdb_data"] = mergedTMDBData
	updates["local_data"] = localData

	return l.svcCtx.DB.Model(&model.TVSeries{}).Where("tmdb_id = ?", req.Id).Updates(updates).Error
}
