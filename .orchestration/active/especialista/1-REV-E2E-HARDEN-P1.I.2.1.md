# REV E2E-HARDEN-P1.I.2.1 (rodada: 0)
Task ID: E2E-HARDEN-P1.I.2.1. Branch: agent/senior/E2E-HARDEN-P1.I.2.1. Repo: app. Report: reports/E2E-HARDEN-P1.I.2.1.md.
Re-run: bun run lint && bun run typecheck && bun run test; go build ./...; golangci-lint run --build-tags integration ./...; go test ./app/http/... ./tests -run 'Pow|A05'; e2e: COMPOSE_PROJECT_NAME=hx-<you> with an override moving ports, E2E_BASE_URL + E2E_APP_CONTAINER, `playwright test auth profile api-keys` twice -> 78 passed each.
