# Task E2E-HARDEN-P1.I.2.1 (priority 2, rodada 0, model: Sonnet, senior pool)

Source: REV-P1.I.2.1 (especialista). Branch `agent/senior/E2E-HARDEN-P1.I.2.1` from `origin/main` AFTER ADOPT-VF-024-fix and P1.I.2.1 are merged.
1. Signup flake: in a full run of auth+profile+api-keys (78 tests) 2 failed with the page still on /signup after submit (mobile-en profile.spec.ts:18 and mobile-es auth.spec.ts:77); re-run passed. Find the cause from the failure screenshot/trace (hydration-lost password value, PoW, rate limit) and fix the root cause in the app or helper; no blind retries.
2. Replace `any` with `t` in post-redirect asserts now that redirects keep the locale prefix (ADOPT-VF-024-fix).
3. Drop global `bypassCSP: true`: make SecureHeaders omit `upgrade-insecure-requests` when APP_URL is http (local container), keep it in production; add a Go test for both cases.
Acceptance: two consecutive full runs of the three specs -> 78 passed each; `bun run lint && bun run typecheck && bun run test`; Go: `go build ./...`, `golangci-lint run --build-tags integration ./...` 0 issues, related tests pass. Use an isolated compose project (COMPOSE_PROJECT_NAME=hx-senior, override ports). Report + 1-REV-E2E-HARDEN-P1.I.2.1.md.
