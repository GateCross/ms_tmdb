package admin

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"ms_tmdb/internal/logic/admin"
	"ms_tmdb/internal/svc"
	"ms_tmdb/internal/types"
)

func UpdateProxySettingsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminProxyReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := admin.NewUpdateProxySettingsLogic(r.Context(), svcCtx)
		resp, err := l.UpdateProxySettings(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
