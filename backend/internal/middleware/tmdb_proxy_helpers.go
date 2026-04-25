package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"ms_tmdb/pkg/tmdbclient"
)

func parseRequestOptions(r *http.Request) *tmdbclient.RequestOption {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))

	opts := &tmdbclient.RequestOption{
		Context:          r.Context(),
		Language:         q.Get("language"),
		Page:             page,
		Region:           q.Get("region"),
		AppendToResponse: q.Get("append_to_response"),
		ExtraParams:      map[string]string{},
	}

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

func parseIntParam(raw string, name string) (int, error) {
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("参数 %s 解析失败: %w", name, err)
	}
	return value, nil
}

func writeJSONResponse(w http.ResponseWriter, data json.RawMessage) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func writeProxyError(w http.ResponseWriter, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	resp, _ := json.Marshal(map[string]interface{}{
		"success":        false,
		"status_code":    code,
		"status_message": "TMDB 代理请求失败，请稍后重试",
	})
	_, _ = w.Write(resp)
}
