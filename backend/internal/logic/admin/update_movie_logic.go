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

type UpdateMovieLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateMovieLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateMovieLogic {
	return &UpdateMovieLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateMovieLogic) UpdateMovie(req *types.AdminUpdateReq) error {
	if req.Id == 0 {
		return errors.New("无效电影 ID")
	}

	var movie model.Movie
	if err := l.svcCtx.DB.Where("tmdb_id = ?", req.Id).First(&movie).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if req.Id < 0 {
				return errors.New("本地电影不存在或已删除，请重新创建")
			}
			if _, syncErr := l.svcCtx.ProxyService.GetMovieDetail(req.Id, nil); syncErr != nil {
				return syncErr
			}
			if retryErr := l.svcCtx.DB.Where("tmdb_id = ?", req.Id).First(&movie).Error; retryErr != nil {
				if errors.Is(retryErr, gorm.ErrRecordNotFound) {
					return errors.New("电影未同步到本地，请刷新详情页后重试")
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
		if nextTmdbID != movie.TmdbID {
			var duplicated model.Movie
			if err := l.svcCtx.DB.Where("tmdb_id = ?", nextTmdbID).First(&duplicated).Error; err == nil {
				return errors.New("目标 tmdb_id 已存在，请使用其他 ID")
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			updates["tmdb_id"] = nextTmdbID
			updates["sync_tmdb_id"] = effectiveSyncTmdbID(movie.SyncTmdbID, movie.TmdbID)
			patch["id"] = nextTmdbID
			hasFieldUpdate = true
		}
	}

	if req.Title != nil {
		value := trimPtrString(req.Title)
		patch["title"] = value
		updates["title"] = value
		hasFieldUpdate = true
	}
	if req.OriginalTitle != nil {
		value := trimPtrString(req.OriginalTitle)
		patch["original_title"] = value
		updates["original_title"] = value
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
	if req.ReleaseDate != nil {
		value := trimPtrString(req.ReleaseDate)
		patch["release_date"] = value
		updates["release_date"] = value
		hasFieldUpdate = true
	}
	if req.Status != nil {
		value := trimPtrString(req.Status)
		patch["status"] = value
		updates["status"] = value
		hasFieldUpdate = true
	}
	if req.Runtime != nil {
		patch["runtime"] = *req.Runtime
		updates["runtime"] = *req.Runtime
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
	if req.GenreNames != nil {
		genres := buildGenresFromNames(req.GenreNames)
		patch["genres"] = genres
		hasFieldUpdate = true
	}

	if !hasFieldUpdate {
		return errors.New("没有可更新的电影字段")
	}

	mergedTMDBData, err := mergeTMDBData(movie.TmdbData, patch)
	if err != nil {
		return err
	}
	localPatch, err := rawJSONToMap(movie.LocalData)
	if err != nil {
		return err
	}
	localData, err := toRawJSON(mergeMap(localPatch, patch))
	if err != nil {
		return err
	}
	updates["tmdb_data"] = mergedTMDBData
	updates["local_data"] = localData

	return l.svcCtx.DB.Model(&model.Movie{}).Where("tmdb_id = ?", req.Id).Updates(updates).Error
}
