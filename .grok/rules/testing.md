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
