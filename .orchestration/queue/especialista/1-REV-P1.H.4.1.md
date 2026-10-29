# REV P1.H.4.1
Task: P1.H.4.1. Branch: agent/senior/P1.H.4.1 (app). Report: .orchestration/reports/P1.H.4.1.md
Re-run: bash scripts/openapi-check.sh ; (cd sdks/ts && bun install --frozen-lockfile && bun run typecheck && bun test) ; bunx tsc --noEmit -p packages/app-client ; golangci-lint run --build-tags integration ./... ; bun run lint && bun run typecheck && bun run test
rodada: 1
