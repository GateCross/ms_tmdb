package config

import (
	"fmt"

	"github.com/zeromicro/go-zero/rest"
)

// Config 服务配置
type Config struct {
	rest.RestConf

	// PostgreSQL 配置
	Postgres PostgresConf

	// TMDB 配置
	Tmdb struct {
		ApiKey          string
		BaseURL         string
		DefaultLanguage string
		RateLimit       int
	}

	// 缓存配置
	Cache struct {
		MovieTTL  int // 电影缓存时长(小时)
		TVTTL     int // 电视剧缓存时长(小时)
		PersonTTL int // 人物缓存时长(小时)
	}
}

// PostgresConf PostgreSQL 连接配置
type PostgresConf struct {
	Host     string `json:",optional"`
	Port     int    `json:",optional"`
	User     string `json:",optional"`
	Password string `json:",optional"`
	DBName   string `json:",optional"`
	SSLMode  string `json:",optional"`
}

// DSN 构建 GORM 使用的 PostgreSQL 连接串
func (p PostgresConf) DSN() string {
	host := p.Host
	if host == "" {
		host = "127.0.0.1"
	}

	port := p.Port
	if port == 0 {
		port = 5432
	}

	user := p.User
	if user == "" {
		user = "postgres"
	}

	dbName := p.DBName
	if dbName == "" {
		dbName = "ms_tmdb"
	}

	sslMode := p.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=%s", host, port, user, dbName, sslMode)
	if p.Password != "" {
		dsn += fmt.Sprintf(" password=%s", p.Password)
	}
	return dsn
}
