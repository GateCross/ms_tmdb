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

type SyncPersonLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSyncPersonLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncPersonLogic {
	return &SyncPersonLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SyncPersonLogic) SyncPerson(req *types.AdminSyncReq) (*types.AdminSyncResp, error) {
	if req.Id <= 0 {
		return nil, errors.New("无效人物 ID")
	}

	mode := normalizeSyncMode(req.Mode)

	remoteRaw, err := l.svcCtx.TmdbClient.GetPerson(req.Id, &tmdbclient.RequestOption{
		AppendToResponse: "combined_credits,images",
	})
	if err != nil {
		return nil, err
	}

	remoteData, err := rawJSONToMap(model.RawJSON(remoteRaw))
	if err != nil {
		return nil, err
	}

	var person model.Person
	exists := true
	if err := l.svcCtx.DB.Where("tmdb_id = ?", req.Id).First(&person).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			exists = false
		} else {
			return nil, err
		}
	}

	localPatch := make(map[string]interface{})
	if exists {
		localPatch, err = rawJSONToMap(person.LocalData)
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
			"name":                 mapString(finalData, "name"),
			"biography":            mapString(finalData, "biography"),
			"birthday":             mapString(finalData, "birthday"),
			"deathday":             mapString(finalData, "deathday"),
			"gender":               mapInt(finalData, "gender"),
			"known_for_department": mapString(finalData, "known_for_department"),
			"place_of_birth":       mapString(finalData, "place_of_birth"),
			"popularity":           mapFloat64(finalData, "popularity"),
			"profile_path":         mapString(finalData, "profile_path"),
			"adult":                mapBool(finalData, "adult"),
			"imdb_id":              mapString(finalData, "imdb_id"),
			"homepage":             mapString(finalData, "homepage"),
			"tmdb_data":            tmdbData,
			"local_data":           localData,
			"is_modified":          isModified,
			"last_synced_at":       &now,
		}
		if err := l.svcCtx.DB.Model(&model.Person{}).Where("tmdb_id = ?", req.Id).Updates(updates).Error; err != nil {
			return nil, err
		}
	} else {
		record := model.Person{
			TmdbID:             req.Id,
			Name:               mapString(finalData, "name"),
			Biography:          mapString(finalData, "biography"),
			Birthday:           mapString(finalData, "birthday"),
			Deathday:           mapString(finalData, "deathday"),
			Gender:             mapInt(finalData, "gender"),
			KnownForDepartment: mapString(finalData, "known_for_department"),
			PlaceOfBirth:       mapString(finalData, "place_of_birth"),
			Popularity:         mapFloat64(finalData, "popularity"),
			ProfilePath:        mapString(finalData, "profile_path"),
			Adult:              mapBool(finalData, "adult"),
			ImdbID:             mapString(finalData, "imdb_id"),
			Homepage:           mapString(finalData, "homepage"),
			TmdbData:           tmdbData,
			LocalData:          localData,
			IsModified:         isModified,
			LastSyncedAt:       &now,
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
		Message:         "人物数据同步完成",
	}, nil
}
