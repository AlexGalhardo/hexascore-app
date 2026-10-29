# Software design documents (SDD)

One SDD per domain. They are the contract the roadmap tasks implement. Decisions come from [../brainstorming.md](../brainstorming.md) (D-xx, A-xx) and are justified in [../adrs/](../adrs/); an SDD never re-decides, it specifies.

## Index

| # | File | Domain | Main decisions |
| --- | --- | --- | --- |
| 01 | [01-catalog.md](01-catalog.md) | Items, works/editions, entities (people, organisations, franchises, platforms, genres), translations, slugs, SEO data | D-01, D-10, D-17, A-06 |
| 02 | [02-scoring.md](02-scoring.md) | Ratings input, aggregate formula, confidence, anti-manipulation, status, time-to-beat | D-00, D-13 |
| 03 | [03-users-auth.md](03-users-auth.md) | Accounts, sessions, 2FA, Google, access tokens, roles, bans, deletion, migration from Better Auth | D-05, D-06, A-01, A-04 |
| 04 | [04-community-edits.md](04-community-edits.md) | Edit suggestions, states, moderation queue, audit trail, anti-abuse, revert | requirement 5, D-15 |
| 05 | [05-public-api.md](05-public-api.md) | Read API v1, keys, plans, quotas, billing (Stripe), OpenAPI, SDKs, exports, changes feed, webhooks; first-party app API | D-03, D-14, D-20, A-01 |
| 06 | [06-ingestion.md](06-ingestion.md) | Sources and licences, connectors in Go, jobs, normalisation, covers, quality asserts, backfills | D-03, D-09, D-11, A-05 |
| 07 | [07-search.md](07-search.md) | Postgres FTS + trigram, suggest, per-locale config, ranking, interface boundary | D-16 |
| 08 | [08-recommendation.md](08-recommendation.md) | Trending primitives and top-10 lists, similar items, content-based recommendations | D-02, A-02 |
| 09 | [09-web.md](09-web.md) | Inertia + React app, SSR, routing/i18n, pages, components, design tokens, Core Web Vitals budgets, accessibility | D-04, D-07, D-17 |
| 10 | [10-mobile.md](10-mobile.md) | React Native (Android first), shared contract and tokens, offline behaviour, release to Google Play | D-12, A-01 |
| 11 | [11-infra-deploy.md](11-infra-deploy.md) | Railway topology, containers, env/secrets, migrations, backups, releases, cutover, rollback | D-08, D-18, D-20, D-21, A-03 |
| 12 | [12-observability.md](12-observability.md) | Logs, metrics, traces, alerts (SSR fallback, job failures), health checks, cost monitoring | D-07, D-08 |
| 13 | [13-security.md](13-security.md) | OWASP, rate limiting, CSRF/CSP, secrets, LGPD, abuse, admin hardening, dependency policy | requirement 7, vukem security defaults |
| 14 | [14-user-library.md](14-user-library.md) | Lists, statuses and time-to-beat views, public profiles, reports, contact, notifications inbox, data export | D-06, D-12, A-07 |

Triage of every SDD's open issues: [open-issues.md](open-issues.md).

## Requirement coverage

Requirements from `PROMPT.md` mapped to documents. Filled by the architect when the set is complete.

| Requirement | SDD | ADR |
| --- | --- | --- |
| 1. Migration to vukem, native tools first | 01–14 (all), chiefly 09, 11 | 0001, 0004, 0008, 0011 |
| 2. Railway instead of Vercel | 11, 12 | 0002, 0014 |
| 3. Mobile app (Android first) sharing API, logic and visual style | 10, 05, 09, 03 | 0009, 0008 |
| 4. Postgres main; Redis/Mongo only with an ADR | 01, 07, 11 | 0003, 0011 |
| 5. Edit suggestions, 3/day, moderation flow, states, audit trail, anti-abuse, revert | 04, 14 (notifications), 12 (SLA alerts) | 0010, 0016 |
| 6. `vukem-feedback.md` maintenance (format: id, date, severity, context, reproduction, workaround, status) | root file `vukem-feedback.md`; every SDD's open issues | 0001 |
| 7. Continuous ingestion, quality asserts, missing-field detection, backups, rollbacks, safe migrations, workflows, releases, licences | 06, 11, 12, 13 | 0006, 0007, 0013, 0014 |
| Scoring more reliable than competitors (product goal 3) | 02, 08 | 0005 |
| Public API DX (product goal 2) | 05 | 0006, 0009, 0015 |
| Security (OWASP, rate limiting, LGPD, secrets) | 13, 03 | 0016, 0008 |
| Agents tooling (skills in AGENTS.md/CLAUDE.md) | root `AGENTS.md`, `CLAUDE.md` | — |
| Git: `legacy` branch, no history rewrite; releases like vukem | 11 | 0013 |

Coverage check (2026-10-28): every numbered requirement maps to at least one SDD and one ADR. SDD 13 and 14 were added beyond the prompt's list (security was in the list; "user library" splits lists, profiles, reports and notifications out of users/auth).

## Cross-SDD conventions (binding; fixes inconsistencies found in the first wave)

