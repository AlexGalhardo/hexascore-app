---
name: hexascore-local-verification
description: Verify Hexascore changes exactly like CI before calling them done. Use before every commit/PR and before merging dev into main.
---

# Local verification (same gates as CI)

From the repo root:

```bash
bun install
bun run lint:ci
bun run typecheck
bun run test
bun run openapi:check
bun audit --audit-level=high
BETTER_AUTH_SECRET=$(openssl rand -hex 32) CRON_SECRET=$(openssl rand -hex 32) bun run build
cd webscrappings && uv run ruff check && uv run ruff format --check && uv run pytest   # when Python changed
```

- Local DB: `setups/setup-*-docker.*` (Postgres 18) is the reliable option for `next start` and E2E; PGlite is fine for unit tests and `next dev` but breaks with multiple processes or under OneDrive.
- UI changes: run the app and check the page at 360 px and desktop, light and dark, in en/pt-br/es.
- QA layers (E2E, Schemathesis, ZAP, k6): see the `hexascore-qa` skill.
- Then follow `hexascore-release` after merging into `main`.
