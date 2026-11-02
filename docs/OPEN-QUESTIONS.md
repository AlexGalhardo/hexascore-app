# Open questions for the owner (architect audit, 2026-11-02)

Each entry: what is blocked, the assumption the audit proceeds with, and the exact action needed. Answer inline under `Answer:`; `docs/orchestration/QUESTIONS.md` keeps the older Q-/FQ- series.

## Production and infrastructure

- OQ-01 Railway pre-deploy command is `/app/app migrate` (MCP `get-service-config`, 2026-11-02) while SDD 11 §7 and AGENTS.md say `hexascore:predeploy` (drops `INVALID` index leftovers first). Assumption: the dashboard should run `/app/app hexascore:predeploy`. Action: change the pre-deploy command in Railway → `app` → Settings → Deploy, production and staging.
  - Answer:
- OQ-02 `staging.hexascore.app` does not resolve (NXDOMAIN on 2026-11-02). Assumption: the CNAME step of `docs/runbooks/staging.md` was not done. Action: add the CNAME on Cloudflare (DNS only) to `app-staging-9215.up.railway.app`.
  - Answer:
- OQ-03 Ingestion is OFF in production (no `FLAG_INGEST_*`, no TMDB/IGDB credentials on the `app` service). RFC-001 assumes ingestion runs in a second Railway service `worker` on the same image with `QUEUE_NAMES=ingest` and the flags on, inside the US$20 cap. Action: confirm the two-service layout (D-08 says "split the worker first when the ceiling is raised"; the ceiling was lowered to US$20 on 2026-11-01, so this is a change of D-08).
  - Answer:
- OQ-04 Telegram operational digest every 30 minutes (PROMPT-ARCHITECT §5.1): page chat (`TELEGRAM_CHAT_ID`, rings) or the silent chat (`TELEGRAM_ERROR_CHAT_ID`)? Assumption: a new silent "ops" chat id variable `TELEGRAM_OPS_CHAT_ID`, falling back to the error chat.
  - Answer:
- OQ-05 LLM usage for extraction/normalisation (RFC-001 §9): which provider and monthly ceiling? Assumption: Anthropic Claude Haiku 5.5, hard ceiling US$3/month, kill switch `FLAG_INGEST_LLM`, off by default.
  - Answer:

## Product decisions

- OQ-06 Missing pages answer `301` to `/` (owner round 2026-11-01). The audit recommends a real `404` page (and `410` for removed items) for crawl budget and user context (W-03). Keep the redirect or switch?
  - Answer:
- OQ-07 Seed users: 10,000 flagged seed accounts still weight ratings, trending and recommendations. RFC-002 assumes `is_seed` users are excluded from every trending signal; purging them (`seed:purge`) is a separate owner call.
  - Answer:
- OQ-08 Affiliate programmes (RFC-003): which accounts exist or will be created (Amazon Associates BR and US, Steam, GOG, itch.io)? Assumption: Amazon BR tag first; everything else plain links with `rel="sponsored nofollow"` and no tag until an account exists.
  - Answer:
- OQ-09 Where-to-watch data: TMDB watch providers (JustWatch data) requires the JustWatch attribution and TMDB's no-commercial-use terms (ADR 0006). Assumption: display-only on the site, never in the paid API, with both attributions. Confirm, or budget for Watchmode.
  - Answer:

## Repository hygiene

- OQ-10 Two stashes (`wip parity rows E24 E25 E26 E39`, `WIP on agent/especialista/P1.G.3.2d-fix`) and three uncommitted changes in the main checkout (`tsconfig.json` types change + `vite-env.d.ts`, deleted architect brief). Keep, commit or drop?
  - Answer:
- OQ-11 81 pre-rewrite local branches "merged by subject only" and 14 with unmerged tips (`docs/audit/CLEANUP.md` B7/B8). Delete all after the backup bundle, or review the 14 first?
  - Answer:
- OQ-12 vukem branches `fix/railway-otlp-traces` (OTLP export kept for traces) and the 4 changelog commits on `release/v0.37.0`: open PRs or drop?
  - Answer:
- OQ-13 Root folder `vercel-env/` holds Vercel-era `.env.*` files. Rotate and delete?
  - Answer:
- OQ-14 Playwright e2e is excluded from CI and the release gate (`RUN_E2E` unset). The audit recommends running it on every release candidate in CI (it caught nothing new today but is the only check of the browser flows). Re-enable?
  - Answer:
