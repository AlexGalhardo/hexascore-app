package middleware

import "net/http"

// VersionHeader carries the deployed release on every response (SDD 11 §6.6).
const VersionHeader = "X-Hexascore-Version"

// Version sets VersionHeader. Install it with Router.UseServer so it also
// covers vukem's /up endpoints, whose body has no version field (VF-027).
func Version(version string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(VersionHeader, version)
			next.ServeHTTP(w, r)
		})
	}
}

// ResolveVersion picks the release name: APP_VERSION (Docker build arg), then
// the commit Railway injects, then "dev".
func ResolveVersion(get func(string) string) string {
	if v := get("APP_VERSION"); v != "" {
		return v
	}
	if sha := get("RAILWAY_GIT_COMMIT_SHA"); len(sha) >= 7 {
		return sha[:7]
	}
	return "dev"
}
