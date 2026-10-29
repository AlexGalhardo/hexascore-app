# Task ADOPT-VF-024 (priority 3, rodada 0, model: Sonnet, senior pool)

Objective: replace app/http/middleware/locale_prefix.go (+ RestoreURL, catalog_redirect coupling, localeQuery in routes/web.go, controllers/search_controller.go localePrefix map) with i18n.MiddlewareConfig.PathPrefixes, i18n.StripPathPrefix and i18n.LocalizePath. Keep the cookie-driven 307 redirect (SDD 09 7.1) if vukem does not ship it. Touches every web route.

Source: deferred from ADOPT-vukem-v0.33.0 (adoption line in `.orchestration/reports/VF-024.md`). Branch `agent/senior/ADOPT-VF-024` from `origin/main` (vukem v0.33.0 already in go.mod). Remove the `workaround(vukem)` code, keep behaviour, tests green.
Acceptance: `go build ./...`; `golangci-lint run --build-tags integration ./...` 0 issues; `TEST_DATABASE_URL=postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable go test -tags integration ./...` no FAIL; update `vukem-feedback.md` status and `CHANGELOG.md`. Report + `1-REV-ADOPT-VF-024.md`.
