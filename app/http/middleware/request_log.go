package middleware

import (
	"net/http"

	"github.com/AlexGalhardo/hexascore-app/app/logging"
)

// RequestLog gives the request a logging.Fields holder. It is a server
// middleware registered before vukem's Logger, so the request line can read
// the route and user that inner middleware record (workaround(vukem): VF-022).
func RequestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(logging.WithFields(r.Context())))
	})
}

// RequestRoute records the matched route pattern for the request line; it is
// a router middleware, so it runs only for matched routes.
func RequestRoute(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logging.SetRoute(r.Context(), r.Pattern)
		next.ServeHTTP(w, r)
	})
}
