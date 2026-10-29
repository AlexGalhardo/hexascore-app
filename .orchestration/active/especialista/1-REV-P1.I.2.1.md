Task P1.I.2.1, repo app, branch agent/senior/P1.I.2.1, report .orchestration/reports/P1.I.2.1.md, rodada: 0
Re-run: APP_KEY=... docker compose --profile app up -d --build --wait app; bunx playwright test -c qa/e2e/playwright.config.ts qa/e2e/specs/auth.spec.ts qa/e2e/specs/profile.spec.ts qa/e2e/specs/api-keys.spec.ts --reporter=line; bun run lint && bun run typecheck && bun run test.
