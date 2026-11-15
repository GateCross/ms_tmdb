package proxy

import (
	"errors"
	"net/http"
	"testing"

	"ms_tmdb/pkg/tmdbclient"
)

func TestShouldUseFallbackCache(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "network error can use cache",
			err:  errors.New("dial tcp timeout"),
			want: true,
		},
		{
			name: "tmdb 404 resource deleted can use cache",
			err:  &tmdbclient.APIError{StatusCode: http.StatusNotFound, Body: `{}`},
			want: true,
		},
		{
			name: "other tmdb http error must be surfaced",
			err:  &tmdbclient.APIError{StatusCode: http.StatusTooManyRequests, Body: `{}`},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldUseFallbackCache(tt.err); got != tt.want {
				t.Fatalf("shouldUseFallbackCache() = %v, want %v", got, tt.want)
			}
		})
	}
}
