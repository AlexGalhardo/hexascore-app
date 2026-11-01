package middleware

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// CanonicalHost answers 301 to appURL (path and query kept) for any request whose Host is neither
// the appURL host nor apiHost. /up is exempt so platform health checks, which send their own Host,
// keep working. apiHost is compared without port, appURL host as configured (dev ports).
func CanonicalHost(appURL, apiHost string) (func(http.Handler) http.Handler, error) {
	u, err := url.Parse(strings.TrimRight(appURL, "/"))
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return nil, fmt.Errorf("canonical host: APP_URL %q needs a scheme and host", appURL)
	}

	canonical := strings.ToLower(u.Host)
	apiHost = strings.ToLower(apiHost)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := strings.ToLower(r.Host)
			if r.URL.Path == "/up" || h == canonical || hostOnly(h) == apiHost {
				next.ServeHTTP(w, r)
				return
			}

			// The target host is the configured APP_URL; the request only supplies path and query.
			target := u.Scheme + "://" + u.Host + r.URL.RequestURI()
			http.Redirect(w, r, target, http.StatusMovedPermanently) //nolint:gosec // G710: host is config
		})
	}, nil
}
