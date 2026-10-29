# DOC-processed-events
STATUS: concluida
Branch: agent/especialista/DOC-processed-events (pushed, from origin/main). Commits 153dcda (test), 04744e9 (docs).
Files: tests/billing_test.go (+TestSecurity_A08_StripeReplayIgnored: same event id replayed for customer.subscription.updated and invoice.paid -> no extra Stripe read; new event id for same sub -> 1 row), docs/sdd/05-public-api.md (§5 no processed_events + 2 conditions; reminders idempotency key moved to the database cache store), docs/sdd/11-infra-deploy.md §7.6 steps 2/5, docs/sdd/13-security.md (data, threat row, A08 row, checklist), ROADMAP P2.F.2.7 (acceptance rewritten, ticked [x]), CHANGELOG.
- `go test ./... -run TestSecurity_A08_StripeReplayIgnored` -> "ok .../tests" (exit 0); `go test ./tests -run 'TestSecurity_A08_StripeReplayIgnored|TestStripeWebhook' -v` -> 6 PASS.
- `go test -tags=integration ./database/legacyimport -run TestBillingStage` -> "--- PASS: TestBillingStage" (scratch DB, dropped).
- golangci-lint --build-tags integration ./tests/... "0 issues"; markdownlint-cli2 on the 5 md files "0 issues".
Doubt: SDD 05 reminders idempotency (not implemented yet) moved from processed_events to a cache key with TTL past current_period_end: my choice, follows the decision's spirit; orchestrator may reword.
PRONTO_PARA_MERGE (self, needs orchestrator sampling) agent/especialista/DOC-processed-events
