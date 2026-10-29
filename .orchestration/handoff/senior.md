# senior handoff (2026-10-29, third session)

Done: P0.F.9.1 (spike b2113ce in worktree wt/spike-senior9; app docs branch 73b910e), P1.A.2.1 (app 905620e), P1.A.4.3 (app c4c7f62). Reports in reports/, REV tasks queued. Earlier four tasks were approved and merged.

Queue left for senior: check queue/senior and reserve/senior (was empty after these).

Pitfalls:
- App worktree `C:\dev\hexascore-app\wt\senior`: branch each task from `origin/main`. Spike worktrees: wt/spike-senior (old), wt/spike-senior9 (P0.F.9.1). RustFS container `hexascore-spike-rustfs-1` (127.0.0.1:9000) and spike Postgres (5432, app/app) are running; real-app Postgres at 127.0.0.1:5433 (hexascore/hexascore).
- Bash tool denies commands that mention `.env` after a pipe (`| tail ...; ls .env`): keep such commands separate.
- Biome: tabs width 4 via .editorconfig; tsc include excludes packages/ (needs @types/node = new dependency, owner yes). golangci-lint 2.15.0 `fmt` (golines) rewrites long lines; nolint directives go on their own line above the statement.
- CI web job now runs `bun run lint` and `bun run test` (vitest); parity test moved from bun:test to vitest.
- Open for orchestrator: setups scripts do not install golangci-lint; Railway pre-deploy still `migrate` (switch to `migrate:locked` is an owner deploy change); covers route in spike resolves the S3 disk lazily (production needs a boot/readiness check).