- **Access-token abilities** (SDD 03, 05, 08, 10): `catalog:read`, `ratings:write`, `status:write`, `lists:write`, `profile:write`, `suggestions:write`. No `app:read`/`app:write` umbrella abilities.
- **Problem codes** (RFC 9457 `code` member): kebab-case (`quota-exceeded`, `rating-before-release`).
- **JSON field names** in HexaScore payloads: camelCase. Envelope keys that vukem emits (`next_cursor`, `request_id`) stay as vukem emits them unless configurable; the gap is filed as VF-014.
- **Scales**: user-facing and API scores are 0–10 with one decimal; storage is 0–100 integers.
- **Identifiers**: UUID v7 primary keys for catalogue, social and billing rows (as legacy); `users.id` stays `text` (legacy Better Auth ids are preserved at migration so no foreign key is remapped; new accounts get a ULID, SDD 03 §5) and every user foreign key is `text`. Public API and URLs use the UUID for entities and `type/slug` for items.
- **Timestamps**: RFC 3339 UTC in JSON; `timestamptz` in Postgres.
- **Locales**: `en`, `pt-BR`, `es`; URL prefixes `/`, `/pt-br`, `/es`.
- **Route prefixes**: paid API `/api/v1` (served as `api.hexascore.app/v1` by host routing); first-party app API `/app/v1`; admin `/admin`; health `/up`.

## How to write an SDD (for the authoring agents)

Audience: Claude Opus 5.5 / Sonnet 5.5 agents that will implement the roadmap without this conversation. Write for them: explicit, self-contained, no ambiguity.

Rules:

- English. Markdown, ATX headings, dash lists, fenced code with a language. Keep lines reasonably short; tables for structured facts.
- Facts about vukem come from the public documentation site **<https://vukem.dev/docs>** (owner rule, 2026-10-28): cite the page URL (for example `https://vukem.dev/docs/frontend/inertia`) and the heading. The local checkout `C:\dev\hexascore-app\vukem\docs\*.md` is only a way to find the right page; when the site has no page for a fact, cite the local file and tag it `[SITE PAGE MISSING]`. Never invent a vukem API, option, command or flag: if the docs do not show it, write `[NOT IN VUKEM DOCS]` and describe the gap as a `vukem-feedback.md` candidate.
- Page map (site sections): getting started `/docs/introduction`, `/docs/installation`, `/docs/configuration`, `/docs/directory-structure`, `/docs/starters`; `basics/` routing, controllers, http-context, middleware, requests, responses, validation, sessions, views, csrf, error-handling, static-files; `rest-apis/` overview, typed-handlers, openapi, pagination, idempotency, client-generation, tutorial, grpc, graphql; `database/` overview, orm, relations, migrations, factories-seeders, redis, mongodb; `security/` overview, authentication, two-factor, social-login, api-tokens, jwt, authorization, rate-limiting, security-headers, encryption, signed-urls, building-blocks; `frontend/` inertia, inertia-i18n, htmx, vite; `digging-deeper/` admin, billing, cache, queue, scheduler, events, mail, storage, uploads, exports, feature-flags, localization, opentelemetry, telegram-logging, webhooks, sse, broadcasting, devtools; `deployment/` docker-and-railway, single-binary; `cli/` overview, reference, reference-new, reference-install, reference-make, reference-database, reference-runtime; `testing/overview`; `architecture/lifecycle`. Not on the site: tooling, maintenance, stability, performance, PLAN and brainstorming (local files only).
- Facts about the legacy app come from `C:\dev\hexascore-app\legacy\` (cite paths). The legacy behaviour is the parity baseline unless a decision in `../brainstorming.md` changes it.
- Decisions are referenced by ID (D-07, A-01). Do not re-argue them; if the SDD finds a decision impossible as stated, add an "Open issues" entry with the smallest change that makes it possible.
- Everything measurable gets a number: limits, sizes, latencies, retention, thresholds, quotas.
- Every requirement line must be testable: say how it is verified (unit test, integration test with Postgres, Playwright/axe, k6, Schemathesis, ZAP, manual checklist).
- No implementation code. Short Go type sketches, SQL DDL sketches and JSON examples are allowed where they remove ambiguity.
- Length target: 1,500–3,000 words. Dense beats long.

Template (use these headings in this order):

```markdown
# SDD NN: <Domain>

Status: draft | reviewed. Decisions: D-xx, A-xx. ADRs: NNNN. Related SDDs: NN, NN.

## 1. Objective
One paragraph: what this domain does for the user and for the product goals (UX speed/clarity/discovery, DX, score reliability).

## 2. Scope
In scope / out of scope (v1 vs later), with the decision IDs that set the line.

## 3. Functional requirements
FR-NN-1, FR-NN-2, …: one sentence each, testable, with the verification method.

## 4. Non-functional requirements
NFR-NN-1, …: performance, availability, security, privacy, accessibility, i18n, cost, with numbers.

## 5. Data model
Tables/columns/constraints/indexes (SQL sketch), enums, invariants, retention, migration notes from the legacy schema (`legacy/apps/web/lib/db/schema/*.ts`).

## 6. Contracts
HTTP routes (method, path, auth, input, output, errors as RFC 9457), events/jobs, Inertia pages and props where relevant, OpenAPI operation IDs, SDK names.

## 7. Behaviour and algorithms
Flows, state machines, formulas, edge cases, concurrency, idempotency, failure handling.

## 8. Edge cases
Explicit list with expected behaviour.

## 9. Acceptance criteria
Verifiable checklist (commands and expected results) the roadmap tasks will reuse.

## 10. Test strategy
Unit, integration (Postgres), contract (OpenAPI/Schemathesis), E2E (Playwright + axe), load (k6), security (ZAP, TestSecurity_*), data quality asserts.

## 11. Open issues
Numbered; each with owner (architect/owner/vukem) and the trigger that closes it.
```
