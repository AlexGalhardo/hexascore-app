Task: P1.C.2.2 | branch: agent/senior-b/P1.C.2.2 | repo: app | report: .orchestration/reports/P1.C.2.2.md | rodada: 0
Re-run: go test ./tests -run 'TestProfileDetails|TestProfilePassword|TestProfilePrivacy|TestSessionsRevoke|TestSecurity_A07_GoogleOnlyMustSetPasswordBefore2FA'; golangci-lint run --build-tags integration ./...; bun run lint && bun run typecheck && bun run test
