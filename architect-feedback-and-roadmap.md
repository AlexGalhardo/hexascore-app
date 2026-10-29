# Architect's assessment and executive summary

Author: principal architect and auditor of the HexaScore programme. Date: 2026-10-28. Inputs: Phase 0 diagnosis ([docs/PROGRESS.md](docs/PROGRESS.md)), competitor research ([docs/web-research.md](docs/web-research.md)), owner decisions ([docs/brainstorming.md](docs/brainstorming.md)), 14 SDDs ([docs/sdd/](docs/sdd/README.md)), 16 ADRs ([docs/adrs/](docs/adrs/README.md)), framework feedback ([vukem-feedback.md](vukem-feedback.md)). The execution plan is [ROADMAP.md](ROADMAP.md).

## 1. Executive summary

HexaScore is being rebuilt from Next.js on Vercel to Go on vukem, deployed on Railway under US$30/month, with Inertia + React rendered server-side, PostgreSQL as the only store, ingestion inside the app, a licence-clean paid API, moderated community edits, an Android app, and a published scoring model. The plan reaches parity with the legacy app first (`v1.0.0`, big-bang cutover with 30-day rollback), then adds the differentiators in order: public API DX, community edits, mobile, continuous ingestion and recommendations.

The product position the research supports: nobody covers four media types with one identity model; nobody publishes its score formula; every incumbent handles review bombing reactively and in secret; developer access everywhere is closed, expensive or dead; Brazil's book community (Skoob, 11 M registered users) has no API and a stalled product. HexaScore's wedge is transparency (formula, provenance, moderation status) plus a developer product that is self-serve and honest about licences.

Size of the work: roughly 26 k lines of TypeScript and 10 k of Python become Go plus a React front end; the roadmap has about 180 microtasks sized for one agent session each. The critical path is catalogue model → auth → scoring → web parity → infra/cutover. Mobile and recommendations are off the critical path and can run in parallel after parity.

## 2. Risks, ranked

| # | Risk | Likelihood | Impact | Mitigation in the plan |
| --- | --- | --- | --- | --- |
| R1 | **Framework maturity.** vukem `auth`, `social` are ALPHA, Inertia is BETA; 29 gaps already logged in `vukem-feedback.md` (VF-001 to VF-029), several on the critical path (cursor pagination VF-005, host routing VF-015, non-transactional migrations VF-017, SSR supervision VF-027). | high | high | Pin the vukem version per release; every gap has an in-app workaround specified in the SDD; the owner controls the framework and can fix upstream; the roadmap front-loads a "framework spike" task group that proves the ten riskiest capabilities before feature work. |
| R2 | **SSR in one container under US$30.** Go + Node SSR + ingestion jobs + Postgres pool in one process budget; SSR CPU competes with HTTP. | medium | high | Render caching for anonymous public pages (VF-025 workaround), job concurrency 1–2 with night windows, k6 gate on TTFB p75 ≤ 400 ms before cutover; first budget increase buys a separate worker. |
| R3 | **Hard spend limit at the ceiling (D-27).** Railway cuts the service at US$30, so a legitimate traffic spike can take the site down at month end. | medium | high | Projected-spend alert at day 20 (SDD 12) and the explicit recommendation below to set the hard limit at 1.5× the ceiling. |
| R4 | **Cutover is big-bang.** One risky moment with data migration, DNS and auth session invalidation. | medium | high | Staging with a production copy, rehearsed migration runbook (SDD 11), read-only window ≤ 60 min, DNS TTL lowered 24 h before, 30-day rollback with a reconciliation rule for post-cutover writes. |
| R5 | **Licensing gaps still open.** Twitch Developer Services Agreement unread (D-36), Open Library licence unchecked (A-05), Steam/GOG terms unread; TMDB 6-month cache rule enforced by a job that must never fail silently. | medium | medium | Owner tasks before the ingestion milestone; connectors switched off until the legal note exists; provenance per field makes any later change a configuration change. |
| R6 | **Moderation load on one person.** 3 suggestions/day/user with a 24 h SLA and admins-only moderation. | medium | medium | Queue ageing alerts, bulk actions, accuracy tracking from day one so trusted moderators can be enabled without migration; quota can be lowered by configuration. |
| R7 | **Score manipulation after publishing the formula.** | medium | medium | Published rules plus unpublished shadow-weight parameters, burst limiter, account-age and verified-activity weights, nightly self-healing recompute, public notice on affected items. |
| R8 | **Rewrite scope creep.** Fourteen SDDs is a lot of surface for a solo owner with agents. | high | medium | Parity cut (D-06) and deferrals (D-30) are explicit; `v1.0.0` is defined as parity, not as "all SDDs"; the roadmap marks post-1.0 work clearly. |
| R9 | **Better Auth migration details unverified** (scrypt parameters, 2FA secret format). | medium | medium | First auth task verifies against a hash produced by the legacy scripts; export of 2FA secrets through Better Auth's own code. |
| R10 | **GitHub Actions billing** was blocked on the legacy repo. | low | medium | Owner confirmed availability (D-26); pre-push mirror kept as fallback. |

