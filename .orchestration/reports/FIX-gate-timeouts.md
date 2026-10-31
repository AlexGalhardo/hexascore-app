# FIX-gate-timeouts Gate timeouts under load
status: concluida
branch: agent/senior/FIX-gate-timeouts @ 777f5ce
files: vitest.config.ts, docker-compose.yml
root cause:
- vitest 5.0.5 hardcodes a 60 s worker start timeout (START_TIMEOUT in vitest dist, not configurable); jsdom startup per worker exceeds it when all cores start workers on a busy host. Fix: `maxWorkers: 4`.
- pgtest (vukem, 30 s dropTimeout) runs DROP DATABASE ... WITH (FORCE), which forces a checkpoint; with ~130 parallel tests fsync-bound checkpoints on Docker/Windows exceed 30 s. Fix: compose Postgres `-c fsync=off -c full_page_writes=off -c synchronous_commit=off` (dev/test only).
acceptance:
- before, shared PG default settings, `go test -tags integration -parallel 4 -timeout 30m ./tests/` → FAIL 717 s, 3x "pgtest: drop database ...: context deadline exceeded"
- after, private PG 18.6 with the flags (port 5440), same command → ok 245.6 s
- `bun run test` → 19 files, 66 tests passed, 0 errors
validation: tests ok (vitest, tests pkg); lint/typecheck not needed (config only)
changelog: Fixed: gate reliability on loaded hosts (vitest worker cap, Postgres fsync off in compose)
roadmap: no box. Proposed gate line change in ROADMAP "Resuming work" step 2: add `-parallel 4 -timeout 30m` to the `go test -tags integration ./...` command.
risks / follow-ups:
- Existing shared Postgres volume got the same settings via ALTER SYSTEM (postgresql.auto.conf); recreate with `docker compose up -d postgres` to pick up the compose flags.
- Runs on the shared DB were disturbed by an unrelated Postgres stop (connection refused / shutting down), so the after-number comes from a private container.
- Framework follow-up option: make pgtest dropTimeout configurable (not filed).

## Review (especialista, 2026-10-31)
verdict: APROVADO
re-run: diff is config only (2 files, +5); `docker exec hexascore-postgres-1 psql -tAc "show fsync"` → off (the ALTER SYSTEM on the shared dev volume is in effect, as the report's risks list states); vitest and the integration suite are re-run by the gate on `main` after merge
checked: `fsync=off`, `full_page_writes=off`, `synchronous_commit=off` live only in `docker-compose.yml` (local dev and the ephemeral CI QA stack via `scripts/ci/qa-up.sh`); `Dockerfile` and Railway config untouched, and Railway does not read the compose file; vitest `maxWorkers: 4`; authz/secrets/sql/i18n n/a; scope ok
roadmap: "Resuming work" step 2 gate line updated at merge with `-parallel 4 -timeout 30m`
gate on main after merge: eda3a96 → go build ok, golangci-lint (integration tag) 0 issues, bun lint/typecheck ok, vitest 19 files / 67 tests ok, go test -tags integration -parallel 4 -timeout 30m ./... 29 packages ok
PRONTO_PARA_MERGE agent/senior/FIX-gate-timeouts
