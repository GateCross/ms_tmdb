package main

import (
	"flag"
	"fmt"
	"net/http"

	"ms_tmdb/config"
	"ms_tmdb/internal/handler"
	adminhandler "ms_tmdb/internal/handler/admin"
	"ms_tmdb/internal/logging"
	adminlogic "ms_tmdb/internal/logic/admin"
	"ms_tmdb/internal/middleware"
	"ms_tmdb/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

var configFile = flag.String("f", "etc/tmdb.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	c.ConfigFile = *configFile

	server := rest.MustNewServer(c.RestConf)
	logging.SetupConsoleWriter(c.Log.Mode)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	autoSyncScheduler := adminlogic.NewLibraryAutoSyncScheduler(ctx)
	adminlogic.SetLibraryAutoSyncScheduler(autoSyncScheduler)
	autoSyncScheduler.Start()
	defer autoSyncScheduler.Stop()

	// 注册 TMDB 代理中间件。/api/v3 的文档路由由 goctl 生成，
	// 这里用全局中间件在路由命中后直接返回代理结果，避免模板 logic 接管真实请求。
	tmdbProxy := middleware.NewTmdbProxyMiddleware(ctx.TmdbClient, ctx.ProxyService)
	proxyHandler := tmdbProxy.Handle(func(w http.ResponseWriter, r *http.Request) {
		httpx.ErrorCtx(r.Context(), w, fmt.Errorf("未知路径: %s", r.URL.Path))
	})
	server.Use(tmdbProxy.Handle)
	handler.RegisterHandlers(server, ctx)
	registerTmdbProxyFallbackRoutes(server, proxyHandler)

	// 文件访问不在 tmdb.api 中声明，保留入口层的静态上传文件读取路由。
	server.AddRoutes(
		[]rest.Route{
			{Method: http.MethodGet, Path: "/:filename", Handler: adminhandler.GetUploadedFileHandler(ctx)},
		},
		rest.WithPrefix("/uploads"),
	)

	logx.Infof("服务启动: %s:%d", c.Host, c.Port)
	server.Start()
}

func registerTmdbProxyFallbackRoutes(server *rest.Server, handler http.HandlerFunc) {
	for _, prefix := range []string{"/api/v3", "/v3", "/3"} {
		server.AddRoutes(
			buildProxyRoutes(handler),
			rest.WithPrefix(prefix),
		)
	}
}

func buildProxyRoutes(handler http.HandlerFunc) []rest.Route {
	paths := []string{
		"/",
		"/:p1",
		"/:p1/:p2",
		"/:p1/:p2/:p3",
		"/:p1/:p2/:p3/:p4",
		"/:p1/:p2/:p3/:p4/:p5",
		"/:p1/:p2/:p3/:p4/:p5/:p6",
		"/:p1/:p2/:p3/:p4/:p5/:p6/:p7",
		"/:p1/:p2/:p3/:p4/:p5/:p6/:p7/:p8",
	}

	routes := make([]rest.Route, 0, len(paths))
	for _, path := range paths {
		routes = append(routes, rest.Route{
			Method:  http.MethodGet,
			Path:    path,
			Handler: handler,
		})
	}
	return routes
}
