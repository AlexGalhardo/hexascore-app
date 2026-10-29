Task P1.H.3.1, branch agent/senior/P1.H.3.1, repo app, report reports/P1.H.3.1.md, rodada: 0
Re-run: go test ./tests -run 'TestStripeWebhook|TestBilling|TestPlan|TestLegacy|TestCancel|TestDeletion'; go test ./...; golangci-lint run --build-tags integration ./...; bun run lint && bun run typecheck && bun run test
Merge blocked until Q-H31 (owner sets Stripe vars on Railway).
