# QA-e2e-main Playwright e2e on a container built from main
status: parcial (run incomplete: stopped by estagiario after the run was still writing; see risks)
branch: none (worktree `C:\dev\hexascore-app\wt\estagiario-QA-e2e-main`, detached at main c9f03cc; `.env` copied in, not printed; `bun install --frozen-lockfile` ok)
files: .orchestration/reports/QA-e2e-main.md
acceptance:
- `docker compose --profile app build app-init` with `E2E_IMAGE_TAG=main` → `Image hexascore-app:main Built`, exit 0
- `E2E_IMAGE_TAG=main docker compose --profile app up -d --wait` → app-1 Healthy, `curl http://127.0.0.1:3000/up` → 200
- `bun run e2e` run 2 (log `scratchpad/qa-e2e-run2.log`, no timeout): 459 ✓, 33 ✘ over 496 log lines when stopped (projects desktop-en, desktop-pt-BR, desktop-es, mobile-en, mobile-pt-BR). Run was still writing; the total is not known. Run 1 was killed by my own 580 s timeout (exit 143) and its counts are not used.
- failures are reported by the list reporter only as `✘` lines; the error text is in `qa/e2e/test-results/*/error-context.md` (all named flows: `Error: page.waitForURL: Test timeout ... waiting for navigation until "load"`).

## The four flows named by the orchestrator (main @ c9f03cc)
| flow | desktop-en | desktop-pt-BR | desktop-es | mobile-en | mobile-pt-BR |
| --- | --- | --- | --- | --- | --- |
| lists: create, fill, reorder, publish, hide and delete | ✘ 4.1m timeout (waitForURL, 240 s) | ✓ 26.1s | ✘ 4.3m | ✓ 10.3s | ✓ 18.6s |
| profile: change password, then log in | ✓ 11.8s | ✓ 12.7s | ✓ 39.9s | ✓ 8.3s | not reached |
| profile: account deletion (countdown, cancel, confirm, undo) | ✓ 23.5s | ✘ 2.1m (waitForURL, 120 s) | ✓ 51.8s | ✘ 2.1m (waitForURL) | not reached |
| ratings: item status set and cleared | ✘ 2.0m (waitForURL, 120 s) | ✘ 2.0m | not reached | not reached | not reached |
| ratings: rate in six categories by keyboard, edit, delete (extra) | ✓ 14.8s | ✘ 19.4s | ✘ 2.1m | ✓ 19.7s | not reached |

Decisive lines (all from `qa/e2e/test-results/*/error-context.md`, line 19 and 21):
- `lists-...-desktop-en`: `Error: page.waitForURL: Test timeout of 240000ms exceeded.` / `waiting for navigation until "load"`
- `profile-...-desktop-pt-BR` (account deletion): `Error: page.waitForURL: Test timeout of 120000ms exceeded.` / `waiting for navigation until "load"`
- `ratings-...-item-status-can-be-set-and-cleared-desktop-en`: `Error: page.waitForURL: Test timeout of 120000ms exceeded.` / `waiting for navigation until "load"`

Other ✘ in the same run (not in the four flows, for the owner's triage): `i18n.spec.ts:4` in all 5 projects; `auth.spec.ts:106` password reset (desktop-pt-BR, desktop-es, mobile-en); `notifications.spec.ts:9` and `:37`; `social.spec.ts:11`; `time-to-beat.spec.ts:11`; `admin.spec.ts:16`; `keyboard.spec.ts:59` (mobile); `ssr-nojs.spec.ts:30` `/login` (run 1 only, not in run 2).

## Answer to "do the four flows fail on main too?"
Yes, on the `main` build at c9f03cc, but not uniformly. Each of the four flows fails in at least one project and passes in another, so they look flaky under load (the host was running several other stacks and Playwright workers at the time), not a deterministic regression. Confirmed by the same `waitForURL` timeout signature as the P2.H.1.1 run. No single-flow rerun was done (no time, context full), so "alone" was not tested.

## Stack and host state
- stopped with `docker compose --profile app stop app app-init` (not `down`). `hexascore-app-1` and `hexascore-app-init-1` Stopped. Port 3000 free at 08:3x UTC (`docker ps` shows no `3000->`). Postgres (`hexascore-postgres-1`) and RustFS still up.
- The e2e background task was stopped with TaskStop before the stack was stopped.

validation: e2e partial as above; lint/typecheck n/a
changelog: none
roadmap: no box
risks / follow-ups: (1) the run is incomplete; a full rerun on a free host is needed for totals. (2) The four flows are flaky on main, not only on P2.H.1.1; the specialist should decide whether they are a test-timing issue (`waitForURL` `load` with 2 min waits) or a real navigation bug. (3) Orchestrator said the run had ended, but log lines kept growing (465 to 490 in 20 s at 08:29); the heartbeat check was based on a `tasklist` that matched unrelated node processes.
