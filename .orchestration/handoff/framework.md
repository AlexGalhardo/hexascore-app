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

## Session 9 (2026-10-29)

- #44 superseded by #57 (VF-001 scaffold rebased onto main after #43 merged; #44 closed with a pointer).
- VF-055 React Email: PR #58 (`feat/vf-055-react-email`). Report `reports/VF-055.md`.
- Next: VF-056 jsonrpc (claimed), then VF-057, VF-011 (notify), VF-014.
- Gate flake: `TestKeyRotateWarnsOnLongPreviousList` Windows TempDir cleanup ("A pasta não está vazia"), passes on re-run.
- Scratch React Email install for real-export checks: scratchpad `re/` (react-email 6.11.6; `npx email export --dir <d> --outDir <o> -s`).
- VF-056 part 1 (server): PR #59 (`feat/vf-056-jsonrpc`). Part 2 = OpenRPC doc + Go client, stacked on #59. TS client generation: optional, likely skip (note in report).
- New queue item from orchestrator: 3-VF-059 (Betterleaks v1.9.0) after VF-057, before 4-VF-011.
- VF-056 done: #59 + #60 (stacked). Next VF-057 (claimed).
- VF-057 part 1: PR #61 (`feat/vf-057-http2-tls`). PR body claims x/net has no production-ready HTTP/3 server: VERIFY before part 2 and correct the PR/docs if wrong.
- Next, in order: VF-057 part 2 (`contrib/http3`, branch from main; may need #61 merged first for WithTLSConfig), 3-VF-059 (Betterleaks v1.9.0, queue file), 4-VF-011 (notify), 4-VF-014 (docs + PageOf/CursorOf[E]).
- Stacked PRs to rebase as NEW branches after their base merges: #60 (on #59).
- VF-057 part 2: PR #62 (stacked on #61). #61 body corrected (x/net check). Next: 3-VF-059 Betterleaks.
- VF-059 part 1: PR #63 (`ci/vf-059-betterleaks`). Part 2 (scaffold + doctor + site page) NOT started; see reports/VF-059.md. Local binary: scratchpad `bl/bin/betterleaks.exe` (v1.9.0, sha256-verified).
- contrib/http3 needs no dependabot change (`/contrib/*` glob).
- Remaining after VF-059 part 2: 4-VF-011 (notify), 4-VF-014. Stacked: #60 on #59, #62 on #61.
- A background `go run github.com/betterleaks/betterleaks@v1.9.0 version` was started only to test feasibility; ignore its output.

## Session 10 (2026-10-29)

- VF-059 part 2: PR #64 (`feat/vf-059-scaffold-secrets`): starter `.betterleaks.toml`, pre-commit scan, `secrets:scan` script, starter CI `secrets` job, new `vukem doctor`, site page `security/secret-scanning`. Report `reports/VF-059.md`. GitHub push protection blocks literal `sk_test_` + 24 alnum placeholders: build such strings at run time in tests.
- VF-052 chain rebased after #46 merged: #48 -> #65 (base main), #49 -> #66 (base `feat/vf-052-cloudflare-visitor-main`), #50 -> #67 (base `feat/vf-052-cloudflare-cache-main`); old PRs closed with pointers. After #65 squash-merges, recreate #66 on main (then #67) the same way.
- Still stacked: #60 on #59, #62 on #61, #66 on #65, #67 on #66.
- Next: 4-VF-011 `notify` (queue/framework, not claimed yet), then 4-VF-014, then low VFs.
- git bash: `git show origin/x:path` needs `MSYS_NO_PATHCONV=1`.
- VF-011 notify: PR #68 (`feat/vf-011-notify`; user ids are string, Postgres checked with a throwaway test; add a drivers/postgres integration test after a core release has `notify`).
- VF-014: PR #69 (`feat/vf-014-envelope-of`, `api.CollectionOf/PageOf/CursorOf`).
- #63 secrets: the `docs/security.md:45` finding seen in a `dir` scan is a history item already in `.betterleaksignore` (6 entries for that file); #63 CI Betterleaks passes.
- Next: remaining low VFs (044 needs design, 049, 050, 004, 012, 020, 023, 033, 034, 060; VF-042 part 2).
- VF-060: PR #70 (totp.HashRecoveryCode). Stopping near context limit; next low VFs: 049, 050, 004, 012, 020, 023, 033, 034, 044 (design), VF-042 part 2.

## Session 11 (2026-10-29)

- VF-049: PR #71 (`feat/vf-049-typed-written`; typed handler that wrote the response skips output; AssertMatchesOpenAPI hints OPENAPI_PREFIX).
- VF-050 + VF-034: PR #72 (`feat/vf-050-openapi-strip`; `Info.StripPrefix`, `OPENAPI_SERVERS`, `OPENAPI_STRIP_PREFIX`).
- VF-004: PR #73 (`docs/vf-004-imaging`; docs + determinism test). WebP encoding = FQ-004 in QUESTIONS.md (owner).
- #61 CI fix pushed (1e075a7): `HTTPServer.Run` used ServeTLS on a caller's tls listener (grpc TLS inside TLS); now TLS only when Run opens the listener. Merged into #62 branch (merge commit, no force). macOS apidocs Chrome "WaitDelay expired" = flake, re-run.
- #60 replaced by #74 (`feat/vf-056-openrpc-main`, on main); #60 closed with pointer.
- Still stacked: #62 on #61, #66 on #65, #67 on #66.
- Next low VFs: 012, 020, 023, 033, 042 part 2; VF-044 design in QUESTIONS.md.
- Later in session 11: VF-012 PR #75 (`SessionGuard.Impersonate`), VF-020 PR #76 (`httpclient.Throttle/Retry/Breaker`), VF-023 PR #77 (site `deployment/logging`, `deployment/releases-and-stability`), VF-042 part 2 PR #78 (`Group.ErrorHandler`). VF-033 wontfix (already on site testing/overview). VF-044: design FD-044 in QUESTIONS.md (owner). VF-004 WebP encode: FQ-004 (owner).
- #61 had merge conflicts with main (so CI did not run): merged origin/main into it (app.go, options_runtime.go, navigation.ts), then #61 into #62 (Taskfile.yml MODULES, go.work). No force-push.
- Low-VF queue now empty except owner-gated FD-044, FQ-004, FD-011/FD-014 done earlier.
- FW-CI-WINDOWS-FLAKES (claimed, active/framework/): PR #79 fixes the pipe-hang root cause (internal/cli timeout, Chrome WaitDelay). Remaining findings and next steps in reports/FW-CI-WINDOWS-FLAKES.md. Then queue 3-VF-044 (owner approved FD-044: opt-in output `required`, `<Name>Input` split, `Info.TypedOnly`, defaults unchanged). FQ-004 answered: VF-004 encoding wontfix (docs-only #73).
- Stopped at context limit (session 11).

## Session 12 (2026-10-29)

- FW-CI-WINDOWS-FLAKES done: #79 (pipe hang), #80 (SSR supervisor tests), #81 (Windows AMX runtime crash = golang/go#81238; CPU log step + docs/maintenance.md "Known CI issues"; drop when a Go release ships CL 828724). Other flakes already fixed on main; details in reports/FW-CI-WINDOWS-FLAKES.md.
- Gate notes: Windows Defender is now disabled by the owner; `internal/testutil` "Acesso negado" and `TestKeyRotate*` TempDir errors are resolved (environment). `task test` runs fully locally.
- Local CPU is AMD Ryzen (non-AMX): golang/go#81238 cannot reproduce here.
- Next: 3-VF-044 (queue/framework), then replacement PRs for #62 (after #61) and #66/#67 (after #65) when the orchestrator says so.
- VF-044: PR #82 (`feat/vf-044-output-required`), report reports/VF-044.md. VF status in vukem-feedback.md: PR open, unreleased.
- Open from session 12: #80, #81, #82 (plus #79). Pending on orchestrator: replacement PRs for #62 (after #61) and #66/#67 (after #65).
- Later in session 12: #77 updated by merging origin/main (nav conflict: kept http2-and-http3 + logging + releases-and-stability). #62 replaced by #87 (`feat/vf-057-contrib-http3-main`, three-dot diff of #62 vs #61 head applied on main; docs/architecture.md conflict resolved; contrib/http3 tests + mod:check ok); #62 closed. #66/#67 still wait for #65.
- DOCS-SITE-REFRESH (active/framework, priority 1): PR #88 = rules + CI check. NEXT: content audit per area (task file list), one PR per area, each running `npm run last-updated` (full history) since the new check requires it; finish with a full refresh. Site checks: junction `site\node_modules` -> `C:\dev\hexascore-app\vukem\site\node_modules`, then check-mdx / eslint / tsc, remove the junction.
- Stopped at context limit (session 12).

## Session 13 (2026-10-30)

- DOCS-SITE-REFRESH item 1 done: #89 (security), #90 (REST), #91 (routing). Report reports/DOCS-SITE-REFRESH.md.
- Item 2 pending: after #88-#91 merge, `git switch -c docs/last-updated-full origin/main`, `cd site && npm run last-updated` (worktree is a full clone), commit, PR.
- Site build via junction: Turbopack cannot resolve `next` through the node_modules junction; use `npx next build --webpack`.
- Touched-only refresh helper: scratchpad lu-touched.mjs (keeps main's json except the branch's changed pages; preserves key order).
- Still waiting: recreate #66/#67 on main after #65 lands.
