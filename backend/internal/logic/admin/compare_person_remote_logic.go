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

type ComparePersonRemoteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewComparePersonRemoteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ComparePersonRemoteLogic {
	return &ComparePersonRemoteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ComparePersonRemoteLogic) ComparePersonRemote(req *types.AdminSyncReq) (resp *types.AdminCompareResp, err error) {
	if req.Id <= 0 {
		return nil, errors.New("无效人物 ID")
	}

	remoteRaw, err := l.svcCtx.TmdbClient.GetPerson(req.Id, &tmdbclient.RequestOption{
		Context:          l.ctx,
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
	if err := l.svcCtx.DB.Where("tmdb_id = ?", req.Id).First(&person).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
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
				Message: "本地不存在该人物数据，建议覆盖拉取",
			}, nil
		}
		return nil, err
	}

	localData, err := rawJSONToMap(person.TmdbData)
	if err != nil {
		return nil, err
	}
	diffFields := diffTopLevelFields(localData, remoteData)
	diffFields = filterIgnoredRemoteDiffFields(diffFields)
	diffFields = filterEquivalentDiffFields(diffFields, localData, remoteData)
	diffDetails := buildCompareDiffDetails(diffFields, []string{}, localData, localData, remoteData)
	return &types.AdminCompareResp{
		HasDiff:                 len(diffFields) > 0,
		DiffFields:              diffFields,
		LocalOverrideDiffFields: []string{},
		DiffDetails:             diffDetails,
		Message:                 fmt.Sprintf("检测到 %d 项远程差异", len(diffFields)),
	}, nil
}
