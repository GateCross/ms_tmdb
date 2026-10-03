package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"ms_tmdb/internal/logic/proxy"
	"ms_tmdb/pkg/tmdbclient"
)

type tmdbExactHandler func(opts *tmdbclient.RequestOption, r *http.Request) (json.RawMessage, error)

type tmdbPatternHandler func(matches []string, opts *tmdbclient.RequestOption, r *http.Request) (json.RawMessage, error)

type tmdbPatternRoute struct {
	pattern *regexp.Regexp
	handler tmdbPatternHandler
}

type tmdbRouteDispatcher struct {
	client       *tmdbclient.Client
	proxyService *proxy.ProxyService
	exacts       map[string]tmdbExactHandler
	patterns     []tmdbPatternRoute
}

type tmdbIDResolver func(id int) int

func newTmdbRouteDispatcher(client *tmdbclient.Client, proxyService *proxy.ProxyService) *tmdbRouteDispatcher {
	return &tmdbRouteDispatcher{
		client:       client,
		proxyService: proxyService,
		exacts:       buildExactHandlers(client),
		patterns: []tmdbPatternRoute{
			{pattern: regexp.MustCompile(`^/movie/(-?\d+)$`), handler: detailRoute(proxyService.GetMovieDetail)},
			{pattern: regexp.MustCompile(`^/movie/(-?\d+)/(.+)$`), handler: passThroughMappedRoute(client, "/movie/%d/%s", map[int]tmdbIDResolver{1: proxyService.ResolveMovieSyncID}, 1, -1)},
			{pattern: regexp.MustCompile(`^/tv/(-?\d+)/season/(\d+)/episode/(\d+)/(.+)$`), handler: passThroughMappedRoute(client, "/tv/%d/season/%d/episode/%d/%s", map[int]tmdbIDResolver{1: proxyService.ResolveTVSyncID}, 1, 2, 3)},
			{pattern: regexp.MustCompile(`^/tv/(-?\d+)/season/(\d+)/episode/(\d+)$`), handler: episodeDetailMappedRoute(client.GetTVEpisode, proxyService.ResolveTVSyncID)},
			{pattern: regexp.MustCompile(`^/tv/(-?\d+)/season/(\d+)/(.+)$`), handler: passThroughMappedRoute(client, "/tv/%d/season/%d/%s", map[int]tmdbIDResolver{1: proxyService.ResolveTVSyncID}, 1, 2)},
			{pattern: regexp.MustCompile(`^/tv/(-?\d+)/season/(\d+)$`), handler: seasonDetailRoute(proxyService.GetTvSeasonDetail)},
			{pattern: regexp.MustCompile(`^/tv/(-?\d+)$`), handler: detailRoute(proxyService.GetTvSeriesDetail)},
			{pattern: regexp.MustCompile(`^/tv/(-?\d+)/(.+)$`), handler: passThroughMappedRoute(client, "/tv/%d/%s", map[int]tmdbIDResolver{1: proxyService.ResolveTVSyncID}, 1, -1)},
			{pattern: regexp.MustCompile(`^/person/(-?\d+)$`), handler: detailRoute(proxyService.GetPersonDetail)},
			{pattern: regexp.MustCompile(`^/person/(-?\d+)/(.+)$`), handler: passThroughRoute(client, "/person/%d/%s", 1)},
			{pattern: regexp.MustCompile(`^/trending/(\w+)/(\w+)$`), handler: trendingRoute(client.GetTrending)},
			{pattern: regexp.MustCompile(`^/find/(.+)$`), handler: findRoute(client.FindByExternalID)},
		},
	}
}

func (d *tmdbRouteDispatcher) dispatch(path string, opts *tmdbclient.RequestOption, r *http.Request) (json.RawMessage, error) {
	data, err := d.dispatchRaw(path, opts, r)
	if err != nil {
		return nil, err
	}
	// 将 /uploads/ 本地图片相对路径改写为完整 URL，供 media-saber 等跨域调用方直接访问。
	if d.proxyService != nil {
		data = d.proxyService.RewriteLocalUploadURLs(data)
	}
	return data, nil
}

