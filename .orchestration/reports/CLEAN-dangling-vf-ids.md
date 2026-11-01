# CLEAN-dangling-vf-ids resolve code comments that cite vukem feedback ids no longer in vukem-feedback.md
status: concluida
branch: agent/senior/CLEAN-dangling-vf-ids @ 6e4bdbb8 (worktree wt/clean-dangling-vf-ids)
files: database/legacyimport/{twofactor.go,recovery_hash_test.go}, internal/ingest/run/continuation.go, config/{queue.go,queue_test.go}, bootstrap/{app.go,queue_limits_test.go}, .env.example, vukem-feedback.md, comments in app/http/middleware/admin_guard.go, routes/schedule.go, tests/queue_order_test.go
acceptance:
- `go test ./config ./app/http/middleware ./database/legacyimport ./internal/ingest/...` -> ok (all admin `TestSecurity_*` in app/http/middleware unchanged and green)
- `TEST_DATABASE_URL=...5433 go test -tags integration ./tests/ingest ./database/legacyimport ./bootstrap -run TestNewDefaultsQueueLimits` -> ok (ingest wiring 65 s, legacy 2FA import incl. TestTwoFactorUsersPassChallenge)
- `golangci-lint run -j 2 --allow-parallel-runners` on config, bootstrap, legacyimport, ingest/run, middleware, routes -> 0 issues; markdownlint: no finding in vukem-feedback.md (6 pre-existing in questions-to-dev.md)
changes:
- VF-060 (twofactor.go): covered, `totp.HashRecoveryCode` is exported in v0.35.0 (<https://vukem.dev/docs/security/two-factor> not re-read; source auth/totp/totp.go). recoveryHashes now calls it; new TestRecoveryHashesMatchVukemFormat (written first, green before and after) checks the hashes against `totp.UseRecoveryCode`.
- VF-018 (continuation.go): covered, `QUEUE_LIMITS=ingest=2` (<https://vukem.dev/docs/digging-deeper/queue> "QUEUE_LIMITS", per process, every driver). The in-process semaphore and `MaxConcurrent` are removed (a waiting job no longer holds a worker either). NFR-06-2 holds even if the operator forgets the variable: `config.WithQueueLimits` (via `vukem.WithEnv` in bootstrap) defaults an unset or empty `QUEUE_LIMITS` to `ingest=2`; a non-empty value replaces it. Tests: config/queue_test.go (default, empty, override) and integration `TestNewDefaultsQueueLimits` through the production `bootstrap.New`. `.env.example` documents it. Caveat: the default beats a `QUEUE_LIMITS` set only in a `.env` file (getenv wrapper cannot see files); set it in the process env.
- VF-021 (schedule.go): not covered. `HourlyAt` exists now but is irrelevant; `cache.Increment` is one store write per call, no buffered counter. New entry VF-NEW-BUFFERED-COUNTER (cache "Counters", scheduler "Frequencies"); comment points at it.
- VF-058 (admin_guard.go): nothing moves, the guard stays. The native options differ in behaviour the tests pin: `AllowCIDRs` answers 403 (app answers 404 to hide the panel), `IdleTimeout` is refreshed only by guarded admin routes (app clock is app-wide for admin and moderator), `TwoFactor` answers 403 (app redirects to the setup page with a flash), and none covers routes beside the panel without `panel.Guard()`. Switching would change answers or loosen/tighten semantics, so new entry VF-NEW-ADMIN-GUARD-SCOPE (admin "Hardening") and the comment points at it. No admin code or test changed.
- tests/queue_order_test.go comment lost its dangling VF-018 mention.
validation: build ok, lint ok, targeted unit and integration tests ok
changelog: Changed: legacy 2FA import uses `totp.HashRecoveryCode`; ingestion cap is vukem `QUEUE_LIMITS` (default `ingest=2`) instead of an in-process semaphore; Added: VF-NEW-ADMIN-GUARD-SCOPE, VF-NEW-BUFFERED-COUNTER
roadmap: no box
risks / follow-ups: production Railway env should not set `QUEUE_LIMITS` to a value without `ingest=` (it would replace the default); `.env.example` was already modified in the main checkout (expect a trivial merge). Orchestrator assigns real VF numbers to the two new entries at merge.

## Review (especialista, 2026-11-01)

verdict: CORRIGIDO_E_APROVADO
re-run: `go test ./config ./database/legacyimport` ok; `go test -tags integration ./bootstrap -run TestNewDefaultsQueueLimits` ok; `go test -tags integration ./tests -run 'TestAdminIngest|TestQueue'` ok (TestAdminIngestRollback, which hung on main after the P6.B batch, now passes in 3 s: the removed semaphore deadlocked the nested rollback, normalize, dq.asserts dispatch of the sync test queue); golangci-lint (`-j 2`, touched packages) 0 issues; `golangci-lint fmt --diff ./...` clean
checked: the ingest cap holds under the queue runner: vukem `QUEUE_LIMITS` is per process for every driver (<https://vukem.dev/docs/digging-deeper/queue> "QUEUE_LIMITS"), `vukem.WithEnv(nil, ...)` keeps the default `os.DirFS(".")` project files (vukem app.go), and `TestNewDefaultsQueueLimits` goes through the production `bootstrap.New`; production sets no `QUEUE_LIMITS`, so `ingest=2` applies. VF-021 and VF-058 kept with reasons; 2FA hash format pinned by `TestRecoveryHashesMatchVukemFormat`
fixes:
1. VF-NEW-ADMIN-GUARD-SCOPE numbered VF-063 and VF-NEW-BUFFERED-COUNTER VF-064 (entries, index and the two code comments; dc92d95d). The four older VF-NEW entries are left for the orchestrator.
2. `app/requests/report.go` was not gofmt'd on main (my P7.C.1.1 merge): fixed in this branch
risk: a non-empty `QUEUE_LIMITS` without `ingest=` drops the cap; documented in `.env.example`
PRONTO_PARA_MERGE agent/especialista/CLEAN-dangling-vf-ids (merged 4d23b04b)
