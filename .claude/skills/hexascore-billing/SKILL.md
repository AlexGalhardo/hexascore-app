---
name: hexascore-billing
description: Test Hexascore billing locally (Stripe Checkout/Portal webhooks via stripe listen, AbacatePay PIX simulate-payment) and follow the idempotency rules. Use when touching apps/web/lib/billing, actions/billing.ts, the webhook routes, prices or the billing cron.
---

# Hexascore billing

Reference: `docs/payments.md`. Code: `apps/web/lib/billing/`, `apps/web/actions/billing.ts`, `apps/web/app/api/{webhooks,billing,cron/billing-reminders}/`.

## Rules (never break)

1. Amounts come only from `lib/billing/catalog.ts` (PIX) or the Stripe Price resolved by `lookup_key`. The browser sends ids (`product`, `purpose`, `currency`), never a value.
2. Webhooks read the **raw** body (`await request.text()`) before anything else; verify first (Stripe `constructEventAsync`, AbacatePay `webhookSecret` + `X-Webhook-Signature`, both constant-time).
3. Idempotency = insert into `processed_events` **inside the same transaction** as the effects. Duplicate → 200. Transient failure → 500 (provider retries; nothing was marked).
4. Credits: ledger row with a unique `ref` (Checkout Session id or `pix:<chargeId>`), then — after commit — `mirrorCreditsToCache` (INCRBY only if `hx:credits:<user>` exists). Never `SET` the Redis balance from a webhook.
5. Subscription state from Stripe is re-fetched (`subscriptions.retrieve`), never trusted from the event snapshot. A non-active state of another subscription never overwrites an active row; a PIX payment never overwrites an active Stripe subscription.
6. New action/route/webhook → add a row to the surfaces table in `docs/security.md`.

## Unit tests (no network)

```bash
cd apps/web
bun test lib/billing          # catalog math, Stripe webhook (signed with generateTestHeaderStringAsync), PIX, reminders
```

Tests use an in-memory PGlite with all migrations (`lib/billing/test-db.ts`). Under Bun, Stripe's crypto is async-only: use `generateTestHeaderStringAsync`.

## Stripe locally (test mode only)

Use **this project's** Stripe sandbox key. Never run `stripe listen` / `stripe trigger` with a CLI that is logged into another business's account — pass the key explicitly.

```bash
cd apps/web
# .env.local: STRIPE_SECRET_KEY=sk_test_...  (sandbox of this project)
bun run stripe:setup                                   # Products, Prices (lookup keys) and Portal config; idempotent
stripe listen --api-key "$STRIPE_SECRET_KEY" --forward-to localhost:3000/api/webhooks/stripe
# copy the printed whsec_... into STRIPE_WEBHOOK_SECRET and restart `bun run dev`
```

- Real flow: log in, `/profile/api/checkout` → Card → test card `4242 4242 4242 4242`. The `checkout.session.completed` + `customer.subscription.*` events update `subscriptions`; buy a credit pack to see `credit_ledger` grow by 25,000.
- `stripe trigger checkout.session.completed --api-key ...` creates sessions **without** our metadata, so they are acknowledged and ignored (by design). Use `stripe events resend <evt_id> --api-key ...` on a real event to prove a duplicate is a no-op (response `{"duplicate":true}`, still one ledger row).
- Tampered amount: the action ignores any client value; the price is the Stripe Price of the `lookup_key`.

## AbacatePay (PIX) locally

```bash
# .env.local: ABACATEPAY_API_KEY=<dev key>  ABACATEPAY_WEBHOOK_SECRET=<any long random string>
# Dashboard (dev mode) → Webhooks → URL: https://<tunnel>/api/webhooks/abacatepay?webhookSecret=<same secret>
#   e.g. `cloudflared tunnel --url http://localhost:3000` or `ngrok http 3000`
```

1. `/pt-br/profile/api/checkout` → PIX → "Gerar código PIX" (row in `pix_charges`, status `pending`).
2. Simulate the payment (dev keys only): with `ABACATEPAY_SANDBOX=true` in `.env.local`, click "PAGAR PIX SANDBOX" in the QR modal (`simulatePixPaymentAction`), or by hand:
   ```bash
   curl -X POST "https://api.abacatepay.com/v2/transparents/simulate-payment?id=<provider_ref>" \
     -H "Authorization: Bearer $ABACATEPAY_API_KEY"
   ```
3. The page polls `/api/billing/pix/<id>` every 4 s; even without a tunnel the poll re-checks AbacatePay and settles the charge. With the tunnel, the `transparent.completed` webhook settles it first; a second delivery is a no-op.

## Cron

```bash
curl -H "Authorization: Bearer $CRON_SECRET" http://localhost:3000/api/cron/billing-reminders
```
Emails (pt-BR) PIX subscribers 7 and 1 day before `current_period_end` (once per period, logged in `processed_events`) and marks ended PIX periods `expired`. Without `RESEND_API_KEY` the email is printed to the console in dev.