func (d *tmdbRouteDispatcher) dispatchRaw(path string, opts *tmdbclient.RequestOption, r *http.Request) (json.RawMessage, error) {
	if handler, ok := d.exacts[path]; ok {
		return handler(opts, r)
	}

	for _, route := range d.patterns {
		if matches := route.pattern.FindStringSubmatch(path); matches != nil {
			return route.handler(matches, opts, r)
		}
	}

	return passThroughPath(d.client, path, opts)
}

func buildExactHandlers(client *tmdbclient.Client) map[string]tmdbExactHandler {
	return map[string]tmdbExactHandler{
		"/search/movie":      searchRoute(client.SearchMovie),
		"/search/tv":         searchRoute(client.SearchTV),
		"/search/person":     searchRoute(client.SearchPerson),
		"/search/multi":      searchRoute(client.SearchMulti),
		"/search/keyword":    searchRoute(client.SearchKeyword),
		"/search/collection": searchRoute(client.SearchCollection),
		"/search/company":    searchRoute(client.SearchCompany),
		"/discover/movie":    optionOnlyRoute(client.DiscoverMovie),
		"/discover/tv":       optionOnlyRoute(client.DiscoverTV),
		"/genre/movie/list":  optionOnlyRoute(client.GetGenreMovieList),
		"/genre/tv/list":     optionOnlyRoute(client.GetGenreTVList),
		"/movie/now_playing": optionOnlyRoute(client.GetNowPlayingMovies),
		"/movie/popular":     optionOnlyRoute(client.GetPopularMovies),
		"/movie/top_rated":   optionOnlyRoute(client.GetTopRatedMovies),
		"/movie/upcoming":    optionOnlyRoute(client.GetUpcomingMovies),
		"/tv/airing_today":   optionOnlyRoute(client.GetAiringTodayTV),
		"/tv/on_the_air":     optionOnlyRoute(client.GetOnTheAirTV),
		"/tv/popular":        optionOnlyRoute(client.GetPopularTV),
		"/tv/top_rated":      optionOnlyRoute(client.GetTopRatedTV),
		"/person/popular":    optionOnlyRoute(client.GetPopularPeople),
		"/configuration":     optionOnlyRoute(client.GetConfiguration),
	}
}

func optionOnlyRoute(handler func(opts *tmdbclient.RequestOption) (json.RawMessage, error)) tmdbExactHandler {
	return func(opts *tmdbclient.RequestOption, _ *http.Request) (json.RawMessage, error) {
		return handler(opts)
	}
}

func searchRoute(handler func(query string, opts *tmdbclient.RequestOption) (json.RawMessage, error)) tmdbExactHandler {
	return func(opts *tmdbclient.RequestOption, r *http.Request) (json.RawMessage, error) {
		return handler(r.URL.Query().Get("query"), opts)
	}
}

func detailRoute(handler func(id int, opts *tmdbclient.RequestOption) (json.RawMessage, error)) tmdbPatternHandler {
	return func(matches []string, opts *tmdbclient.RequestOption, _ *http.Request) (json.RawMessage, error) {
		id, err := parseIntParam(matches[1], "id")
		if err != nil {
			return nil, err
		}
		return handler(id, opts)
	}
}

func seasonDetailRoute(handler func(seriesID, seasonNum int, opts *tmdbclient.RequestOption) (json.RawMessage, error)) tmdbPatternHandler {
	return func(matches []string, opts *tmdbclient.RequestOption, _ *http.Request) (json.RawMessage, error) {
		seriesID, err := parseIntParam(matches[1], "series_id")
		if err != nil {
			return nil, err
		}
		seasonNum, err := parseIntParam(matches[2], "season_number")
		if err != nil {
			return nil, err
		}
		return handler(seriesID, seasonNum, opts)
	}
}

