package admin

import (
	"net/http"

	"ms_tmdb/internal/logic/admin"
	"ms_tmdb/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

type adminTvSeasonLocalUpdateReq struct {
	Id           int                    `path:"id"`
	SeasonNumber int                    `path:"season_number"`
	Payload      map[string]interface{} `json:"payload"`
}

func UpdateTvSeasonLocalHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req adminTvSeasonLocalUpdateReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := admin.NewTvSeasonLocalLogic(r.Context(), svcCtx)
		data, err := l.UpdateLocalSeason(req.Id, req.SeasonNumber, req.Payload)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, map[string]interface{}{
			"saved":   true,
			"data":    data,
			"message": "季明细本地修改已保存",
		})
	}
}
