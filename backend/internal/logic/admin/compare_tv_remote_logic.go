package admin

import (
	"context"
	"errors"
	"fmt"

	"ms_tmdb/internal/model"
	"ms_tmdb/internal/svc"
	"ms_tmdb/internal/types"
	"ms_tmdb/pkg/tmdbclient"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type CompareTvRemoteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCompareTvRemoteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CompareTvRemoteLogic {
	return &CompareTvRemoteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CompareTvRemoteLogic) CompareTvRemote(req *types.AdminSyncReq) (resp *types.AdminCompareResp, err error) {
	if req.Id <= 0 {
		return nil, errors.New("无效剧集 ID")
	}

	var tv model.TVSeries
	if err := l.svcCtx.DB.Where("tmdb_id = ?", req.Id).First(&tv).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			remoteRaw, remoteErr := l.svcCtx.TmdbClient.GetTVSeries(req.Id, &tmdbclient.RequestOption{
				Context:          l.ctx,
				AppendToResponse: "credits,videos,images",
			})
			if remoteErr != nil {
				return nil, remoteErr
			}
			if _, parseErr := rawJSONToMap(model.RawJSON(remoteRaw)); parseErr != nil {
				return nil, parseErr
			}
			return &types.AdminCompareResp{
				HasDiff:                 true,
				DiffFields:              []string{"local_record_missing"},
				LocalOverrideDiffFields: []string{},
				DiffDetails: []types.AdminCompareFieldDetail{
					{
						Field:    "local_record_missing",
						DiffType: "remote",
						Local:    "本地不存在",
						Remote:   "TMDB 存在该条目",
					},
				},
				Message: "本地不存在该剧集数据，建议覆盖拉取",
			}, nil
		}
		return nil, err
	}

	remoteTmdbID := l.svcCtx.ProxyService.ResolveTVSyncID(req.Id)
	if remoteTmdbID <= 0 {
		return nil, errors.New("无效剧集 TMDB 同步 ID")
	}

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
	remoteData["id"] = tv.TmdbID

	localData, err := rawJSONToMap(tv.TmdbData)
	if err != nil {
		return nil, err
	}
	localPatch, err := rawJSONToMap(tv.LocalData)
	if err != nil {
		return nil, err
	}
	allDiffFields := diffTopLevelFields(localData, remoteData)
	diffFields, localOverrideDiffFields := splitDiffFieldsByLocalPatch(allDiffFields, localPatch, remoteData)
	diffFields = filterIgnoredRemoteDiffFields(diffFields)
	diffFields = filterEquivalentDiffFields(diffFields, localData, remoteData)
	localOverrideDiffFields = filterIgnoredRemoteDiffFields(localOverrideDiffFields)
	localOverrideDiffFields = filterEquivalentDiffFields(localOverrideDiffFields, localPatch, remoteData)
	diffDetails := buildCompareDiffDetails(diffFields, localOverrideDiffFields, localData, localPatch, remoteData)
	hasDiff := len(diffFields) > 0 || len(localOverrideDiffFields) > 0
	return &types.AdminCompareResp{
		HasDiff:                 hasDiff,
		DiffFields:              diffFields,
		LocalOverrideDiffFields: localOverrideDiffFields,
		DiffDetails:             diffDetails,
		Message:                 fmt.Sprintf("检测到远程差异 %d 项，本地修改字段 %d 项", len(diffFields), len(localOverrideDiffFields)),
	}, nil
}