func episodeDetailMappedRoute(
	handler func(seriesID, seasonNum, episodeNum int, opts *tmdbclient.RequestOption) (json.RawMessage, error),
	resolver tmdbIDResolver,
) tmdbPatternHandler {
	return func(matches []string, opts *tmdbclient.RequestOption, _ *http.Request) (json.RawMessage, error) {
		displaySeriesID, err := parseIntParam(matches[1], "series_id")
		if err != nil {
			return nil, err
		}
		seriesID := displaySeriesID
		if resolver != nil {
			seriesID = resolver(seriesID)
		}
		seasonNum, err := parseIntParam(matches[2], "season_number")
		if err != nil {
			return nil, err
		}
		episodeNum, err := parseIntParam(matches[3], "episode_number")
		if err != nil {
			return nil, err
		}
		data, err := handler(seriesID, seasonNum, episodeNum, opts)
		if err != nil {
			return nil, err
		}
		// 回源用的是旧剧集 ID，响应中的 show_id 需改写为对外 ID。
		return proxy.NormalizeOutwardRaw(data, displaySeriesID), nil
	}
}

func passThroughRoute(client *tmdbclient.Client, pattern string, intParamIndexes ...int) tmdbPatternHandler {
	return passThroughMappedRoute(client, pattern, nil, intParamIndexes...)
}

func passThroughMappedRoute(client *tmdbclient.Client, pattern string, resolvers map[int]tmdbIDResolver, intParamIndexes ...int) tmdbPatternHandler {
	rewriteIDMatchIndex := 0
	indexParams := intParamIndexes
	if len(intParamIndexes) > 0 && intParamIndexes[len(intParamIndexes)-1] < 0 {
		rewriteIDMatchIndex = -intParamIndexes[len(intParamIndexes)-1]
		indexParams = intParamIndexes[:len(intParamIndexes)-1]
	}

	intParamSet := map[int]bool{}
	for _, index := range indexParams {
		intParamSet[index] = true
	}

	return func(matches []string, opts *tmdbclient.RequestOption, _ *http.Request) (json.RawMessage, error) {
		values := make([]any, 0, len(matches)-1)
		originalID := 0
		for i := 1; i < len(matches); i++ {
			item := matches[i]
			if intParamSet[i] {
				intValue, err := parseIntParam(item, fmt.Sprintf("match_%d", i))
				if err != nil {
					return nil, err
				}
				if i == rewriteIDMatchIndex {
					originalID = intValue
				}
				if resolver, ok := resolvers[i]; ok && resolver != nil {
					intValue = resolver(intValue)
				}
				values = append(values, intValue)
				continue
			}
			values = append(values, item)
		}
		data, err := passThroughPath(client, fmt.Sprintf(pattern, values...), opts)
		if err != nil {
			return nil, err
		}
		if rewriteIDMatchIndex > 0 && originalID > 0 {
			return rewriteTopLevelID(data, originalID)
		}
		return data, nil
	}
}

func trendingRoute(handler func(mediaType, timeWindow string, opts *tmdbclient.RequestOption) (json.RawMessage, error)) tmdbPatternHandler {
	return func(matches []string, opts *tmdbclient.RequestOption, _ *http.Request) (json.RawMessage, error) {
		return handler(matches[1], matches[2], opts)
	}
}

func findRoute(handler func(externalID, externalSource string, opts *tmdbclient.RequestOption) (json.RawMessage, error)) tmdbPatternHandler {
	return func(matches []string, opts *tmdbclient.RequestOption, r *http.Request) (json.RawMessage, error) {
		externalSource := r.URL.Query().Get("external_source")
		return handler(matches[1], externalSource, opts)
	}
}

func passThroughPath(client *tmdbclient.Client, path string, opts *tmdbclient.RequestOption) (json.RawMessage, error) {
	return client.Request(path, opts)
}

func rewriteTopLevelID(raw json.RawMessage, id int) (json.RawMessage, error) {
	if id <= 0 {
		return raw, nil
	}
	payload := map[string]interface{}{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return raw, nil
	}
	payload["id"] = id
	proxy.NormalizeOutwardIDs(payload, id)
	normalized, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return normalized, nil
}
