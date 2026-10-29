# REV-ADOPT-VF-024 (rodada 0)
STATUS: CORRIGIDO_E_APROVADO
Final branch: agent/especialista/ADOPT-VF-024-fix (from agent/senior/ADOPT-VF-024 + 0048d3b, 12b62e1).
Found (fixed, red-green): `localizedRedirects` prefixed every in-site Location under a prefixed request, including targets already prefixed. `web.Ctx.Back` (contact, exports, library, lists controllers) returns the prefixed Referer path, so a POST on /pt-br/... redirected to /pt-br/pt-br/... (404); a switch to /es/... became /pt-br/es/... New cases in TestLocaleCookieRedirect failed with "/pt-br/pt-br/lists/x", "/pt-br/pt-br", "/pt-br/es/games"; fix skips targets whose first segment is a configured prefix -> pass.
Also fixed (my P2.C.1.1 code, separate commit): log sampler matched /pt-BR/ and /en/, real prefixes are /pt-br and /es.
Checked OK: vukem i18n.StripPathPrefix + inertia page url re-localized by vukem (inertia.go:420-424); search/seo use i18n.PathPrefix ("/pt-br", same as the old map); case-sensitive prefix change noted in CHANGELOG.
Re-run: `go build ./...` ok; `golangci-lint run --build-tags integration ./...` 0 issues; `go test -tags integration ./tests/ -run "Redirect|LocalePrefix|Locale"` ok; full `go test -tags integration ./...` exit 0, no FAIL (main's ceb8921 not in this branch; flake did not trigger).
Not re-run: qa/e2e.
