package main

import (
	"flag"
	"fmt"
	"net/http"

	"ms_tmdb/config"
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

	// 注册 TMDB 代理中间件（拦截所有 /api/v3/* 请求）
	tmdbProxy := middleware.NewTmdbProxyMiddleware(ctx.TmdbClient, ctx.ProxyService)
	proxyHandler := tmdbProxy.Handle(func(w http.ResponseWriter, r *http.Request) {
		httpx.ErrorCtx(r.Context(), w, fmt.Errorf("未知路径: %s", r.URL.Path))
	})
	for _, prefix := range []string{"/api/v3", "/v3", "/3"} {
		server.AddRoutes(
			buildProxyRoutes(proxyHandler),
			rest.WithPrefix(prefix),
		)
	}

	// 注册 Admin 路由
	server.AddRoutes(
		[]rest.Route{
			{Method: http.MethodPost, Path: "/movie", Handler: adminhandler.CreateMovieHandler(ctx)},
			{Method: http.MethodDelete, Path: "/movie/:id", Handler: adminhandler.DeleteMovieHandler(ctx)},
			{Method: http.MethodPut, Path: "/movie/:id", Handler: adminhandler.UpdateMovieHandler(ctx)},
			{Method: http.MethodPost, Path: "/tv", Handler: adminhandler.CreateTvSeriesHandler(ctx)},
			{Method: http.MethodDelete, Path: "/tv/:id", Handler: adminhandler.DeleteTvSeriesHandler(ctx)},
			{Method: http.MethodPut, Path: "/tv/:id", Handler: adminhandler.UpdateTvSeriesHandler(ctx)},
			{Method: http.MethodGet, Path: "/tv/:id/season/:season_number/local", Handler: adminhandler.GetTvSeasonLocalHandler(ctx)},
			{Method: http.MethodPost, Path: "/tv/:id/season/:season_number/local", Handler: adminhandler.SaveTvSeasonLocalHandler(ctx)},
			{Method: http.MethodPut, Path: "/tv/:id/season/:season_number/local", Handler: adminhandler.UpdateTvSeasonLocalHandler(ctx)},
			{Method: http.MethodDelete, Path: "/tv/:id/season/:season_number/local", Handler: adminhandler.DeleteTvSeasonLocalHandler(ctx)},
			{Method: http.MethodPut, Path: "/person/:id", Handler: adminhandler.UpdatePersonHandler(ctx)},
			{Method: http.MethodGet, Path: "/compare/movie/:id", Handler: adminhandler.CompareMovieRemoteHandler(ctx)},
			{Method: http.MethodGet, Path: "/compare/tv/:id", Handler: adminhandler.CompareTvRemoteHandler(ctx)},
			{Method: http.MethodGet, Path: "/compare/person/:id", Handler: adminhandler.ComparePersonRemoteHandler(ctx)},
			{Method: http.MethodPost, Path: "/sync/movie/:id", Handler: adminhandler.SyncMovieHandler(ctx)},
			{Method: http.MethodPost, Path: "/sync/tv/:id", Handler: adminhandler.SyncTvSeriesHandler(ctx)},
			{Method: http.MethodPost, Path: "/sync/person/:id", Handler: adminhandler.SyncPersonHandler(ctx)},
			{Method: http.MethodDelete, Path: "/movie/:id/local", Handler: adminhandler.ClearMovieLocalHandler(ctx)},
			{Method: http.MethodDelete, Path: "/tv/:id/local", Handler: adminhandler.ClearTvSeriesLocalHandler(ctx)},
			{Method: http.MethodGet, Path: "/stats", Handler: adminhandler.GetStatsHandler(ctx)},
			{Method: http.MethodGet, Path: "/movies", Handler: adminhandler.ListMoviesHandler(ctx)},
			{Method: http.MethodGet, Path: "/tv-series", Handler: adminhandler.ListTvSeriesHandler(ctx)},
			{Method: http.MethodGet, Path: "/proxy", Handler: adminhandler.GetProxySettingsHandler(ctx)},
			{Method: http.MethodPut, Path: "/proxy", Handler: adminhandler.UpdateProxySettingsHandler(ctx)},
			{Method: http.MethodGet, Path: "/auto-sync", Handler: adminhandler.GetAutoSyncSettingsHandler(ctx)},
			{Method: http.MethodPut, Path: "/auto-sync", Handler: adminhandler.UpdateAutoSyncSettingsHandler(ctx)},
			{Method: http.MethodPost, Path: "/auto-sync/run", Handler: adminhandler.RunAutoSyncNowHandler(ctx)},
			{Method: http.MethodGet, Path: "/auto-sync/logs", Handler: adminhandler.ListAutoSyncLogsHandler(ctx)},
			{Method: http.MethodDelete, Path: "/auto-sync/logs", Handler: adminhandler.ClearAutoSyncLogsHandler(ctx)},
			{Method: http.MethodGet, Path: "/auto-sync/logs/:id", Handler: adminhandler.GetAutoSyncLogDetailHandler(ctx)},
			{Method: http.MethodPost, Path: "/upload/image", Handler: adminhandler.UploadImageHandler(ctx)},
		},
		rest.WithPrefix("/api/admin"),
	)

	server.AddRoutes(
		[]rest.Route{
			{Method: http.MethodGet, Path: "/:filename", Handler: adminhandler.GetUploadedFileHandler(ctx)},
		},
		rest.WithPrefix("/uploads"),
	)

	logx.Infof("服务启动: %s:%d", c.Host, c.Port)
	server.Start()
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
