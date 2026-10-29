# ADR 0015: Monetisation: paid read API via Stripe in v1, PIX in the second release

Status: accepted (2026-10-28). Decisions: brainstorming D-06, D-14 (clarified in Round 4).

## Context

Legacy revenue model: a read-only public API with plans free (100 lifetime requests), pro (50k/month, R$29 / US$6 / €6) and business (500k/month, R$149 / US$29 / €27), credit packs, Stripe Checkout and Customer Portal in BRL/USD/EUR, and AbacatePay PIX behind a feature flag. Prices and limits in `legacy/apps/web/lib/billing/catalog.ts` and `lib/api/plans.ts`. vukem ships `contrib/stripe` (API and webhooks, stdlib) and an `install:billing` scaffold for subscriptions (`https://vukem.dev/docs/digging-deeper/billing`).

## Decision

- v1: the site and the mobile app are free, without ads. Revenue comes from the paid API with the legacy plans and prices (reviewed by the owner before launch), billed through Stripe subscriptions on vukem's `contrib/stripe` and `install:billing`. Webhook idempotency, plan changes, past-due handling and usage metering are specified in SDD 05.
- Second release: AbacatePay PIX (one-off payments for 1 or 12 months and credit packs) rebuilt in Go behind the same billing abstraction; credit packs return with it.
- Not planned: consumer subscriptions, ads.

## Alternatives considered

- **Stripe + PIX in v1.** More scope on the critical path for a flow that was behind a flag. Rejected by the owner (Round 4).
- **Consumer subscription.** Needs premium features defined now; none identified. Deferred.
- **Ads.** Hurts Core Web Vitals and trust; competitors are criticised for it. Rejected.

## Consequences

- Positive: smallest billing surface for cutover; vukem's Stripe module gets a real consumer.
- Negative: Brazilian developers without international cards wait for PIX.
- Obligations: SDD 05 defines plans, quotas, metering, Stripe flows and tests; the pricing page states the data coverage of the API (ADR 0006).
