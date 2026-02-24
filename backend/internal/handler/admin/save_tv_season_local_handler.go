package admin

import (
	"net/http"

	"ms_tmdb/internal/logic/admin"
	"ms_tmdb/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

type adminTvSeasonLocalSaveReq struct {
	Id               int    `path:"id"`
	SeasonNumber     int    `path:"season_number"`
	Language         string `form:"language,optional"`
	AppendToResponse string `form:"append_to_response,optional"`
}

func SaveTvSeasonLocalHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req adminTvSeasonLocalSaveReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := admin.NewTvSeasonLocalLogic(r.Context(), svcCtx)
		data, err := l.SaveSeasonFromTMDB(req.Id, req.SeasonNumber, req.Language, req.AppendToResponse)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, map[string]interface{}{
			"saved":   true,
			"data":    data,
			"message": "季明细已保存到本地数据库",
		})
	}
}
