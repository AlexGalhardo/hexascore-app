# VF-NEW-FIX-gate-timeouts pgtest drop timeout configurable
status: concluida
branch: fw: fix/vf-pgtest-drop-timeout @ d88f52b2 (not pushed, no PR; worktree wt/fw-VF-NEW-FIX-gate-timeouts)
files: drivers/postgres/pgtest/pgtest.go, pgtest_test.go, site/content/docs/testing/overview.mdx, site/content/last-updated.json
acceptance:
- `go test -short ./drivers/postgres/pgtest/` → ok (TestDropTimeoutFrom: default 30s, 2m, 90s ok; abc, 0s, -5s, "30" rejected)
- `task ci` → EXIT=0, no FAIL
validation: golangci-lint ok, vet ok, tests ok (-short; integration drop path unchanged, needs Postgres)
changelog: Added - pgtest: PGTEST_DROP_TIMEOUT (Go duration) overrides the 30 s database drop timeout; an invalid value fails the test.
roadmap: no box
risks / follow-ups: env var only, no pgtest.Options (Database(tb, adminURL) signature kept, no API change). Async drop not done. Not run against a live Postgres. App adoption: set PGTEST_DROP_TIMEOUT=2m in the Taskfile/CI of hexascore-vukem after release; the fsync=off workaround may stay.
vukem-feedback status line: VF-NEW-FIX-gate-timeouts -> fixed in vukem branch fix/vf-pgtest-drop-timeout (PGTEST_DROP_TIMEOUT), pending PR/release.

## Review (especialista, 2026-10-31)
verdict: APROVADO
re-run: `task ci` exit 0 (after a golangci lock retry), `task api:check` rc 0; `go test ./drivers/postgres/pgtest/` ok after merging origin/main
checked: env var only, `Database(tb, adminURL)` unchanged; invalid, zero and negative values fail the test (no silent fallback); default 30 s kept; docs page and last-updated refreshed
fixes: branch merged origin/main (after #105) and last-updated.json regenerated (5abdd13a), no code change
PR: #106
merged: PR #106 33/33 green, squash cb6a6f4a
