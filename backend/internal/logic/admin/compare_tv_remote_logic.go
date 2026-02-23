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

	remoteRaw, err := l.svcCtx.TmdbClient.GetTVSeries(req.Id, &tmdbclient.RequestOption{
		AppendToResponse: "credits,videos,images",
	})
	if err != nil {
		return nil, err
	}

	remoteData, err := rawJSONToMap(model.RawJSON(remoteRaw))
	if err != nil {
		return nil, err
	}

	var tv model.TVSeries
	if err := l.svcCtx.DB.Where("tmdb_id = ?", req.Id).First(&tv).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &types.AdminCompareResp{
				HasDiff:    true,
				DiffFields: []string{"local_record_missing"},
				Message:    "本地不存在该剧集数据，建议覆盖拉取",
			}, nil
		}
		return nil, err
	}

	localData, err := rawJSONToMap(tv.TmdbData)
	if err != nil {
		return nil, err
	}
	diffFields := diffTopLevelFields(localData, remoteData)
	return &types.AdminCompareResp{
		HasDiff:    len(diffFields) > 0,
		DiffFields: diffFields,
		Message:    fmt.Sprintf("检测到 %d 项远程差异", len(diffFields)),
	}, nil
}
