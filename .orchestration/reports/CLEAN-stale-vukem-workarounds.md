# CLEAN-stale-vukem-workarounds remove workarounds vukem v0.35.0 fixed
status: concluida
branch: agent/senior/CLEAN-stale-vukem-workarounds @ 1db43155
files: app/controllers/publicapi/{items,trending}.go, app/controllers/search_controller.go, routes/api_v1.go, app/billing/{stripe_calls.go,stripe_calls_test.go,billing.go}, tests/api_edge_test.go, comments in api_edge.go, request_log.go, exports_controller.go, export_ids.go, eval.go, bootstrap/app.go, vukem-feedback.md
acceptance:
- Baseline first: app/billing/stripe_calls_test.go (new; paging with starting_after, status=all, limit=100; open-session expiry, non-subscription skipped, 400 on expire ignored) green before the refactor and after it
- `go test ./app/billing ./app/controllers/...` -> ok
- `go test -tags integration ./tests -run 'TestAPI|TestAppCatalog|TestAppSearch|TestBilling|TestStripe|TestSecurity_A08_StripeReplay'` (Postgres 5433) -> ok, includes the 304 paths (items, entities, trending, /app/v1 catalogue and suggest) and the unknown-{type} 400s
- `TestAPIEdge` -> ok with 3 new rows (/api/v1/films, /films/{id}, /films/{id}/activity -> 400 invalid-query + code/type), guarding the TypeGuard removal
- `DATABASE_URL=... scripts/openapi-check.sh` -> exit 0, openapi/ unchanged (enum tag renders the same enum as oneof)
- `golangci-lint run --timeout 15m -j 2 --allow-parallel-runners` on billing, controllers, http, jobs, recommend, routes, bootstrap -> 0 issues; markdownlint vukem-feedback.md -> 0 issues
changes:
- VF-049: c.NotModified now ends typed handlers. errNotModified and errSuggestNotModified removed; publicapi.NotModified returns nil after a 304; trending and SuggestRead return the zero value with nil; the web Suggest wrapper checks c.Writer().Written() before writing JSON.
- VF-NEW-P2.H.3.1: Type fields use enum:"games,movies,series,books" (no validation 422); publicapi.TypeGuard and its route Use removed, handlers already answer 400 invalid-query.
- VF-051: customerSubscriptions and expireOpenCheckouts use ListSubscriptions, ListCheckoutSessions, ExpireCheckoutSession; getSubscription removed (Subscription.Created is native in contrib/stripe v0.5.0), billing.go uses Stripe.GetSubscription. Same HTTP calls and parameters.
- Still gaps in v0.35.0, kept, new entries in vukem-feedback.md with docs citations: VF-016/VF-039 -> VF-NEW-PROBLEM-TYPE (WithExtension exists, but type/title cannot be set and vukem-raised problems have no hook); VF-015 -> VF-NEW-HOST-ONLY (r.Host exists, a host cannot be exclusive); VF-022 -> VF-NEW-REQUEST-LINE-FIELDS (docs: request line has no route/user). Comments in api_edge.go, exports_controller.go, request_log.go, bootstrap/app.go cite them.
- VF-029 comments (export_ids.go, eval.go) now cite VF-NEW-EXPORT-WRITER.
validation: build ok, lint ok, targeted tests ok, openapi-check ok
changelog: Changed: typed 304s, enum tag and Stripe list calls use native vukem APIs (no behaviour change); Removed: TypeGuard middleware, VF-049/VF-051 workarounds
roadmap: no box
risks / follow-ups: other workaround comments cite ids that are no longer in vukem-feedback.md and were outside this row: VF-058 (admin_guard.go; vukem admin.Panel has AllowCIDRs, MaxSessionAge, IdleTimeout, TwoFactor, but AdminGuard is app-wide and role-scoped, needs its own check), VF-060 (legacyimport/twofactor.go), VF-018 (ingest/run/continuation.go), VF-021 (routes/schedule.go). Suggest a follow-up row. SDD 05/12 and open-issues.md still cite VF-015/016/022/029 historically, not edited.

## Review (especialista, 2026-11-01)

verdict: CORRIGIDO_E_APROVADO
re-run (main merged in): `go test ./app/billing ./app/controllers/... ./app/http/...` ok; `go test -tags integration ./tests -run 'TestAPI|TestAppCatalog|TestAppSearch|TestBilling|TestStripe|TestSecurity_A08_StripeReplay|TestExportIDs|TestSearch|TestTrending'` ok after the fixes; `scripts/openapi-check.sh` exit 0 (goldens unchanged); `golangci-lint run --build-tags integration -j 2` on the touched packages 0 issues after golines; markdownlint `vukem-feedback.md` clean
checked: vukem v0.35.0 APIs cited in vukem-feedback.md, same Stripe calls and parameters (baseline test), 304 paths, FR-05-4 (unknown type stays a 400 that the quota does not count), scope
fixes:
1. merge with main: P3.A.4.2 (merged after this branch was cut) relied on `TypeGuard` for `getExport`; without it an unknown type answered 422. `ExportInput.Type` takes the `enum` tag and the handler answers 400 `invalid-query` like the catalogue handlers (3cb393cf).
2. the new `TestAPIEdge` row `/api/v1/films/{id}/activity` expected 400, but P3.A.4.2 split activity into four literal routes, so an unknown section is an unknown path: the row now expects 404 `not-found` (3cb393cf).
3. `export_ids.go` comment conflict: kept VF-NEW-EXPORT-WRITER and the Q-A11 popularity line.
follow-up (orchestrator): the stale ids listed in the report (VF-058 admin_guard.go, VF-060 twofactor.go, VF-018 continuation.go, VF-021 schedule.go) need their own row.
PRONTO_PARA_MERGE agent/especialista/CLEAN-stale-vukem-workarounds (merged 614a9442)
