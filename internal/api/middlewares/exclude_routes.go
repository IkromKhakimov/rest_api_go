package middlewares

import (
	"net/http"
	"strings"
)

func MiddlewaresExcludePaths(middleware func(handler http.Handler) http.Handler, excludePaths ...string) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, path := range excludePaths {
				if strings.HasPrefix(r.URL.Path, path) {
					next.ServeHTTP(w, r)
					return
				}
			}
			middleware(next).ServeHTTP(w, r)
		})
	}
}
