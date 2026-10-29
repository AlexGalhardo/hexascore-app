
Orchestrator: two findings from the automated security review are mandatory fixes on `agent/especialista/P1.H.3.1-fix` (the branch is not merged; Q-H31 also blocks it):
1. `app/billing/billing.go` sync marks another active Stripe subscription of the user `canceled` locally without cancelling it at Stripe, so the user keeps paying for a subscription the app ignores. Fix: cancel the superseded subscription through the Stripe API in the sync (or keep its real status and let the resolver pick one row); refuse Checkout while the user has an open Checkout Session or a live subscription (portal for changes), or use an idempotency key / per-user lock.
2. `app/billing/canceller.go` (deletion hook) trusts the local `state` column. Fix: list the Stripe customer's subscriptions (status=all) and cancel every non-terminal one; if the list call fails, return an error so the purge retries.
Each fix needs a test against the fake Stripe server.
