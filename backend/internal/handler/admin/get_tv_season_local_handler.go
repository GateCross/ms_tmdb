package admin

import (
	"net/http"

	"ms_tmdb/internal/logic/admin"
	"ms_tmdb/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

type adminTvSeasonLocalPathReq struct {
	Id           int `path:"id"`
	SeasonNumber int `path:"season_number"`
}

func GetTvSeasonLocalHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req adminTvSeasonLocalPathReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := admin.NewTvSeasonLocalLogic(r.Context(), svcCtx)
		data, saved, err := l.GetLocalSeason(req.Id, req.SeasonNumber)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, map[string]interface{}{
			"saved": saved,
			"data":  data,
		})
	}
}
