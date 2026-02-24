package admin

import (
	"net/http"

	"ms_tmdb/internal/logic/admin"
	"ms_tmdb/internal/svc"
	"ms_tmdb/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func UpdateAutoSyncSettingsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminAutoSyncReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := admin.NewUpdateAutoSyncSettingsLogic(r.Context(), svcCtx)
		resp, err := l.UpdateAutoSyncSettings(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
