package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
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
}

func NewTmdbProxyMiddleware(client *tmdbclient.Client, proxyService *proxy.ProxyService) *TmdbProxyMiddleware {
	return &TmdbProxyMiddleware{Client: client, ProxyService: proxyService}
}

func (m *TmdbProxyMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 提取版本前缀后的 TMDB 路径
		tmdbPath, ok := resolveTmdbPath(r.URL.Path)
		if !ok {
			next(w, r)
			return
		}

		// 从查询参数构建请求选项
		opts := m.parseOpts(r)

		// 根据路径模式匹配处理
		data, err := m.dispatch(tmdbPath, opts, r)
		if err != nil {
			logx.Errorf("TMDB 代理请求失败: %s, 错误: %v", tmdbPath, err)
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}

		writeJSON(w, data)
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

// parseOpts 从查询参数中提取请求选项
func (m *TmdbProxyMiddleware) parseOpts(r *http.Request) *tmdbclient.RequestOption {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))

	opts := &tmdbclient.RequestOption{
		Language:         q.Get("language"),
		Page:             page,
		Region:           q.Get("region"),
		AppendToResponse: q.Get("append_to_response"),
		ExtraParams:      map[string]string{},
	}

	// 把所有未识别的查询参数作为 ExtraParams 透传
	knownKeys := map[string]bool{
		"language": true, "page": true, "region": true,
		"append_to_response": true, "api_key": true,
	}
	for k, v := range q {
		if !knownKeys[k] && len(v) > 0 {
			opts.ExtraParams[k] = v[0]
		}
	}

	return opts
}

// 路径匹配正则
var (
	reMovieDetail  = regexp.MustCompile(`^/movie/(-?\d+)$`)
	reMovieSub     = regexp.MustCompile(`^/movie/(-?\d+)/(.+)$`)
	reTVDetail     = regexp.MustCompile(`^/tv/(-?\d+)$`)
	reTVSub        = regexp.MustCompile(`^/tv/(-?\d+)/(.+)$`)
	reTVSeason     = regexp.MustCompile(`^/tv/(-?\d+)/season/(\d+)$`)
	reTVSeasonSub  = regexp.MustCompile(`^/tv/(-?\d+)/season/(\d+)/(.+)$`)
	reTVEpisode    = regexp.MustCompile(`^/tv/(-?\d+)/season/(\d+)/episode/(\d+)$`)
	reTVEpisodeSub = regexp.MustCompile(`^/tv/(-?\d+)/season/(\d+)/episode/(\d+)/(.+)$`)
	rePersonDetail = regexp.MustCompile(`^/person/(-?\d+)$`)
	rePersonSub    = regexp.MustCompile(`^/person/(-?\d+)/(.+)$`)
	reTrending     = regexp.MustCompile(`^/trending/(\w+)/(\w+)$`)
	reFind         = regexp.MustCompile(`^/find/(.+)$`)
)

