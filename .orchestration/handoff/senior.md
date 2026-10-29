# senior handoff (2026-10-29, second session)

Done this session (reports in reports/, REV tasks queued): P0.F.2.3, P0.F.7.1, P0.F.10.1, P0.F.7.2. All four are awaiting especialista review; none merged by me.

Branch state:
- Spike worktree `C:\dev\hexascore-app\wt\spike-senior` is on `agent/senior/P0.F.7.2` (c2feafe, from spike main 5383b05). Other spike branches (local commits, unpushed, spike is local only): `agent/senior/P0.F.2.3` (88c1def, already merged into spike main by orchestrator), `agent/senior/P0.F.7.1` (merged), `agent/senior/P0.F.10.1` (f70cb51: lighthouserc.json, @lhci/cli 0.15.1, middleware.Gzip, .gitignore `.lighthouseci/`; branched from an older main, not yet merged). Start new tasks with `git switch -c agent/senior/<ID> main`.
- App worktree `C:\dev\hexascore-app\wt\senior` is on `agent/senior/P0.F.7.2`; each task branch there was cut from origin/main and holds only docs/spike-report.md + vukem-feedback.md edits (VF-NEW-<ID> rows). Branch new tasks from `origin/main`.

Queue left: 1-P0.F.9.1, 2-P1.A.2.1, 2-P1.A.4.3 in queue/senior; reserve/senior has 2-P0.F.6.1 (ROADMAP already shows it [x], check before doing).

Setup facts and pitfalls (new this session):
- Port 3001 is held by a foreign process (not mine, kill denied); run the spike on `PORT=3011` (env overrides .env). Node SSR for `serve` needs `app ssr:start` in a second process; `SSR_SUPERVISE` only works in the container/production.
- `resources/public/hot` makes `vite` tags point to the dev server: move it away when running a built server.
- Postgres tests: export `TEST_DATABASE_URL=postgres://app:<pw from docker inspect>@localhost:5432/postgres?sslmode=disable` in the same shell command.
- Host runs Kaspersky web shield: it injects scripts into pages in Chrome; Lighthouse runs need `blockedUrlPatterns: ["*kaspersky-labs.com*"]` (already in lighthouserc.json) and `CHROME_PATH="C:/Program Files/Google/Chrome/Application/chrome.exe"`. lhci autorun takes ~80 s and gets backgrounded by the tool after 180 s.
- python on Windows reads files as cp1252 by default: always pass `encoding='utf-8'` (and newline='') when editing files with non-ASCII text; `§` in Go source broke the build once. `uv` is not installed; the Python SDK generator ran from a venv in the session scratchpad.
- vukem facts: group `Use` middleware cannot change handler errors (use `r.Services().Errors`, see app/http/problems); custom view funcs are not request-bound (pass data via Inertia shared props, read `.Page.Props.x` in app.html); i18n locale from URL prefix via query injection (`_locale`) + RestoreURL; `api.Op{Status:202}` with `struct{}` writes `{}`.
- Open decision for the orchestrator: legacy SDK operation names (getGame per type) vs SDD getItem; see reports/P0.F.7.2.md.
