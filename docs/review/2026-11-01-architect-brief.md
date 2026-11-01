# Architect review brief (state at 2026-11-01)

Snapshot for the architect's full review of HexaScore (this repo) and the vukem framework (sibling repo `AlexGalhardo/vukem`). It lists what exists, what changed in the 2026-10-31 to 2026-11-01 sessions, the decisions behind it, the known risks and the questions that need direction. Sources of truth stay [AGENTS.md](../../AGENTS.md), [ROADMAP.md](../../ROADMAP.md), [docs/brainstorming.md](../brainstorming.md), the SDDs and the ADRs; this file only points into them.

## 1. Where things stand

| Area | State |
| --- | --- |
| Production | <https://hexascore.app> (+ `www`, paid API at `api.hexascore.app/v1`) on Railway since the 2026-10-31 cutover; Postgres, bucket, healthcheck `/up`. DNS zone on Cloudflare since 2026-11-01 (proxy off, D-28); mail: Resend sends `noreply@`, Cloudflare Email Routing receives `support@` and `privacy@`, DMARC `p=none` |
| Staging | Railway environment `staging` requested 2026-11-01: small seed, no production secrets, no analytics, `noindex` (see `INFRA-staging` once merged) |
| Data | Legacy import complete (users 10,003 of which 10,000 fake seeds now flagged `is_seed` and blocked from sign-in; 2.1 M ratings; 2.05 M statuses; 677 k list items; 40 k entities). Seed users kept with their data by owner decision; `seed:purge` exists but was not run |
| Framework | vukem v0.36.0 released and adopted (analytics, Turnstile and step-up, `llmstxt`, IndexNow, Bing client, export writers, problem decorator, request-log attrs, mail Reply-To). A second feedback batch (VF-063..076) is in a vukem PR, not released |
| Releases | GitHub release series rebuilt as v0.1.0–v0.39.0 (about 35 commits each) on branch `release/history-linear`; v0.39.0 carries a sideload Android APK. PR #65 was squash-merged, so `main` holds one squash commit for that history |
| Roadmap | 286+ of 343 boxes done; 43 open after the 2026-11-01 closures; almost all open boxes are owner gates (store, device tests, decisions Q-A5/A10/A11/A12, flags on in production, 30-day gates). Code-only remainder: the rest of P6.B.3.1 (edition publisher, primary edition, Open Library editions) |
| Mobile | Expo SDK 58 Android app with v1 screens, search, lists, account; release bundle fixed; EAS profiles, Gradle fallback workflow and `mobile-v*` tag workflow in place; Play Console and EAS credentials parked by the owner; app icon chosen (candidate 3) |

## 2. What changed in the last two sessions

- **Product and UI (owner round, 27 items)**: icons and favicon, button depth, landing carousel (12 trending in 30 days, 4 s autoplay with pause) and a client-randomised recommended grid, category icons, item page in three columns, ratings modal, sign-up/login bounds, contact with session identity and 0/512 counter, 2FA merged into `/profile`, `/about`, footer links, `/how-score-works` and `/privacy` with 301s, category pages with a sticky filter sidebar and 25/50/75/100 per page, entity pages under their category (`/games/platforms/...`), covers by slug (`/covers/{category}/{slug}.webp`), HTML 404 → 301 home, notifications removed from the navbar only. Brief: `.orchestration/briefs/OWNER-ROUND-2026-11-01.md` (local).
- **Security**: page scrape guard with verified crawlers, honeypot and step-up (ADR 0022); Turnstile on sign-up, login and contact; log redaction of path tokens and signed URLs (VF-071); signed off-site backup link (P2.E.2.3); session identity on contact; no account enumeration on sign-up.
- **Analytics and SEO**: GA4 behind the consent banner (ADR 0023), Umami cookieless, Search Console verified by DNS, Bing and IndexNow wired, `llms.txt`.
- **Performance**: script budget fixed by lazy-loading the menu sheet, report dialog and TanStack Query; budgets raised by the owner to LCP 2500 ms and script 170 KB (Q-A20) after CI runners measured slower than the dev host.
- **Ingestion (P6)**: Wikidata, Wikipedia, TMDB, IGDB, Open Library, Steam and GOG connectors behind kill switches (legal reads cleared by Q-A16), covers job with slug keys and deferred deletion, rollback with `api_changes` and a visibility trigger, pressure gate, A7 alerts as log lines (D-35).
- **Social (P7)**: follows, feed, reviews, comments, reactions, moderation, friends' score, OG cards, all behind `FLAG_SOCIAL` (off; Q-A5 open).
- **Process**: release history rebuilt; temporary direct pushes to `main` without PR (Q-A21, ruleset disabled); a one-time history rewrite (vukem-style messages, purge of `.orchestration/`, `PROMPT-WORKERS.md`, `vukem-feedback.md`, `questions-to-dev.md`, `architect-feedback-and-roadmap.md`, `prompt.md`) is planned right after the owner round lands.

