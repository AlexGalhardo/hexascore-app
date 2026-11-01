# FIX-nojs-428-page HTML 428 page for no-JS form posts
status: concluida
branch: agent/senior/FIX-nojs-428-page @ 649a28b
files: app/http/middleware/challenge.go, tests/pow_nojs_page_test.go (resources/views untouched: the page is built in the middleware, no layout needed)
acceptance:
- go test ./tests -run 'NoJS428Page|PowHardening|A07' -> ok
- TEST_DATABASE_URL=<local ci pg> go test -tags integration ./tests -run 'NoJS|TestPow|A07' -> ok (existing TestSecurity_A07_NoJSFormPost still sees 428 and no user stored)
validation: go vet ok; golangci-lint not run (per 8.2)
design: a request is a plain form post when its body is urlencoded or multipart, it accepts text/html and carries no X-Inertia header. Only then the 428 is a small self-contained HTML page (no script, no inline style, CSP safe) built from the existing keys auth.challenge.noscript and auth.google, linking GET /auth/google/redirect. JSON clients, Inertia visits and JSON bodies keep the problem+json with a fresh challenge. Status stays 428, nothing is stored (fail closed).
changelog: Fixed: a no-JS user past the proof-of-work threshold now sees an HTML page offering Google sign-in instead of raw 428 JSON.
roadmap: no box
risks / follow-ups: the page is unstyled and not wrapped in the app layout (kept CSP-safe and dependency-free); a designer pass could skin it. Locale comes from the request context, so unprefixed URLs show the default locale.

## Review (especialista, 2026-10-31)
verdict: CORRIGIDO_E_APROVADO
re-run: `go test -tags integration ./tests -run 'NoJS|TestPow|A07'` ok; `go vet` ok
checked: SDD 13 §7.2 (fail closed: still 428, nothing stored), every interpolated value HTML-escaped, no script or inline style (CSP), JSON/Inertia/JSON-body clients keep problem+json (test), `/auth/{provider}/redirect` exists and is in the route inventory, i18n keys exist in en, pt-BR, es
fixes: 1. golangci-lint golines on the Fprintf call (b6cfbf8, branch agent/especialista/FIX-nojs-428-page-fix)
PRONTO_PARA_MERGE agent/especialista/FIX-nojs-428-page-fix (merged 4bf6057 + 8772bf2)
