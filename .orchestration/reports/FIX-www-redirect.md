# FIX-www-redirect Redirect non-canonical hosts to APP_URL
status: concluida
branch: agent/senior/FIX-www-redirect @ 3695568 (from main 1599d40)
files: app/http/middleware/canonical_host.go, app/http/middleware/canonical_host_test.go, bootstrap/app.go, tests/content_pages_test.go, tests/api_reference_test.go
acceptance:
- `go test ./app/http/middleware -run CanonicalHost` -> ok (www -> apex 301 with path+query, railway domain -> apex, api host untouched with and without port/case, /up exempt on any host, /upload not exempt, dev host with port untouched, bad APP_URL errors)
- `go test -tags integration ./tests -run 'TestRobots_Production|TestCanonicalHost_Production|TestAPIReference'` -> ok (production-mode integration: www and railway hosts 301 keeping path+query, /up exempt, api host not redirected; the two existing production tests now request the APP_URL host explicitly, since vukemtest defaults to example.com)
- `golangci-lint run --allow-parallel-runners ./app/http/middleware/... ./bootstrap/...` -> 0 issues
decisions:
- New server middleware `CanonicalHost(appURL, apiHost)`, installed before APIEdge in bootstrap only when `app.Config.IsProduction()` (dev and vukemtest are reached by localhost/test hosts). APP_URL defaults to https://hexascore.app like seoConfig.
- Redirect target is built from the configured scheme+host plus `RequestURI()`, so path and query survive and the request cannot choose the host (gosec G710 nolint with that reason).
- Only exact `/up` is exempt (Railway health check sends its own Host).
risks: any new production-mode test must request the APP_URL host with an absolute URL. POST to a redirected host becomes GET (301); acceptable for browser traffic. API_HOST on a custom port is matched without port.
changelog: Fixed: requests to www.hexascore.app or the Railway domain now 301 to the canonical host, keeping path and query (/up and the API host excluded).
roadmap: none (FIX row)

## Review (especialista, 2026-10-31)

verdict: APROVADO
re-run: go test ./app/http/middleware -run Canonical -> ok; ./tests content/api reference -> ok; golangci 0 issues
checked: CanonicalHost: /up exact-match exempt, API_HOST (port-insensitive) passes, installed only when app.Config.IsProduction(); APP_URL validated at boot; target host from config only
risk: the QA compose stack sets APP_URL=http://localhost:PORT while Lighthouse and e2e default to 127.0.0.1, so they now see one 301 first (batch gate will show)
gate on main after merge: see batch gate in handoff/especialista.md (scripts/ci/local.sh on 127eeeb)

PRONTO_PARA_MERGE agent/senior/FIX-www-redirect (merged 6da4602)
