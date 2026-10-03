# Changelog

All notable changes to this project are documented here.
Format: [Keep a Changelog 1.1.0](https://keepachangelog.com/en/1.1.0/). Versioning: [SemVer](https://semver.org/).

## [Unreleased]

### Added

- Billing (0.5): price catalog `lib/billing/catalog.ts` (Pro/Business/credits in BRL/USD/EUR, PIX 12 months for the price of 10, currency by locale).
- Stripe (`stripe@23.0.0`, API `2026-09-30.endive`): `bun run stripe:setup` creates Products/Prices by `lookup_key` and the Customer Portal config; `startStripeCheckout`/`openBillingPortal` actions; `POST /api/webhooks/stripe` (signature, idempotency via `processed_events`, subscriptions and credit packs).
- AbacatePay PIX (API v2): `createPixCharge` action, `GET /api/billing/pix/[id]` polling with provider re-check, `POST /api/webhooks/abacatepay` (URL secret + HMAC, idempotent, period stacking).
- Credits purchase writes `credit_ledger` once per `ref` and bumps the cached Redis balance (`addCachedCredits` in `lib/api/quota.ts`).
- `stripe_customers` table (migration `0002_stripe_customers`).
- Daily cron `/api/cron/billing-reminders`: PIX renewal emails 7d/1d before the end and expiration of ended PIX periods.
- Pages `/pricing` and `/profile/api/checkout` (Base UI radios, PIX QR + copy-paste + live status), `billing` i18n namespace (en, pt-BR, es), "payments not configured" fallbacks.
- Project skill `hexascore-billing` (local webhook testing, idempotency rules).

- `PLAN.md` with decisions and milestones; reference docs in `docs/` (architecture, data model, API, frontend, design system, payments, ingestion, security, testing, process).
- `AGENTS.md` and `CLAUDE.md` agent guides.

### Fixed

- `proxy.ts` matcher skipped every path longer than one character (`.*..*` instead of `.*\..*`), so localized routes 404ed.
- PGlite is kept out of the server bundle (`serverExternalPackages`) so its WASM/extension files load during build and `next start`.

### Removed

- Files inherited from the previous project (RespondeAê): old plan, docs, setup scripts, tests, REST client samples, graphify output, `vercel.json`.
