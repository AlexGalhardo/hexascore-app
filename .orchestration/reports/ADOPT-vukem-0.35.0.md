# ADOPT-vukem-0.35.0 adopt vukem v0.35.0
status: concluida
branch: agent/senior/ADOPT-vukem-0.35.0 @ ea87205
files: go.mod, go.sum, app/console/predeploy.go, bootstrap/app.go, tests/covers_etag_test.go, docs/sdd/04-*.md (7.4), vukem-feedback.md
acceptance:
- `GOWORK=off go get` core v0.35.0, drivers/postgres v0.6.0, drivers/s3 v0.4.0, contrib/stripe v0.5.0 + `go mod tidy` -> the only vukem modules go.mod requires; proxy returned 200 for all four; no go.work/replace
- `go build ./...` -> ok
- `golangci-lint run --allow-parallel-runners --build-tags integration ./app/console/... ./bootstrap/... ./tests/...` -> 0 issues
- `scripts/openapi-check.sh` -> goldens unchanged (git status clean for openapi/)
- `go test -tags integration ./tests -run 'Predeploy|Admin|Covers|OpenAPI|Moderation'` (Postgres 5433, RustFS 9000) -> ok; new TestCoversRoute_S3ConditionalGetIs304 passes (real ETag, 304 on matching If-None-Match)
validation: lint ok, build ok, tests ok (targeted)
changes: predeployLock removed (migrator polls pg_try_advisory_lock since #123; DropInvalidIndexes + Up kept); apiInfo(apiHost) uses Info.Docs["/app/v1"] instead of os.Args; admin: dedicated catalogue route kept, SDD 04 7.4 notes #121; pgtest nothing; the changelog lists no breaking change
vukem-feedback.md: 4 fixed entries + index rows removed, header line says adopted in v0.35.0; AGENTS.md needs no change (no command changed)
changelog: Changed: bump vukem to v0.35.0 (postgres 0.6.0, s3 0.4.0, stripe 0.5.0); Removed: predeploy try-lock and os.Args OpenAPI hack workarounds
roadmap: no box
risks / follow-ups: predeploy no longer holds its own lock around DropInvalidIndexes (only the migrator's lock covers Up); the invalid-index query already skips in-progress builds. Q-A14 in QUESTIONS.md can be closed. VF-049 (items.go 304 workaround) is not in this release, untouched. Stripe v0.5.0 compiled without changes; stripe tests not run.

## Review (especialista, 2026-10-31)
verdict: CORRIGIDO_E_APROVADO
re-run (main merged in): `go build ./...` ok; `golangci-lint run --build-tags integration ./app/console/... ./bootstrap/...` → 0 issues; `go test -tags integration ./tests -run 'Predeploy|Covers|OpenAPI'` with RustFS up → ok, `TestCoversRoute_S3ConditionalGetIs304` PASS (not skipped), `TestPredeploy_ConcurrentRunsSerialise` PASS; `scripts/openapi-check.sh` → exit 0, `openapi/` unchanged; markdownlint `vukem-feedback.md`, SDD 04 → 0 issues
checked: orchestrator points 1-4, go.mod pins exact (no replace/go.work), secrets none (test uses config.NewSecret with local compose defaults), scope, tests
fixes: 1. predeployLock restored around the whole run (13796ad). The migrator's lock covers only Up, so without ours a second concurrent run could list an INVALID leftover, then (after the first run dropped it and its CREATE INDEX CONCURRENTLY IF NOT EXISTS started a same-named build) DROP INDEX CONCURRENTLY would wait on that build and drop the fresh index. The lock still polls a try-lock, so no snapshot is held; the comment now states that reason instead of the VF workaround.
gate on main after merge (8a73a3c): go build ok
PRONTO_PARA_MERGE agent/especialista/ADOPT-vukem-0.35.0-fix