## 3. Decisions recorded these sessions

| Id | Decision |
| --- | --- |
| Q-A16 | Legal gates for IGDB/Twitch, Open Library, Steam, GOG cleared; licences cite each source's terms |
| Q-A17 | `noreply@` sends automatic mail; `support@` is the only inbox users write to |
| Q-A18 | AI crawlers allowed (rate-limited); no UI copy blocking (except the owner's later item-page `select-none`) |
| Q-A19 | LGPD requests stay on `privacy@` |
| Q-A20 | Lighthouse budgets LCP 2500 ms, script 170 KB |
| Q-A21 | Temporary direct pushes to `main` without PR, after the local gate |
| D-24 (edited) | Password maximum 32 characters for new passwords (was 128); login unaffected |
| ADR 0014 (amended) | Source covers are overwritten in place under slug keys; renamed or rekeyed keys still get the 30-day cleanup |
| ADR 0022 / 0023 | Anti-scraping page guard; GA consent banner |

## 4. Known risks and debts

1. **History and process**: after the rewrite all commit SHAs change; the 39 releases must be recreated again; the ruleset must be re-enabled when PRs return. Until then a push to `main` deploys with only the local gate in front.
2. **Budgets have little headroom**: item pages sit near the script budget; `/games` LCP on CI runners was close to the old 2000 ms limit.
3. **Host limits**: the Windows dev host is memory-starved, Kaspersky invalidates host Lighthouse, Docker Desktop lacks `--network host`; much verification is container- or CI-only. No browser check was done for the latest UI round before review.
4. **vukem gaps still worked around**: VF-063 (app-wide admin clock), VF-064 (bounded counter), VF-065 (client IP and crawler verification), VF-068 (`/up` on an `Only()` host), VF-071..076. Each workaround is marked `workaround(vukem): VF-xxx` in code.
5. **Data**: 10,000 seed users and their ratings still weight scores and trending; the owner decided to keep them for now.
6. **Open Library**: data allowed in the API, covers not (no explicit grant); the covers pipeline copies covers while the provider's guidelines ask for hot-linking. Needs a decision.
7. **Paged URLs** are not crawlable (`?cursor=` disallowed in robots, ADR 0022); only `?per=` pages are.
8. **Observability** is Railway-native only (D-35): alert log lines plus Telegram notifications (P2.C.1.2, in progress); no metrics backend.

## 5. Questions for the architect

1. Keep or revisit D-35 (no metrics backend) now that traffic and ingestion are live?
2. Is the flag-gated social layer (P7) ready to turn on (Q-A5), or does moderation SLA need more first?
3. Seed users: purge before or after the next ranking/score review? Their weight affects trending and recommendations.
4. Release model after the rewrite: keep about-35-commit releases, or return to one release per task group (ADR 0013)?
5. Staging: is a small-seed staging enough, or should staging mirror a sampled production dataset for ingestion and performance tests?
6. vukem: which open VF entries should become framework features versus stay app-level code?
7. Mobile: when to unpark Play Console and EAS, and whether OTA (`expo-updates`) is wanted.

## 6. Where to look

- Plans and status: `ROADMAP.md` (status lines dated 2026-11-01 on open boxes), `docs/PROGRESS.md`.
- Decisions: `docs/brainstorming.md`, `docs/adrs/`, `docs/orchestration/QUESTIONS.md`.
- Designs: `docs/sdd/` (01–14) and `open-issues.md`.
- Operations: `docs/runbooks/` (mobile release, release range, restore and off-site copy, incidents).
- Framework feedback: `vukem-feedback.md` (local after the rewrite; the vukem PRs and CHANGELOG keep the record).
