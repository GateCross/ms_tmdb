package middleware

import (
	"net/http"
	"strings"

	"ms_tmdb/internal/logic/proxy"
	"ms_tmdb/pkg/tmdbclient"

	"github.com/zeromicro/go-zero/core/logx"
)

// TmdbProxyMiddleware TMDB API 代理中间件
// 拦截所有 /api/v3/*、/v3/*、/3/* 请求，直接代理到 TMDB 并返回原始 JSON
type TmdbProxyMiddleware struct {
	Client       *tmdbclient.Client
	ProxyService *proxy.ProxyService
	dispatcher   *tmdbRouteDispatcher
}

func NewTmdbProxyMiddleware(client *tmdbclient.Client, proxyService *proxy.ProxyService) *TmdbProxyMiddleware {
	return &TmdbProxyMiddleware{
		Client:       client,
		ProxyService: proxyService,
		dispatcher:   newTmdbRouteDispatcher(client, proxyService),
	}
}

func (m *TmdbProxyMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 提取版本前缀后的 TMDB 路径
		tmdbPath, ok := resolveTmdbPath(r.URL.Path)
		if !ok {
			next(w, r)
			return
		}

		data, err := m.dispatcher.dispatch(tmdbPath, parseRequestOptions(r), r)
		if err != nil {
			logx.Errorf("TMDB 代理请求失败: %s, 错误: %v", tmdbPath, err)
			writeProxyError(w, http.StatusBadGateway)
			return
		}

		writeJSONResponse(w, data)
	}
}

func resolveTmdbPath(path string) (string, bool) {
	for _, prefix := range []string{"/api/v3", "/v3", "/3"} {
		if strings.HasPrefix(path, prefix) {
			return strings.TrimPrefix(path, prefix), true
		}
	}
	return "", false
}
