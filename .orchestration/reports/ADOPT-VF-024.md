# ADOPT-VF-024 report
STATUS: concluida
Branch: agent/senior/ADOPT-VF-024 (pushed). Files: app/http/middleware/{locale_prefix,catalog_redirect}.go (+test), app/seo/seo.go, app/controllers/search_controller.go, routes/web.go, vukem-feedback.md, CHANGELOG.md.
- Removed workaround: LocalePrefix strip/query, RestoreURL, `_locale` query, search_controller localePrefix map. Now MiddlewareConfig{Order: SourcePath, PathPrefixes}, vukem StripPathPrefix, i18n.PathPrefix/LocalizePath (seo, search).
- Kept in app (vukem lacks it): LocaleCookieRedirect = cookie-driven 307 for unprefixed page GETs; plus it prefixes same-site Location of redirects from prefixed requests (so POST redirects keep the locale; CatalogRedirect relies on it).
- Behaviour change: prefix match is case-sensitive (/PT-BR/.. now 404; was case-insensitive). Noted in CHANGELOG.
- `go build ./...` ok; `golangci-lint run --build-tags integration ./...` 0 issues; new unit test TestLocaleCookieRedirect pass; `go test ./tests/ -run 'Redirect|LocalePrefix|Seo|Search|Locale'` ok.
- Full `go test -tags integration ./...`: only FAIL TestAPILicenceFilter, a pre-existing flake (random UUID containing "1234" trips a leak check at tests/api_items_test.go:100); not caused by this change. Suggest fixing the test (separate task).
- E2E (qa/e2e) not re-run; the post-redirect `any` helpers in specs can later be tightened to `t`.

REV (especialista): CORRIGIDO_E_APROVADO, see reports/REV-ADOPT-VF-024.md
PRONTO_PARA_MERGE agent/especialista/ADOPT-VF-024-fix
