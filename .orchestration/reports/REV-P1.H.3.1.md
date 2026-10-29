# REV-P1.H.3.1 (rodada 0)
VERDICT: CORRIGIDO_E_APROVADO (merge still blocked by Q-H31: Stripe vars on Railway — human gate)
Branch: agent/especialista/P1.H.3.1-fix (from agent/senior/P1.H.3.1, commit d7d182e). Files: app/billing/canceller.go, tests/plan_state_test.go.
Fix: CancelLive (account deletion) selected by local state active/past_due; a row retired locally by a newer subscription (state canceled, Stripe status active) was never cancelled -> deleted account kept being billed. Now filters on Stripe status != canceled/incomplete_expired. TestCancelLiveOnDeletion extended (fails before: "cancels = [sub_1 ...]; want sub_1 and sub_old").
Checked OK: contrib/stripe v0.4.1 approved by ADR 0015 and latest stable; webhook signature (forged 400), event dedupe, re-read via GetSubscription (order-safe), unknown price -> free (fail closed), checkout plan/currency from server catalogue only, linkCustomer never moves another user's customer, .env.example has no secrets, boot check not weakened.
Re-run: go test -count=1 ./tests -run 'TestStripeWebhook|TestBilling|TestPlan|TestLegacy|TestCancel|TestDeletion' -> 14 PASS, ok; -tags integration same -> ok; go test -tags integration ./... -> no failures; golangci-lint 0 issues; bun lint ok, typecheck 0 errors, bun test 50 passed.
Risks/new: double Stripe subscription (two Checkout tabs) -> local flip-flop and double billing -> queue/senior/2-P1.H.3.1b.md (needs owner decision). No rate limit on POST /billing/checkout (Stripe calls per click; low). PlanResolver one query/request (as reported).
PRONTO_PARA_MERGE agent/especialista/P1.H.3.1-fix (after Q-H31)

## Rodada 2 (orchestrator security findings) — CORRIGIDO_E_APROVADO
Branch agent/especialista/P1.H.3.1-fix, commit b51b5f1 (on top of d7d182e). Files: app/billing/{billing,canceller,stripe_calls}.go, tests/{billing,plan_state}_test.go, CHANGELOG.md, vukem-feedback.md.
1. Sync: a second live subscription -> newest by Stripe `created` stays active, the other is cancelled AT STRIPE (immediate DELETE, no proration/refund of the overlap: owner policy if wanted) and its row retired; late webhooks of the loser no longer flip the plan. Checkout first expires the customer's other open subscription Checkout Sessions. Tests: TestStripeWebhookSupersedesSubscription (newer, older-late, loser-late orders), TestBillingCheckoutExpiresOpenSessions.
2. Deletion hook: lists the customer's subscriptions at Stripe (status=all, paginated) and cancels every non-terminal one; list failure or cancel failure -> error (purge retries); no customer -> nothing to cancel. TestCancelLiveOnDeletion covers a Stripe-only subscription and a failed list.
VF-NEW-P1.H.3.1 (contrib/stripe lacks list/expire calls and Subscription.Created; workaround Client.Call).
Re-run: go test -count=1 ./tests -run 'TestStripeWebhook|TestBilling|TestPlan|TestLegacy|TestCancel|TestDeletion' -> 16 PASS; go test -count=1 ./... and -tags integration ./... -> no failures; golangci-lint run --build-tags integration ./... -> 0 issues; markdownlint CHANGELOG/vukem-feedback 0 issues.
PRONTO_PARA_MERGE agent/especialista/P1.H.3.1-fix (after Q-H31)
