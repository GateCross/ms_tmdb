package admin

import (
	"net/http"

	"ms_tmdb/internal/logic/admin"
	"ms_tmdb/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetAutoSyncSettingsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := admin.NewGetAutoSyncSettingsLogic(r.Context(), svcCtx)
		resp, err := l.GetAutoSyncSettings()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