// dispatch 根据路径分发请求到对应的 TMDB API
func (m *TmdbProxyMiddleware) dispatch(path string, opts *tmdbclient.RequestOption, r *http.Request) (json.RawMessage, error) {
	q := r.URL.Query()

	// 电影详情（Read-Through 缓存）
	if matches := reMovieDetail.FindStringSubmatch(path); matches != nil {
		id, _ := strconv.Atoi(matches[1])
		return m.ProxyService.GetMovieDetail(id, opts)
	}

	// 电影子资源
	if matches := reMovieSub.FindStringSubmatch(path); matches != nil {
		id, _ := strconv.Atoi(matches[1])
		sub := matches[2]
		return m.Client.Request(fmt.Sprintf("/movie/%d/%s", id, sub), opts)
	}

	// TV 剧集（先匹配更具体的路径）
	if matches := reTVEpisodeSub.FindStringSubmatch(path); matches != nil {
		seriesID, _ := strconv.Atoi(matches[1])
		seasonNum, _ := strconv.Atoi(matches[2])
		epNum, _ := strconv.Atoi(matches[3])
		sub := matches[4]
		return m.Client.Request(fmt.Sprintf("/tv/%d/season/%d/episode/%d/%s", seriesID, seasonNum, epNum, sub), opts)
	}
	if matches := reTVEpisode.FindStringSubmatch(path); matches != nil {
		seriesID, _ := strconv.Atoi(matches[1])
		seasonNum, _ := strconv.Atoi(matches[2])
		epNum, _ := strconv.Atoi(matches[3])
		return m.Client.GetTVEpisode(seriesID, seasonNum, epNum, opts)
	}

	// TV 季子资源
	if matches := reTVSeasonSub.FindStringSubmatch(path); matches != nil {
		seriesID, _ := strconv.Atoi(matches[1])
		seasonNum, _ := strconv.Atoi(matches[2])
		sub := matches[3]
		// 排除 episode 路径（已被上面匹配）
		if strings.HasPrefix(sub, "episode/") {
			return nil, fmt.Errorf("未匹配到路径: %s", path)
		}
		return m.Client.Request(fmt.Sprintf("/tv/%d/season/%d/%s", seriesID, seasonNum, sub), opts)
	}
	if matches := reTVSeason.FindStringSubmatch(path); matches != nil {
		seriesID, _ := strconv.Atoi(matches[1])
		seasonNum, _ := strconv.Atoi(matches[2])
		return m.ProxyService.GetTvSeasonDetail(seriesID, seasonNum, opts)
	}

	// TV 详情（Read-Through 缓存）
	if matches := reTVDetail.FindStringSubmatch(path); matches != nil {
		id, _ := strconv.Atoi(matches[1])
		return m.ProxyService.GetTvSeriesDetail(id, opts)
	}

	// TV 子资源
	if matches := reTVSub.FindStringSubmatch(path); matches != nil {
		id, _ := strconv.Atoi(matches[1])
		sub := matches[2]
		// 排除 season 路径
		if strings.HasPrefix(sub, "season/") {
			return nil, fmt.Errorf("未匹配到路径: %s", path)
		}
		return m.Client.Request(fmt.Sprintf("/tv/%d/%s", id, sub), opts)
	}

	// 人物详情（Read-Through 缓存）
	if matches := rePersonDetail.FindStringSubmatch(path); matches != nil {
		id, _ := strconv.Atoi(matches[1])
		return m.ProxyService.GetPersonDetail(id, opts)
	}

	// 人物子资源
	if matches := rePersonSub.FindStringSubmatch(path); matches != nil {
		id, _ := strconv.Atoi(matches[1])
		sub := matches[2]
		return m.Client.Request(fmt.Sprintf("/person/%d/%s", id, sub), opts)
	}

	// 趋势
	if matches := reTrending.FindStringSubmatch(path); matches != nil {
		return m.Client.GetTrending(matches[1], matches[2], opts)
	}

	// Find
	if matches := reFind.FindStringSubmatch(path); matches != nil {
		externalSource := q.Get("external_source")
		return m.Client.FindByExternalID(matches[1], externalSource, opts)
	}

	// 搜索
	switch path {
	case "/search/movie":
		return m.Client.SearchMovie(q.Get("query"), opts)
	case "/search/tv":
		return m.Client.SearchTV(q.Get("query"), opts)
	case "/search/person":
		return m.Client.SearchPerson(q.Get("query"), opts)
	case "/search/multi":
		return m.Client.SearchMulti(q.Get("query"), opts)
	case "/search/keyword":
		return m.Client.SearchKeyword(q.Get("query"), opts)
	case "/search/collection":
		return m.Client.SearchCollection(q.Get("query"), opts)
	case "/search/company":
		return m.Client.SearchCompany(q.Get("query"), opts)
	}

	// 发现
	switch path {
	case "/discover/movie":
		return m.Client.DiscoverMovie(opts)
	case "/discover/tv":
		return m.Client.DiscoverTV(opts)
	}

	// 类型列表
	switch path {
	case "/genre/movie/list":
		return m.Client.GetGenreMovieList(opts)
	case "/genre/tv/list":
		return m.Client.GetGenreTVList(opts)
	}

	// 电影列表
	switch path {
	case "/movie/now_playing":
		return m.Client.GetNowPlayingMovies(opts)
	case "/movie/popular":
		return m.Client.GetPopularMovies(opts)
	case "/movie/top_rated":
		return m.Client.GetTopRatedMovies(opts)
	case "/movie/upcoming":
		return m.Client.GetUpcomingMovies(opts)
	}

	// 电视剧列表
	switch path {
	case "/tv/airing_today":
		return m.Client.GetAiringTodayTV(opts)
	case "/tv/on_the_air":
		return m.Client.GetOnTheAirTV(opts)
	case "/tv/popular":
		return m.Client.GetPopularTV(opts)
	case "/tv/top_rated":
		return m.Client.GetTopRatedTV(opts)
	}

	// 人物列表
	if path == "/person/popular" {
		return m.Client.GetPopularPeople(opts)
	}

	// 配置
	if path == "/configuration" {
		return m.Client.GetConfiguration(opts)
	}

	// 未匹配的路径，直接透传到 TMDB
	return m.Client.Request(path, opts)
}

// writeJSON 写入原始 JSON 响应
func writeJSON(w http.ResponseWriter, data json.RawMessage) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// writeError 写入错误响应
func writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	resp, _ := json.Marshal(map[string]interface{}{
		"success":        false,
		"status_code":    code,
		"status_message": msg,
	})
	w.Write(resp)
}
