# Open questions for the owner (architect audit, 2026-11-02)

Each entry: what is blocked, the assumption the audit proceeds with, and the exact action needed. Answer inline under `Answer:`; `docs/orchestration/QUESTIONS.md` keeps the older Q-/FQ- series.

## Production and infrastructure

- OQ-01 Railway pre-deploy command is `/app/app migrate` (MCP `get-service-config`, 2026-11-02) while SDD 11 §7 and AGENTS.md say `hexascore:predeploy` (drops `INVALID` index leftovers first). Assumption: the dashboard should run `/app/app hexascore:predeploy`. Action: change the pre-deploy command in Railway → `app` → Settings → Deploy, production and staging.
  - Answer (owner, 2026-11-02): done. Railway pre-deploy command is `/app/app hexascore:predeploy` (verified with `get-service-config`).
- OQ-02 `staging.hexascore.app` does not resolve (NXDOMAIN on 2026-11-02). Assumption: the CNAME step of `docs/runbooks/staging.md` was not done. Action: add the CNAME on Cloudflare (DNS only) to `app-staging-9215.up.railway.app`.
  - Answer (owner, 2026-11-02): CNAME added; `staging.hexascore.app` resolves to `app-staging-9215.up.railway.app` and the custom domain is ACTIVE on the Railway service. TLS handshake still failed at 20:15 UTC (certificate provisioning); re-check with `curl https://staging.hexascore.app/up`.
- OQ-03 Ingestion is OFF in production (no `FLAG_INGEST_*`, no TMDB/IGDB credentials on the `app` service). RFC-001 assumes ingestion runs in a second Railway service `worker` on the same image with `QUEUE_NAMES=ingest` and the flags on, inside the US$20 cap. Action: confirm the two-service layout (D-08 says "split the worker first when the ceiling is raised"; the ceiling was lowered to US$20 on 2026-11-01, so this is a change of D-08).
  - Answer (owner, 2026-11-02): separate Railway service `worker` on the same image (RFC-001 §3 decision 1). Amends D-08; recorded as D-40 in `docs/brainstorming.md`.
- OQ-04 Telegram operational digest every 30 minutes (PROMPT-ARCHITECT §5.1): page chat (`TELEGRAM_CHAT_ID`, rings) or the silent chat (`TELEGRAM_ERROR_CHAT_ID`)? Assumption: a new silent "ops" chat id variable `TELEGRAM_OPS_CHAT_ID`, falling back to the error chat.
  - Answer (owner, 2026-11-02): new silent chat through `TELEGRAM_OPS_CHAT_ID`, falling back to `TELEGRAM_ERROR_CHAT_ID`; the page chat keeps only critical alerts.
- OQ-05 LLM usage for extraction/normalisation (RFC-001 §9): which provider and monthly ceiling? Assumption: Anthropic Claude Haiku 5.5, hard ceiling US$3/month, kill switch `FLAG_INGEST_LLM`, off by default.
  - Answer (owner, 2026-11-02): Claude Haiku 5.5 behind `FLAG_INGEST_LLM` (default off), hard ceiling `LLM_MONTHLY_BUDGET_USD=3`, outputs always to `field_reviews`. Pricing to be verified when CW-10 starts.

## Product decisions

- OQ-06 Missing pages answer `301` to `/` (owner round 2026-11-01). The audit recommends a real `404` page (and `410` for removed items) for crawl budget and user context (W-03). Keep the redirect or switch?
  - Answer (owner, 2026-11-02): switch to a real 404 page (410 for removed items); merge redirects stay in `catalog_redirects`. Backlog B-021 unblocked; recorded as D-41.
- OQ-07 Seed users: 10,000 flagged seed accounts still weight ratings, trending and recommendations. RFC-002 assumes `is_seed` users are excluded from every trending signal; purging them (`seed:purge`) is a separate owner call.
  - Answer (owner, 2026-11-02): exclude `users.is_seed` accounts from every trending and recommendation signal (RFC-002); `seed:purge` only on an explicit later request, dry run first.
- OQ-08 Affiliate programmes (RFC-003): which accounts exist or will be created (Amazon Associates BR and US, Steam, GOG, itch.io)? Assumption: Amazon BR tag first; everything else plain links with `rel="sponsored nofollow"` and no tag until an account exists.
  - Answer (owner, 2026-11-02): programmes to open: Amazon Associates BR first, then Amazon Associates US, then the game stores (Epic creator code, itch.io partners, GOG affiliate). Until each account exists its links stay plain with `rel="sponsored nofollow"`. Owner action: enrolment; tags go into `AFFILIATE_*` variables.
- OQ-09 Where-to-watch data: TMDB watch providers (JustWatch data) requires the JustWatch attribution and TMDB's no-commercial-use terms (ADR 0006). Assumption: display-only on the site, never in the paid API, with both attributions. Confirm, or budget for Watchmode.
  - Answer (owner, 2026-11-02): TMDB watch providers, display-only on the site, with the TMDB and JustWatch attributions, 180-day expiry; never in `/api/v1`. Watchmode not funded.

## Repository hygiene

- OQ-10 Two stashes (`wip parity rows E24 E25 E26 E39`, `WIP on agent/especialista/P1.G.3.2d-fix`) and three uncommitted changes in the main checkout (`tsconfig.json` types change + `vite-env.d.ts`, deleted architect brief). Keep, commit or drop?
  - Answer (owner, 2026-11-02): discard the three loose changes (`tsconfig.json`, `resources/js/vite-env.d.ts`, the deleted brief restored) and drop both stashes. Done in this session.
- OQ-11 81 pre-rewrite local branches "merged by subject only" and 14 with unmerged tips (`docs/audit/CLEANUP.md` B7/B8). Delete all after the backup bundle, or review the 14 first?
  - Answer (owner, 2026-11-02): delete the 81 branches merged by subject (bundle keeps them); keep the 14 with unmerged tips for a senior review task (cherry-pick what still matters, delete the rest). Done for the 81 in this session.
- OQ-12 vukem branches `fix/railway-otlp-traces` (OTLP export kept for traces) and the 4 changelog commits on `release/v0.37.0`: open PRs or drop?
  - Answer (owner, 2026-11-02): open a PR for `fix/railway-otlp-traces`; drop the changelog commits of `release/v0.37.0` and the older `agent/implementer/railway-observability` variant. Outcome: the OTLP change was already squash-merged as vukem PR #104 on 2026-10-31 (`drivers/otel/provider.go` on `main` is byte-identical to the branch), so the duplicate PR #143 was closed and the three local branches deleted; nothing was lost.
- OQ-13 Root folder `vercel-env/` holds Vercel-era `.env.*` files. Rotate and delete?
  - Answer (owner, 2026-11-02): delete `vercel-env/` without rotation (nothing there is considered valid). Done in this session.
- OQ-14 Playwright e2e is excluded from CI and the release gate (`RUN_E2E` unset). The audit recommends running it on every release candidate in CI (it caught nothing new today but is the only check of the browser flows). Re-enable?
  - Answer (owner, 2026-11-02): run the Playwright `desktop-en` project on release candidates in CI after the helper fix (B-012); normal pushes stay without e2e. Recorded as D-42.
- OQ-15 (2026-11-02, from B-020): the item page at phone width shows cover, actions, details, the score card and the full rating widget before the H1 and synopsis (owner order of 2026-11-02, `Show.tsx`, commit 2b107e2a); the audit (W-02) asks for the title and synopsis first. Keep the owner order or follow the audit?
  - Answer (owner, 2026-11-02): follow the audit: title and synopsis before the rating widget on phones (B-020).
