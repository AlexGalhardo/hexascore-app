# senior handoff (2026-10-29)

Done this session (reports in reports/, review tasks queued): P0.F.1.1, P0.F.4.1, P0.F.5.1, P0.F.8.1, P0.F.3.1, P0.F.5.2. P0.F.2.3 was claimed and returned to queue untouched.

Branch state: spike worktree `C:\dev\hexascore-app\wt\spike-senior` is on `agent/senior/P0.F.5.2` (376aeb1, not yet merged into spike main); start each task with `git switch -c agent/senior/<ID> main` (commit or `git checkout --` stray CRLF diffs first, e.g. database/seeders/items_seeder.go). App worktree `C:\dev\hexascore-app\wt\senior` is on `agent/senior/P0.F.5.2` from origin/main; branch new tasks from `origin/main` (no dev branch any more).

Setup facts: Postgres container `hexascore-spike-postgres-1` (user app, password via `docker inspect`, never written to files); export TEST_DATABASE_URL pointing at database `postgres` for pgtest. Spike `.env` (gitignored, copy per worktree) sets PORT=3001, DATABASE_URL on spike_senior, API_HOST=api.localhost. Databases left: spike_senior, spike_senior_test, spike_lock (400k extra items rows). psql is not installed locally: use `docker exec -i hexascore-spike-postgres-1 psql -U app -d <db>`.

Pitfalls learned: python writes CRLF on Windows (use newline='' or gofmt -w); `sed`/`cat` on any `.env` path via Bash is blocked by a deny rule (use Read/Edit); broad `pkill` is denied (kill by PID from netstat); heredoc `python -` hangs when stdin is odd (use `python3 - <<'E'` only after heredoc quoting, worked later); in Git Bash use MSYS_NO_PATHCONV=1 for `-prefix /api/v1`; raw Postgres-only migrations/seeders must guard `s.Dialect() == "postgres"` because scaffold tests run on SQLite; `vukemtest.Options.Env DATABASE_URL` + `pgtest.Database(t, "")` gives a Postgres test app; ORM writes every mapped column, composite keys need QueryOf(...).Where(...).Update; VF numbers: use `VF-NEW-<task ID>` plus an index row.

Pending queue for the next senior: 1-P0.F.2.3 (SSR now on spike main; locale prefix, i18n Once prop, theme cookie), 1-P0.F.7.1, 2-P0.F.6.1.
