package middleware

import "net/http"

// GzipPublicPages compresses the anonymous public catalogue HTML (the pages
// RenderCache marks publicPageCache): no session cookie on the request, no
// Set-Cookie on the response, 200 only. Those pages carry no CSRF token or
// session secret; the one per-request secret is the CSP nonce, which is fresh
// on every response, so there is nothing stable for a BREACH oracle to guess
// (SDD 13 §7.5). Every other HTML response stays uncompressed.
//
// It must wrap RenderCache: the cache stores and replays the plain body (it
// swaps the nonce in it) and sets the Cache-Control this check reads.
func GzipPublicPages(sessionCookie string, localePrefixes map[string]string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet || !acceptsGzip(r) || hasSession(r, sessionCookie) ||
				!cacheablePage(r, localePrefixes) {
				next.ServeHTTP(w, r)

				return
			}

			gw := &gzipStaticWriter{ResponseWriter: w, html: true}
			defer gw.finish()
			next.ServeHTTP(gw, r)
		})
	}
}
