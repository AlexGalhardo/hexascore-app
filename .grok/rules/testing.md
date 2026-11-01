---
paths:
  - "**/*_test.go"
  - "tests/**"
  - "e2e/**"
---

# Testing rules

- Test first. A ROADMAP.md microtask is done only when its acceptance commands pass.
- Go: stdlib + `github.com/google/go-cmp/cmp`, table-driven, `t.Parallel()` when safe, `t.Context()`, no testify.
- Unit tests (`go test ./...`) never touch the network or Docker.
- Integration tests carry `//go:build integration` and read `TEST_DATABASE_URL` (Postgres 18.6 from `docker compose up -d --wait`, port 5433). Skip with a clear message when unset.
- Money, auth and webhooks need an abuse test next to the happy path (see `docs/sdd/13-security.md` and ROADMAP.md).
- TypeScript: `bun run test` (Vitest, Testing Library) and `bun run typecheck`.
- E2E: Playwright in `e2e/`, against the local stack or a built container. Email assertions use Resend test addresses only, never real addresses. Specs skip when the required key is unset.
- Time-dependent code takes a clock; never `time.Sleep` in tests.
- Record the red run before the fix in the report (the reviewer checks it); when a red run is only inferred, say so.
- Two TypeScript runners: Vitest for `resources/js` (`bun run test`), `bun test` for `scripts/**/*.test.ts` that import `bun:test`. A `bun:test` file must be excluded from Vitest (`vitest.config` `exclude`) and wired into `ci.yml` and `scripts/ci/local.sh`, or `bun run test` breaks.
- QA stacks run with `APP_ENV=production`, so `CanonicalHost` redirects any Host but the compose `APP_URL`: use `http://localhost:<port>` (never `127.0.0.1`) for e2e, Lighthouse, Schemathesis and curl. From another container reach the app at `host.docker.internal` and send the original `Host` header.
- Lighthouse is valid only in the `lhci-runner:local` Linux container on this Windows host (Kaspersky injects into host Chrome); budgets LCP 2500 ms, script 170 KB (Q-A20).
- Tool user agents (curl, python-requests, Go-http-client) get 403 from the page scrape guard (ADR 0022); use a browser UA or `/up` for manual checks of production pages.
- Run tests on a loaded host with `-p 1`/`--maxWorkers=1`; a rate-limit test that depends on wall time must stretch its window rather than assume a fast host.
