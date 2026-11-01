package middleware

import "net/http"

// Noindex sets X-Robots-Tag: noindex, nofollow on every response of a deployment that must not be
// indexed (staging, previews). Install it with Router.UseServer so it covers every host and /up.
func Noindex(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Robots-Tag", "noindex, nofollow")
		next.ServeHTTP(w, r)
	})
}
