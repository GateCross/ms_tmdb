package main

import (
	"flag"
	"fmt"
	"net/http"

	"ms_tmdb/config"
	"ms_tmdb/internal/handler/admin"
	"ms_tmdb/internal/middleware"
	"ms_tmdb/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
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
	defer server.Stop()

	ctx := svc.NewServiceContext(c)

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
			{Method: http.MethodPost, Path: "/movie", Handler: admin.CreateMovieHandler(ctx)},
			{Method: http.MethodDelete, Path: "/movie/:id", Handler: admin.DeleteMovieHandler(ctx)},
			{Method: http.MethodPut, Path: "/movie/:id", Handler: admin.UpdateMovieHandler(ctx)},
			{Method: http.MethodPost, Path: "/tv", Handler: admin.CreateTvSeriesHandler(ctx)},
			{Method: http.MethodDelete, Path: "/tv/:id", Handler: admin.DeleteTvSeriesHandler(ctx)},
			{Method: http.MethodPut, Path: "/tv/:id", Handler: admin.UpdateTvSeriesHandler(ctx)},
			{Method: http.MethodGet, Path: "/tv/:id/season/:season_number/local", Handler: admin.GetTvSeasonLocalHandler(ctx)},
			{Method: http.MethodPost, Path: "/tv/:id/season/:season_number/local", Handler: admin.SaveTvSeasonLocalHandler(ctx)},
			{Method: http.MethodPut, Path: "/tv/:id/season/:season_number/local", Handler: admin.UpdateTvSeasonLocalHandler(ctx)},
			{Method: http.MethodPut, Path: "/person/:id", Handler: admin.UpdatePersonHandler(ctx)},
			{Method: http.MethodGet, Path: "/compare/movie/:id", Handler: admin.CompareMovieRemoteHandler(ctx)},
			{Method: http.MethodGet, Path: "/compare/tv/:id", Handler: admin.CompareTvRemoteHandler(ctx)},
			{Method: http.MethodGet, Path: "/compare/person/:id", Handler: admin.ComparePersonRemoteHandler(ctx)},
			{Method: http.MethodPost, Path: "/sync/movie/:id", Handler: admin.SyncMovieHandler(ctx)},
			{Method: http.MethodPost, Path: "/sync/tv/:id", Handler: admin.SyncTvSeriesHandler(ctx)},
			{Method: http.MethodPost, Path: "/sync/person/:id", Handler: admin.SyncPersonHandler(ctx)},
			{Method: http.MethodDelete, Path: "/movie/:id/local", Handler: admin.ClearMovieLocalHandler(ctx)},
			{Method: http.MethodDelete, Path: "/tv/:id/local", Handler: admin.ClearTvSeriesLocalHandler(ctx)},
			{Method: http.MethodGet, Path: "/stats", Handler: admin.GetStatsHandler(ctx)},
			{Method: http.MethodGet, Path: "/movies", Handler: admin.ListMoviesHandler(ctx)},
			{Method: http.MethodGet, Path: "/tv", Handler: admin.ListTvSeriesHandler(ctx)},
			{Method: http.MethodGet, Path: "/proxy", Handler: admin.GetProxySettingsHandler(ctx)},
			{Method: http.MethodPut, Path: "/proxy", Handler: admin.UpdateProxySettingsHandler(ctx)},
			{Method: http.MethodPost, Path: "/upload/image", Handler: admin.UploadImageHandler(ctx)},
		},
		rest.WithPrefix("/api/admin"),
	)

	server.AddRoutes(
		[]rest.Route{
			{Method: http.MethodGet, Path: "/:filename", Handler: admin.GetUploadedFileHandler(ctx)},
		},
		rest.WithPrefix("/uploads"),
	)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
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
