package admin

import (
	"context"

	"ms_tmdb/internal/model"
	"ms_tmdb/internal/svc"
	"ms_tmdb/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTvSeriesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListTvSeriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTvSeriesLogic {
	return &ListTvSeriesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListTvSeriesLogic) ListTvSeries(req *types.LibraryListReq) (*types.TvSeriesListResp, error) {
	page, pageSize := normalizePage(req.Page, req.PageSize)
	searchMode := normalizeSearchMode(req.SearchMode)

	var total int64
	query := applyKeywordFilter(
		l.svcCtx.DB.Model(&model.TVSeries{}),
		req.Keyword,
		searchMode,
		"tmdb_id",
		"name",
		"original_name",
	)
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var items []model.TVSeries
	if err := query.Select(tvSeriesListSelect).
		Order("popularity DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&items).Error; err != nil {
		return nil, err
	}

	results := make([]types.TvSeriesListItem, len(items))
	for i, m := range items {
		results[i] = types.TvSeriesListItem{
			TmdbId:           m.TmdbID,
			Name:             m.Name,
			OriginalName:     m.OriginalName,
			PosterPath:       m.PosterPath,
			VoteAverage:      m.VoteAverage,
			FirstAirDate:     m.FirstAirDate,
			NumberOfSeasons:  m.NumberOfSeasons,
			NumberOfEpisodes: m.NumberOfEpisodes,
			Popularity:       m.Popularity,
			IsModified:       m.IsModified,
			SkipSync:         m.SkipSync,
			GenreNames:       mergeGenreNames(genreNamesFromRaw(m.LocalData), genreNamesFromRaw(m.TmdbData)),
		}
	}

	return &types.TvSeriesListResp{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Results:  results,
	}, nil
}
