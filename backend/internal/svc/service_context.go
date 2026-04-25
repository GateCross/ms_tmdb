package svc

import (
	"ms_tmdb/config"
	"ms_tmdb/internal/logic/proxy"
	"ms_tmdb/internal/model"
	"ms_tmdb/pkg/tmdbclient"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config       config.Config
	DB           *gorm.DB
	TmdbClient   *tmdbclient.Client
	ProxyService *proxy.ProxyService
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 初始化 PostgreSQL 连接
	db, err := gorm.Open(postgres.Open(c.Postgres.DSN()), &gorm.Config{})
	if err != nil {
		logx.Must(err)
	}

	// 自动建表迁移
	if err := model.AutoMigrate(db); err != nil {
		logx.Must(err)
	}
	// 启动时清理历史软删除残留数据（前端删除已改为物理删除）
	if err := model.CleanupSoftDeletedRows(db); err != nil {
		logx.Must(err)
	}
	if err := model.EnsureQueryIndexes(db); err != nil {
		logx.Must(err)
	}

	// 初始化 TMDB 客户端
	client := tmdbclient.NewClient(
		c.Tmdb.ApiKey,
		c.Tmdb.BaseURL,
		c.Tmdb.DefaultLanguage,
		c.Tmdb.RateLimit,
		c.Tmdb.ProxyURL,
	)

	return &ServiceContext{
		Config:       c,
		DB:           db,
		TmdbClient:   client,
		ProxyService: proxy.NewProxyService(db, client, c.Tmdb.DefaultLanguage, c.Tmdb.LocalWriteEnabled),
	}
}
