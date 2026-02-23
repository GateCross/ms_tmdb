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

type SyncMovieLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSyncMovieLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncMovieLogic {
	return &SyncMovieLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SyncMovieLogic) SyncMovie(req *types.AdminSyncReq) (*types.AdminSyncResp, error) {
	if req.Id <= 0 {
		return nil, errors.New("无效电影 ID")
	}

	mode := normalizeSyncMode(req.Mode)

	remoteRaw, err := l.svcCtx.TmdbClient.GetMovie(req.Id, &tmdbclient.RequestOption{
		AppendToResponse: "credits,videos,images",
	})
	if err != nil {
		return nil, err
	}

	remoteData, err := rawJSONToMap(model.RawJSON(remoteRaw))
	if err != nil {
		return nil, err
	}

	var movie model.Movie
	exists := true
	if err := l.svcCtx.DB.Where("tmdb_id = ?", req.Id).First(&movie).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			exists = false
		} else {
			return nil, err
		}
	}

	localPatch := make(map[string]interface{})
	if exists {
		localPatch, err = rawJSONToMap(movie.LocalData)
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
			"title":             mapString(finalData, "title"),
			"original_title":    mapString(finalData, "original_title"),
			"overview":          mapString(finalData, "overview"),
			"release_date":      mapString(finalData, "release_date"),
			"popularity":        mapFloat64(finalData, "popularity"),
			"vote_average":      mapFloat64(finalData, "vote_average"),
			"vote_count":        mapInt(finalData, "vote_count"),
			"poster_path":       mapString(finalData, "poster_path"),
			"backdrop_path":     mapString(finalData, "backdrop_path"),
			"original_language": mapString(finalData, "original_language"),
			"adult":             mapBool(finalData, "adult"),
			"status":            mapString(finalData, "status"),
			"runtime":           mapInt(finalData, "runtime"),
			"budget":            mapInt64(finalData, "budget"),
			"revenue":           mapInt64(finalData, "revenue"),
			"tagline":           mapString(finalData, "tagline"),
			"homepage":          mapString(finalData, "homepage"),
			"imdb_id":           mapString(finalData, "imdb_id"),
			"tmdb_data":         tmdbData,
			"local_data":        localData,
			"is_modified":       isModified,
			"last_synced_at":    &now,
		}
		if err := l.svcCtx.DB.Model(&model.Movie{}).Where("tmdb_id = ?", req.Id).Updates(updates).Error; err != nil {
			return nil, err
		}
	} else {
		record := model.Movie{
			TmdbID:           req.Id,
			Title:            mapString(finalData, "title"),
			OriginalTitle:    mapString(finalData, "original_title"),
			Overview:         mapString(finalData, "overview"),
			ReleaseDate:      mapString(finalData, "release_date"),
			Popularity:       mapFloat64(finalData, "popularity"),
			VoteAverage:      mapFloat64(finalData, "vote_average"),
			VoteCount:        mapInt(finalData, "vote_count"),
			PosterPath:       mapString(finalData, "poster_path"),
			BackdropPath:     mapString(finalData, "backdrop_path"),
			OriginalLanguage: mapString(finalData, "original_language"),
			Adult:            mapBool(finalData, "adult"),
			Status:           mapString(finalData, "status"),
			Runtime:          mapInt(finalData, "runtime"),
			Budget:           mapInt64(finalData, "budget"),
			Revenue:          mapInt64(finalData, "revenue"),
			Tagline:          mapString(finalData, "tagline"),
			Homepage:         mapString(finalData, "homepage"),
			ImdbID:           mapString(finalData, "imdb_id"),
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
		Message:         "电影数据同步完成",
	}, nil
}
