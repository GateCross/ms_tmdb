package main

import (
	"context"
	"flag"
	"net/http"

	"ms_tmdb/config"
	"ms_tmdb/internal/handler"
	adminhandler "ms_tmdb/internal/handler/admin"
	"ms_tmdb/internal/logging"
	adminlogic "ms_tmdb/internal/logic/admin"
	"ms_tmdb/internal/middleware"
	"ms_tmdb/internal/response"
	"ms_tmdb/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/tmdb.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	c.ConfigFile = *configFile

	logging.SetupConsoleWriter(c.Log.Mode)
	response.RegisterErrorHandler()

	ctx := svc.NewServiceContext(c)
	requestLog := middleware.NewRequestLogMiddleware(ctx.LogService)
	tmdbProxy := middleware.NewTmdbProxyMiddleware(ctx.TmdbClient, ctx.ProxyService)
	proxyHandler := requestLog.Handle(tmdbProxy.Proxy)
	proxyRouter := middleware.NewTmdbProxyRouter(http.HandlerFunc(proxyHandler))

	server := rest.MustNewServer(c.RestConf, rest.WithRouter(proxyRouter), rest.WithNotFoundHandler(nil))
	defer server.Stop()

	handler.RegisterHandlers(server, ctx)

	// 文件访问不在 tmdb.api 中声明，保留入口层的静态上传文件读取路由。
	server.AddRoutes(
		[]rest.Route{
			{Method: http.MethodGet, Path: "/:filename", Handler: adminhandler.GetUploadedFileHandler(ctx)},
		},
		rest.WithPrefix("/uploads"),
	)

	// 日志保留期清理移至后台异步执行，避免在启动时阻塞 HTTP 端口监听导致 502。
	stopLogCleaner := ctx.LogService.StartRetentionCleaner(context.Background())
	defer stopLogCleaner()

	autoSyncScheduler := adminlogic.NewLibraryAutoSyncScheduler(ctx)
	adminlogic.SetLibraryAutoSyncScheduler(autoSyncScheduler)
	autoSyncScheduler.Start()
	defer autoSyncScheduler.Stop()

	logx.Infof("服务启动: %s:%d", c.Host, c.Port)
	server.Start()
}