## 3. Technical debt inherited or created

- Legacy: synthetic users in production (10 k users, ~2.1 M ratings) must be purged before the data copy; docs drift (AGENTS.md "no code yet"); unused Telegram and Blob env vars; `unaccent` installed but unused; enum values never written; orphan Blob files. All handled in the migration runbook, none carried into the rewrite.
- Created by decisions: the Node SSR process (ADR 0004) is the only non-Go runtime and the only piece that can fall back silently (VF-002); the workaround is an alert, not a fix. Two implementations of the star widget (DOM and React Native) by design (ADR 0009). Cover pipeline depends on a `cwebp` binary through `contrib/engine` until vukem's imaging supports WebP (VF-004).
- Documentation debt: several competitor facts rest on secondary sources (list in `docs/web-research.md` §10); the vukem site lacks pages for health checks, logging options, tooling and stability (VF-023).

## 4. Decisions I disagree with (recorded, not reopened)

1. **Inertia + React instead of HTMX (D-04).** For a catalogue whose growth is organic search, server-rendered HTML with minimal JS is the stronger default: smaller bundles, no Node, vukem's CSP tests cover it, and the owner's own PageSpeed references were achieved that way. The owner chose React reuse and interactivity; the SSR condition (D-07) limits the damage. If Core Web Vitals budgets fail on staging, the fallback is the hybrid (HTMX public pages, Inertia logged-in area), already written up in Round 2.
2. **US$30 ceiling with one process (D-08).** Reasonable at launch, but it forces ingestion, scheduler, HTTP and SSR into one CPU budget and leaves no room for a worker. My recommendation was US$60 with a separate worker. The first thing to buy when revenue appears is that worker.
3. **Hard limit at US$30 (D-27).** A hard limit equal to the ceiling trades a surprise bill for a surprise outage. Set the hard limit at US$45 and keep the alert at US$24; the projected-spend alert gives time to react.
4. **Stars 0–10 everywhere (D-13).** The 0–100 "HexaScore" was a brand asset and finer-grained; stars are more familiar. Acceptable; the confidence indicator survives as a field and UI element, which was the part that mattered.
5. **No commercial agreements with TMDB/IGDB now (D-03).** Cheapest and compliant, but it makes the paid API thinner for films and games than the site. The provenance model keeps the door open; I would still send the two emails this month because the answers take months.

## 5. Recommendations, prioritised

1. **Run the framework spike first** (ROADMAP phase 0 group F): prove auth scaffold + Better Auth hash compatibility, Inertia SSR in one container with readiness, cursor pagination, host routing, Postgres migrations with concurrent index script, queue + scheduler inside `serve`, OpenAPI export + generated client, Stripe scaffold, storage to a Railway bucket, Lighthouse CI on a sample page. Two sessions; outcome: VF entries confirmed or closed, risk R1 quantified.
2. **Decide the hard-limit number again after the first Railway bill** (R3).
3. **Send the licence emails now** (TMDB sales, IGDB partner) even though nothing depends on them; record answers in `docs/legal/`.
4. **Treat moderation SLA as a product metric** from day one; show it publicly ("median decision time") once it is good.
5. **Keep the parity bar honest**: the cutover gate is the legacy QA suite ported (Playwright specs, k6, Schemathesis, ZAP) plus Lighthouse budgets; no feature not in the legacy app enters the parity milestone.
6. **Publish the scoring formula page with the first beta**; it is the cheapest differentiator and it disciplines the implementation.
7. **Plan the second budget step** (US$60): worker service, WAL archiving for RPO ≤ 1 h, Redis only if a second app instance appears.

## 6. What "done" looks like

- `v1.0.0`: legacy parity (minus D-06 deferrals), cutover complete, rollback window closed, Core Web Vitals and API latency budgets met on production, backups restored once in a drill.
- `v1.x`: public API DX items (changes feed, exports, webhooks), community edits live with SLA, Android app on the production track, trending and similar items live, ingestion continuous with asserts green for 30 days.
- Metrics to watch: p75 LCP item page, API p95, moderation median time, ingestion assert pass rate, ratings/day from verified users, API paid conversions, monthly cost vs ceiling.

## 7. Open items owned by the owner

- Read and record the Twitch Developer Services Agreement (`docs/legal/igdb.md`, D-36) and the Open Library licence (A-05) before the ingestion milestone.
- Answer SDD 14 open issue 7 (public lists of deleted users: anonymise or delete; default anonymise) and any SDD 13 owner items when it lands.
- Legal review of terms, privacy policy and LGPD bases before cutover (D-29, D-21).
- Review prices 90 days after cutover (D-33).
