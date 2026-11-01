---
paths:
  - "app/billing/**"
  - "app/payments/**"
---

# Money and payments rules

- Money is `int64` cents (`*_cents` columns, `*Cents` JSON). Never floats. Currencies: BRL, USD, EUR; prices are per currency.
- Plans are Stripe prices `pro` and `business` (D-14). PIX and AbacatePay are deferred to the second release (D-06, D-14): do not add them.
- Agents use Stripe test mode only: `STRIPE_SECRET_KEY` must start with `sk_test_`. Never read or write live keys.
- The server prices everything: checkout ignores any client price, currency or plan amount.
- Every money event (charge, refund, subscription change) writes one append-only ledger row in the same transaction as the state change. The ledger table is not in an SDD yet: flag it before coding.
- Webhooks: verify the signature first, then store the provider event id (unique) and process idempotently; mark processed only after success so retries work.
- Outbound provider calls carry deterministic idempotency keys derived from local records (e.g. `checkout-<subscriptionID>`).
- State transitions are conditional updates (`UPDATE ... WHERE status = 'pending'`) and are covered by abuse tests (replay, double activation, price tampering).
