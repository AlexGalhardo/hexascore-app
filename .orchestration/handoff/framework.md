# Framework engineer handoff (2026-10-29, session 7)

Reason: HANDOFF after VF-001 part 1 (large feature, context budget).

## Environment

- Worktree `C:\dev\hexascore-app\wt\framework` (currently on `feat/vf-001-webauthn`; switch to a new branch from `origin/main` for the next work). `task` v3.54.0.
- Gate by hand: `golangci-lint run --allow-parallel-runners ./... && task vet && task mod:check && task test && task api:check`, plus `go run ./internal/tools/snippetcheck` and `npx --yes markdownlint-cli2 <changed md>` for docs. Output to a scratchpad log, `echo "exit=$?"`.
- Site MDX check: PowerShell junction `site\node_modules` -> `C:\dev\hexascore-app\vukem\site\node_modules`, `node scripts/check-mdx.ts <files>`, then `(Get-Item node_modules).Delete()`.
- Windows flake: `task test` may fail `internal/testutil [build failed]` ("Acesso negado"); re-run.
- Cookie sessions are JSON: a `[]byte` put in a session comes back as a base64 string. That is why `webauthn.NewChallenge()` returns a string.
- Python edit scripts: write with the Write tool, run `python <file>`; a failing replace in a multi-file script leaves earlier files already written.

## Merges (FR-003)

Owner answered "merge your own green PRs" in chat, but `gh pr merge 39 --squash` was still denied by the auto-mode classifier this session; noted under FR-003, not retried. Queue for an owner merge, in order: #39, #36, #37, #38, #40, #41, #42, #43. No release or tag until the orchestrator relays an owner request.

## This session

- PR #42 `fix(vukem): stop reporting a finished health check as busy`: `runHealthCheck` sent the result before clearing `running`, so a back-to-back probe got a false 503. Caused the macOS `TestWithHealthCheck` failure on #38 (failed job re-run). Regression test `TestCheckHealthBackToBackNeverBusy`.
- VF-001 part 1, PR #43: stdlib-only `auth/webauthn` (see `$ORCH/reports/VF-001.md` for the app-facing API). Docs: `docs/authentication.md` "Passkeys", `docs/security.md` row (post-1.0 line removed), site `security/passkeys.mdx`, nav, fuzz.yml targets.

## Next

1. VF-001 part 2 (task `$ORCH/active/framework/1-VF-001.md` stays claimed): `install:auth --passkeys` scaffold. Model it on the `--2fa` scaffold in `internal/cli` (look for the 2FA templates and `TestSecurity_A07_TwoFactorScaffoldDefaults`): `passkey_credentials` table (id bytes unique, user_id, public_key, algorithm, sign_count, user_handle, backup flags, transports, name, created_at, last_used_at) and a `users.passkey_handle` column; routes for register options/verify, login options/verify, list and delete (delete needs the current password); challenge in the session, pulled before verifying; throttle like login; a passkey can only be added when the account has a password (recovery path); generated tests for replayed challenge and cross-user credential. Run `go test ./internal/cli -count=1` (~80 s) and the end-to-end project tests. Then set VF-001 `merged in main` when merged.
2. Then the queued task `$ORCH/queue/framework/2-VF-052.md` (Cloudflare `contrib/cloudflare`, high). New deps allowed there only with an exact pinned stable version recorded in the PR and docs.
3. VF-022 triage note: trace/log correlation already exists (`otel.LogHandler` puts `trace_id` on the request log line, `TestProvider_CorrelatesRequestLogs`); the request logger does not log a user id. Remaining real gap: an option to truncate `ip` (/24, /48) in `middleware.Logger`. Answer VF-022 with that and implement only IP truncation.
4. FD-011 / FD-014 still have no owner answer: do not code them.

## Session 8 (2026-10-29)

- VF-001 part 2: PR #44 (`feat/vf-001-passkeys-scaffold`, base `feat/vf-001-webauthn`). Rebase onto main on a NEW branch after #43 squash-merges (no force-push). Report `reports/VF-001.md`.
- ISSUES-19-5: PR #45 (`chore/scorecard-vulns`: x/net v0.62.0, bench echo v5.4.2, maintenance.md Scorecard history + levers). `gh issue comment/close` denied by the classifier: texts for #19/#5 in `reports/ISSUES-19-5.md` for the orchestrator.
- Next: VF-052 (`queue/framework/2-VF-052.md`). Owner rule: after each task open PR, report, end with `STATUS: IDLE`.
- Windows: `internal/testutil` test exe "Acesso negado" now persistent (AV lock), not code.

## Session 8, later (2026-10-29)

- Open PRs from this session: #44 (VF-001 scaffold, stacked on #43), #45 (Scorecard vulns), #46/#48/#49/#50 (VF-052 chain, in that order), #47 (Turnstile action), #51 (VF-054), #52 (VF-053), #53 (VF-022 IP truncation), #54 (VF-039 signed.Token), #55 (VF-042 no-body status).
- After squash merges: rebase stacked work onto main as NEW branches + replacement PRs when the orchestrator says a base landed.
- Remaining open VFs (all low): 004, 007, 012, 020, 023, 033, 034, 042, 044, 049, 050 (skip 011, 014: owner).
- A throwaway container `vukem-pg-test` (postgres:17, 127.0.0.1:55432) may still run: `docker rm -f vukem-pg-test`.

## Session 8, end (2026-10-29): context limit reached

- All open PRs from this session (orchestrator merges): #44 (VF-001 scaffold, stacked on #43: rebase onto main as a NEW branch once #43 lands), #45 (Scorecard vulns; close issue #5 after), #46 -> #48 -> #49 -> #50 (VF-052 chain, rebase each onto main as NEW branches after its base squash-merges), #47 (Turnstile action), #51 (VF-054), #52 (VF-053, changes OpenAPI output: plain routes get 4XX), #53 (VF-022), #54 (VF-039), #55 (VF-042 part 3), #56 (VF-007 sitemap).
- Not started, next in order (owner priority): queue/framework 3-VF-055 (React Email), 3-VF-056 (jsonrpc), 3-VF-057 (HTTP/2 over TLS via http.Protocols; HTTP/3 opt-in contrib/http3), 4-VF-011 (notify per FD-011, approved), 4-VF-014 (FD-014 (a) docs + (b) generic PageOf/CursorOf[E]).
- Then remaining low VFs: 044 (output `required` needs a design: components are shared by input and output types), 049, 050, 004, 012, 020, 023, 033, 034; VF-042 part (2) per-group error handler.
- Gate notes: Windows `internal/testutil` "Acesso negado" is an AV exe lock, not code. Postgres integration: `docker run -d --rm --name vukem-pg-test -e POSTGRES_USER=vukem -e POSTGRES_PASSWORD=vukem -e POSTGRES_DB=vukem_test -p 127.0.0.1:55432:5432 postgres:17`, then `go test -tags integration ./...` in drivers/postgres; remove the container after.
- Python edit scripts: write with Write tool into the scratchpad and run them (heredoc python through bash mangles backslash escapes).
