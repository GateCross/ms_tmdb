package admin

import (
	"net/http"

	"ms_tmdb/internal/logic/admin"
	"ms_tmdb/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

type adminTvSeasonLocalDeleteReq struct {
	Id           int `path:"id"`
	SeasonNumber int `path:"season_number"`
}

func DeleteTvSeasonLocalHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req adminTvSeasonLocalDeleteReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := admin.NewTvSeasonLocalLogic(r.Context(), svcCtx)
		if err := l.DeleteLocalSeason(req.Id, req.SeasonNumber); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, map[string]interface{}{
			"saved":   false,
			"data":    nil,
			"message": "季本地数据已删除",
		})
	}
}
