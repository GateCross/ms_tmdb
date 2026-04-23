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
		var req adminTvSeasonLocalPathReq
		if err := httpx.ParsePath(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		body := make(map[string]interface{})
		if err := httpx.ParseJsonBody(r, &body); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		payload := resolveTvSeasonLocalPayload(body)

		l := admin.NewTvSeasonLocalLogic(r.Context(), svcCtx)
		data, err := l.UpdateLocalSeason(req.Id, req.SeasonNumber, payload)
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

func resolveTvSeasonLocalPayload(body map[string]interface{}) map[string]interface{} {
	if len(body) == 0 {
		return map[string]interface{}{}
	}
	if raw, ok := body["payload"]; ok {
		if payload, ok := raw.(map[string]interface{}); ok {
			return payload
		}
	}
	return body
}
