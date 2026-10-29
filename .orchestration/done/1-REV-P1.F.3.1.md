Task: P1.F.3.1 | branch: agent/senior-b/P1.F.3.1 | repo: app | report: .orchestration/reports/P1.F.3.1.md | rodada: 0
Re-run: TEST_DATABASE_URL=... go test -tags=integration ./tests -run 'TestNotificationDedupeOnRetry|TestSecurity_A01_NotificationsOwnerOnly'; golangci-lint run --build-tags integration ./...; bun run lint && bun run typecheck && bun run test
