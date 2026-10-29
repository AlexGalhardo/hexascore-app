# Framework engineer handoff (2026-10-29, session 2)

Reason: HANDOFF after 5 VFs plus a flaky-test fix (brief "Limits", context budget).

## Environment

- Worktree `C:\dev\hexascore-app\wt\framework` (leave detached at origin/main when idle). `task` v3.54.0 (`go install github.com/go-task/task/v3/cmd/task@v3.54.0`).
- Site MDX check: PowerShell junction `site\node_modules` -> `C:\dev\hexascore-app\vukem\site\node_modules`, `node scripts/check-mdx.ts <files>`, then `(Get-Item node_modules).Delete()`.
- Ruleset: PR + required checks + strict up-to-date + linear history. Serial merge loop script: `C:\Users\aleex\AppData\Local\Temp\claude\C--dev-hexascore-app\35ef99b5-5e29-468c-b239-25473d83c159\scratchpad\merge-loop.sh <pr>...` (BEHIND -> update-branch, CLEAN -> squash, failing check -> exit 1). Scratchpad may be gone in a new session: it is 20 lines, rewrite if needed.
- `task ci | tail` hides the exit code; redirect to a log and `echo $?` instead. Long runs go to the background (180 s tool limit).
- Python heredocs collapse `\\n`; use the Edit tool for Go strings with `\n`.
- Port 13714 is shared: another agent's SSR server was running there. Never POST `/shutdown` without checking who owns the listener (see "Incident").

## Done this session

| Item | PR | State |
| --- | --- | --- |
| VF-018 queue per-queue limits (`WorkOptions.Limits`, `QUEUE_LIMITS`) | #14 | merged (2b5ed18) |
| VF-019 app metrics docs + `TestProvider_AppMetrics` (no new API) | #15 | merged (ae9276e) |
| VF-025 SSR render cache (`SSRConfig.Cache/CacheTTL/CacheWhen`) | #16 | merged (d9aec2d) |
| VF-038 SSR starter entry, plugin `isSsrBuild` branch, `build:ssr`, `ssr:start` error on immediate exit | #17 | merged (0a03f26) |
| VF-040 | - | status only: duplicate of VF-017/026, fixed by PR #13 |
| session flaky test (all-digit handle) | #18 | merged (4a3bef2) |

Reports: `$ORCH/reports/VF-{018,019,025,038,040}.md`.

## Incident

While checking the VF-038 SSR bundle I POSTed `/shutdown` to 127.0.0.1:13714, which was another process's SSR server (the port was taken; my own bundle failed with EADDRINUSE). That server stopped cleanly. Whoever runs SSR locally (spike/senior) must restart it with `vukem ssr:start` or `serve -with-ssr`.

vukem `main` head: 0a03f26.

## Open

- Release v0.33.0 still BLOCKED (Q-D-006 / FR-001). Do not run the release tool. It now also carries VF-018, 019, 025, 038 (and 040 via #13). After the release: `$ORCH/queue/senior/1-ADOPT-vukem-v0.33.0.md` targeting `C:\dev\hexascore-app\hexascore-spike` (bump, remove workarounds: per-family semaphore -> `QUEUE_LIMITS`; anonymous HTML cache -> `SSRConfig.Cache`; `migrate:locked` and `scripts/concurrent-index.sh` -> built-in lock/`NoTransaction`; SSR plugin branch now upstream), VF statuses `fixed in v0.33.0`.
- Next VFs (all medium, all sizeable): VF-030 (admin hardening), VF-003, VF-008, VF-031, VF-035, VF-037, then lows.
- VF-030 design notes (not started): 2FA state lives in the app (install:auth scaffold), so the panel needs a callback, e.g. `admin.Config{TwoFactor func(ctx) (bool, error), MaxSessionAge time.Duration, AllowCIDRs []netip.Prefix}`; IP from `web.Ctx.IP()` (trusted-proxy aware); session age from the session `_created` key (`session/manager.go`); fail closed (callback error -> 403); `TestSecurity_A01_*`/`A07_*` tests. API design: consider asking the owner first (vukem AGENTS.md "When stuck").
- Docker compose services from session 1 may still run (`docker compose -f .config/compose.yaml stop` in the worktree).
- Do not touch vukem PR #6 (owner's).
