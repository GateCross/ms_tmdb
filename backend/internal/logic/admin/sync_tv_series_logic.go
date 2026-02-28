package admin

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"ms_tmdb/internal/model"
	"ms_tmdb/internal/svc"
	"ms_tmdb/internal/types"
	"ms_tmdb/pkg/tmdbclient"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type SyncTvSeriesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSyncTvSeriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncTvSeriesLogic {
	return &SyncTvSeriesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SyncTvSeriesLogic) SyncTvSeries(req *types.AdminSyncReq) (*types.AdminSyncResp, error) {
	if req.Id <= 0 {
		return nil, errors.New("无效剧集 ID")
	}

	mode := normalizeSyncMode(req.Mode)

	var tv model.TVSeries
	exists := true
	if err := l.svcCtx.DB.Where("tmdb_id = ?", req.Id).First(&tv).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			exists = false
		} else {
			return nil, err
		}
	}

	remoteTmdbID := l.svcCtx.ProxyService.ResolveTVSyncID(req.Id)

	remoteRaw, err := l.svcCtx.TmdbClient.GetTVSeries(remoteTmdbID, &tmdbclient.RequestOption{
		Context:          l.ctx,
		AppendToResponse: "credits,videos,images",
	})
	if err != nil {
		return nil, err
	}

	remoteData, err := rawJSONToMap(model.RawJSON(remoteRaw))
	if err != nil {
		return nil, err
	}
	remoteData["id"] = req.Id

	localPatch := make(map[string]interface{})
	if exists {
		localPatch, err = rawJSONToMap(tv.LocalData)
		if err != nil {
			return nil, err
		}
	}

	localPatch = sanitizeLocalPatch(localPatch, remoteData)
	changedFields := sortedKeys(localPatch)
	if mode == syncModePreview {
		return &types.AdminSyncResp{
			Mode:            mode,
			ChangedFields:   changedFields,
			Overwritten:     []string{},
			KeptLocalFields: changedFields,
			IsModified:      len(changedFields) > 0,
			Message:         fmt.Sprintf("检测到 %d 个有变化字段", len(changedFields)),
		}, nil
	}

	remainingPatch := localPatch
	overwritten := make([]string, 0)

	switch mode {
	case syncModeOverwriteAll:
		remainingPatch = map[string]interface{}{}
		overwritten = changedFields
	case syncModeSelective:
		pendingOverwrite := make(map[string]struct{}, len(req.OverwriteFields))
		for _, field := range req.OverwriteFields {
			name := strings.TrimSpace(field)
			if name == "" {
				continue
			}
			if _, ok := localPatch[name]; ok {
				pendingOverwrite[name] = struct{}{}
			}
		}

		remainingPatch = removeFieldsFromPatch(localPatch, req.OverwriteFields)
		overwritten = make([]string, 0, len(pendingOverwrite))
		for field := range pendingOverwrite {
			overwritten = append(overwritten, field)
		}
		sort.Strings(overwritten)
	default:
		mode = syncModeUpdateUnchanged
	}

	finalData := mergeMap(remoteData, remainingPatch)
	tmdbData, err := toRawJSON(finalData)
	if err != nil {
		return nil, err
	}

	var localData model.RawJSON
	if len(remainingPatch) > 0 {
		localData, err = toRawJSON(remainingPatch)
		if err != nil {
			return nil, err
		}
	}

	now := time.Now()
	isModified := len(remainingPatch) > 0

	if exists {
		updates := map[string]interface{}{
			"name":               mapString(finalData, "name"),
			"original_name":      mapString(finalData, "original_name"),
			"overview":           mapString(finalData, "overview"),
			"first_air_date":     mapString(finalData, "first_air_date"),
			"last_air_date":      mapString(finalData, "last_air_date"),
			"popularity":         mapFloat64(finalData, "popularity"),
			"vote_average":       mapFloat64(finalData, "vote_average"),
			"vote_count":         mapInt(finalData, "vote_count"),
			"poster_path":        mapString(finalData, "poster_path"),
			"backdrop_path":      mapString(finalData, "backdrop_path"),
			"original_language":  mapString(finalData, "original_language"),
			"status":             mapString(finalData, "status"),
			"type":               mapString(finalData, "type"),
			"number_of_seasons":  mapInt(finalData, "number_of_seasons"),
			"number_of_episodes": mapInt(finalData, "number_of_episodes"),
			"homepage":           mapString(finalData, "homepage"),
			"in_production":      mapBool(finalData, "in_production"),
			"tagline":            mapString(finalData, "tagline"),
			"tmdb_data":          tmdbData,
			"local_data":         localData,
			"is_modified":        isModified,
			"sync_tmdb_id":       remoteTmdbID,
			"last_synced_at":     &now,
		}
		if err := l.svcCtx.DB.Model(&model.TVSeries{}).Where("tmdb_id = ?", req.Id).Updates(updates).Error; err != nil {
			return nil, err
		}
	} else {
		record := model.TVSeries{
			TmdbID:           req.Id,
			SyncTmdbID:       remoteTmdbID,
			Name:             mapString(finalData, "name"),
			OriginalName:     mapString(finalData, "original_name"),
			Overview:         mapString(finalData, "overview"),
			FirstAirDate:     mapString(finalData, "first_air_date"),
			LastAirDate:      mapString(finalData, "last_air_date"),
			Popularity:       mapFloat64(finalData, "popularity"),
			VoteAverage:      mapFloat64(finalData, "vote_average"),
			VoteCount:        mapInt(finalData, "vote_count"),
			PosterPath:       mapString(finalData, "poster_path"),
			BackdropPath:     mapString(finalData, "backdrop_path"),
			OriginalLanguage: mapString(finalData, "original_language"),
			Status:           mapString(finalData, "status"),
			Type:             mapString(finalData, "type"),
			NumberOfSeasons:  mapInt(finalData, "number_of_seasons"),
			NumberOfEpisodes: mapInt(finalData, "number_of_episodes"),
			Homepage:         mapString(finalData, "homepage"),
			InProduction:     mapBool(finalData, "in_production"),
			Tagline:          mapString(finalData, "tagline"),
			TmdbData:         tmdbData,
			LocalData:        localData,
			IsModified:       isModified,
			LastSyncedAt:     &now,
		}
		if err := l.svcCtx.DB.Create(&record).Error; err != nil {
			return nil, err
		}
	}

	return &types.AdminSyncResp{
		Mode:            mode,
		ChangedFields:   changedFields,
		Overwritten:     overwritten,
		KeptLocalFields: sortedKeys(remainingPatch),
		IsModified:      isModified,
		Message:         "剧集数据同步完成",
	}, nil
}
