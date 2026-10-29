# E-006 table: markers outside open-issues.md

Method: grep of TODO|FIXME|TBD|NÃO VERIFICADO|NOT IN VUKEM DOCS|SITE PAGE MISSING in docs/, ROADMAP.md, vukem-feedback.md (open-issues.md excluded). tracked = the line is a numbered item of the SDD's Open issues section that open-issues.md lists (id SDD-#); else VF-nnn if the line cites one; else no. Base: origin/dev fc4fa7f.

file:line | marker | text (cut 90) | tracked in open-issues
--- | --- | --- | ---
docs/adrs/0013-versioning-releases-cutover.md:7 | SITE PAGE MISSING | The legacy app is at 0.0.25 with release-please. vukem releases are tagged `vX.Y.Z`, one p... | no
docs/PROGRESS.md:9 | SITE PAGE MISSING | - vukem facts cite the public site <https://vukem.dev/docs> (owner rule, 2026-10-28), neve... | no
docs/PROGRESS.md:119 | SITE PAGE MISSING | - [vukem-feedback.md](../vukem-feedback.md): 31 entries (VF-001 to VF-031) with the owner'... | VF-001 (text ref)
docs/PROGRESS.md:128 | NOT IN VUKEM DOCS | - Several vukem capabilities are `[NOT IN VUKEM DOCS]` and rely on app-level workarounds; ... | no
docs/PROGRESS.md:145 | NOT IN VUKEM DOCS | - Trending API follows SDD 08 (top 10, 30 days) and depends on P6.C; Google sign-in librar... | no
docs/sdd/01-catalog.md:353 | NOT IN VUKEM DOCS | vukem `CursorPaginate` pages by primary key only and rejects `?sort` with a cursor (<https... | no
docs/sdd/01-catalog.md:429 | NOT IN VUKEM DOCS | 8. No sitemap helper in vukem [NOT IN VUKEM DOCS]; an app handler generates it; helper fil... | 01-8
docs/sdd/02-scoring.md:191 | NOT IN VUKEM DOCS | Admin (vukem `https://vukem.dev/docs/digging-deeper/admin`): read-only resources for `scor... | no
docs/sdd/02-scoring.md:313 | NOT IN VUKEM DOCS | i18n: labels in `resources/lang/<locale>.json` as `categories.game.gameplay` etc. (vukem `... | no
docs/sdd/02-scoring.md:320 | NOT IN VUKEM DOCS | 4. Admin custom row actions are [NOT IN VUKEM DOCS]; bursts/flags use custom routes. Owner... | 02-4
docs/sdd/02-scoring.md:321 | NOT IN VUKEM DOCS | 5. orm handling of generated columns (`ratings.overall`) on insert/update is [NOT IN VUKEM... | 02-5
docs/sdd/03-users-auth.md:207 | NOT IN VUKEM DOCS | ### 7.3 Bot protection (replaces Vercel BotID; [NOT IN VUKEM DOCS]) | no
docs/sdd/03-users-auth.md:269 | NOT IN VUKEM DOCS | 4. Bot protection [NOT IN VUKEM DOCS]; ADR 0016 (PoW vs Turnstile) pending. Owner: archite... | 03-4
docs/sdd/03-users-auth.md:271 | NOT IN VUKEM DOCS | 6. Impersonation [NOT IN VUKEM DOCS], not in v1. Minimal design: `POST /admin/_users/{id}/... | 03-6
docs/sdd/04-community-edits.md:355 | NOT IN VUKEM DOCS | 4. vukem admin `BeforeSave` does not document access to the stored record's previous value... | 04-4
docs/sdd/04-community-edits.md:356 | NOT IN VUKEM DOCS | 5. The vukem admin panel documents no custom record actions, bulk actions, assignment or d... | 04-5
docs/sdd/04-community-edits.md:357 | NOT IN VUKEM DOCS | 6. vukem has no persistent in-app notifications [NOT IN VUKEM DOCS] (`https://vukem.dev/do... | 04-6
docs/sdd/05-public-api.md:107 | NOT IN VUKEM DOCS | `api.hexascore.app/v1/*` maps to the `/api/v1` group (legacy `apps/web/proxy.ts`). Host ro... | VF-015 (text ref)
docs/sdd/05-public-api.md:172 | NOT IN VUKEM DOCS | Codes: `unauthorized` 401, `invalid-query`/`invalid-cursor`/`changes-window-exceeded` 400,... | no
docs/sdd/05-public-api.md:195 | NOT IN VUKEM DOCS | Prices (minor units, `legacy/apps/web/lib/billing/catalog.ts`; kept at launch, reviewed 90... | no
docs/sdd/05-public-api.md:234 | NOT IN VUKEM DOCS | Serving: `vukem.ServeOpenAPIWith(app, r, "/api/v1/openapi.json", vukem.OpenAPIOptions{Path... | no
docs/sdd/05-public-api.md:328 | NOT IN VUKEM DOCS | 1. Host routing is [NOT IN VUKEM DOCS]. VF-015; decided: app middleware rewrites the path ... | 05-1
docs/sdd/05-public-api.md:334 | NOT IN VUKEM DOCS | 7. Problem extension members and Scalar UI translation are [NOT IN VUKEM DOCS]. Owner: vuk... | 05-7
docs/sdd/05-public-api.md:335 | NOT IN VUKEM DOCS | 8. Keyset cursor on allowlisted sorts is [NOT IN VUKEM DOCS]; custom cursor used. Owner: v... | 05-8
docs/sdd/06-ingestion.md:135 | NOT IN VUKEM DOCS | Admin: read-only panel resources for `job_runs`, `ingest_failures`, `dq_results`; actions ... | VF-008 (text ref)
docs/sdd/06-ingestion.md:171 | NOT IN VUKEM DOCS | Per-host throttle shared across goroutines (slot reservation, legacy `common/http.py` `_th... | no
docs/sdd/06-ingestion.md:213 | SITE PAGE MISSING | Covers (port of `covers.py`): Open Library (books) → TMDB → IGDB → Steam → GOG → Commons P... | no
docs/sdd/06-ingestion.md:298 | NOT IN VUKEM DOCS | - Observability (SDD 12): `ingest.run.duration{job,source,status}`, `ingest.records{source... | no
docs/sdd/07-search.md:45 | NOT IN VUKEM DOCS | Requires Postgres ≥ 15 (`^@` on SP-GiST, `starts_with`); vukem's suite runs `postgres:17` ... | no
docs/sdd/07-search.md:129 | NOT IN VUKEM DOCS | Index builds on populated tables use `CREATE INDEX CONCURRENTLY`, which cannot run inside ... | VF-017 (text ref)
docs/sdd/07-search.md:291 | NOT IN VUKEM DOCS | 3. Non-transactional migrations (`CREATE INDEX CONCURRENTLY`) are [NOT IN VUKEM DOCS]. VF-... | 07-3
docs/sdd/08-recommendation.md:183 | NOT IN VUKEM DOCS | Dedup uses `cache.Add` ("atomic exclusive claim") on `cache.NewMemory(300_000)` with a TTL... | no
docs/sdd/08-recommendation.md:317 | NOT IN VUKEM DOCS | 10. A buffered counter flush and the `HourlyAt` argument are [NOT IN VUKEM DOCS]. Owner: v... | 08-10
docs/sdd/09-web.md:127 | NOT IN VUKEM DOCS | - Locale order: URL prefix (authoritative); on unprefixed GETs only, cookie `locale`, then... | no
docs/sdd/09-web.md:137 | NOT IN VUKEM DOCS | - No documented hook caches SSR output [NOT IN VUKEM DOCS]; every first load runs SSR. | no
docs/sdd/09-web.md:202 | SITE PAGE MISSING | - `/docs/reference`: `apidocs`' own nonce CSP (<https://vukem.dev/docs/rest-apis/openapi#a... | no
docs/sdd/09-web.md:213 | NOT IN VUKEM DOCS | - `vukem build` = `vite build` through the detected package manager (`bun x vite build`) +... | VF-027 (text ref)
docs/sdd/09-web.md:251 | NOT IN VUKEM DOCS | 1. No URL-prefix source in `i18n.Middleware` [NOT IN VUKEM DOCS]. VF-024; decided: app mid... | 09-1
docs/sdd/09-web.md:253 | NOT IN VUKEM DOCS | 3. No SSR output cache hook [NOT IN VUKEM DOCS]. Owner: vukem. Closes with `SSRConfig.Cach... | 09-3
docs/sdd/09-web.md:256 | NOT IN VUKEM DOCS | 6. `theme` for `data-theme` is [NOT IN VUKEM DOCS] (`inertia.RootData` has `.Inertia`, `.H... | 09-6
docs/sdd/09-web.md:257 | NOT IN VUKEM DOCS | 7. SSR bundle build not covered by `vukem build` [NOT IN VUKEM DOCS]. VF-027; decided: per... | 09-7
docs/sdd/11-infra-deploy.md:5 | SITE PAGE MISSING | Sources, accessed 2026-10-28. Railway (`https://docs.railway.com/…`): [pricing] `reference... | no
docs/sdd/11-infra-deploy.md:95 | NOT IN VUKEM DOCS | 2. `golang:1.27.3-alpine3.24`: `CGO_ENABLED=0 go build -trimpath -tags "$(cat .buildtags)"... | no
docs/sdd/11-infra-deploy.md:98 | NOT IN VUKEM DOCS | Process: `serve` runs the SSR server as a child with `ssr:start` behaviour (restart after ... | no
docs/sdd/11-infra-deploy.md:149 | SITE PAGE MISSING | GitHub Actions is available for `hexascore-app` (D-26; legacy had disabled them 2026-10-03... | no
docs/sdd/11-infra-deploy.md:164 | NOT IN VUKEM DOCS | Actions pinned by SHA. Staging: Railway GitHub autodeploy of `main` [vk-deploy]. Productio... | no
docs/sdd/11-infra-deploy.md:168 | SITE PAGE MISSING | ### 7.1 Release (ADR 0013; `vukem/docs/tooling.md` "Release process", "Tag pushes" [SITE P... | no
docs/sdd/11-infra-deploy.md:181 | NOT IN VUKEM DOCS | - Runner: Railway pre-deploy command, as vukem's Railway setup (`/app/app migrate` [vk-dep... | VF-026 (text ref)
docs/sdd/11-infra-deploy.md:188 | SITE PAGE MISSING | 2. `pg_dump -Fc -Z 6 --no-owner` via `contrib/engine` `RunStream` (`vukem/docs/subsystems.... | no
docs/sdd/11-infra-deploy.md:203 | SITE PAGE MISSING | 1. Status page notice (GitHub Pages, SDD 12 §7.7); `down` [vk-cli] via `railway ssh` (Git ... | no
docs/sdd/11-infra-deploy.md:267 | NOT IN VUKEM DOCS | 3. Migration lock [NOT IN VUKEM DOCS]. VF-026; app-level advisory lock around the runner (... | 11-3
docs/sdd/11-infra-deploy.md:269 | NOT IN VUKEM DOCS | 5. SSR supervision inside `serve`, version in `/up`, extra ldflags in `vukem build` [NOT I... | 11-5
docs/sdd/12-observability.md:32 | SITE PAGE MISSING | Source keys, all accessed 2026-10-28. vukem facts cite <https://vukem.dev/docs>. A fact fo... | no
docs/sdd/12-observability.md:50 | SITE PAGE MISSING | / L-sub / `vukem/docs/subsystems.md` "Logging & observability" [SITE PAGE MISSING] / | no
docs/sdd/12-observability.md:51 | SITE PAGE MISSING | / L-sec / `vukem/docs/security.md` A09/A10 [SITE PAGE MISSING] / | no
docs/sdd/12-observability.md:124 | NOT IN VUKEM DOCS | - `user_hash` = hex(HMAC-SHA256(`app.DeriveKey("log", 32)`, user_id))[:16] [V-conf]. Two c... | no
docs/sdd/12-observability.md:125 | NOT IN VUKEM DOCS | - Sampling: 2xx/3xx on `/up`, `/build/*`, `/search/suggest` and `/search/events` is kept a... | no
docs/sdd/12-observability.md:189 | NOT IN VUKEM DOCS | - Correlation: responses carry `X-Request-ID` (vukem) and `X-Trace-ID` (app middleware). M... | no
docs/sdd/12-observability.md:410 | NOT IN VUKEM DOCS | 2. Request-logger policy: hash the user id, truncate the IP, make the request id equal the... | 12-2
docs/sdd/12-observability.md:415 | NOT IN VUKEM DOCS | 7. Metric Views to drop high-cardinality attributes are [NOT IN VUKEM DOCS]. Owner: vukem.... | 12-7
docs/sdd/12-observability.md:418 | SITE PAGE MISSING | 10. Site documentation gaps, all [SITE PAGE MISSING]. Owner: vukem. Closes when a logging/... | 12-10
docs/sdd/13-security.md:217 | NOT IN VUKEM DOCS |   3. Re-seal the `orm.Encrypted` columns that `NeedsRotation` reports. This job is app cod... | no
docs/sdd/13-security.md:234 | NOT IN VUKEM DOCS |   - Session: idle 30 min, absolute 12 h. Optional `ADMIN_IP_ALLOWLIST` (CIDRs, checked wit... | no
docs/sdd/13-security.md:269 | SITE PAGE MISSING |   - `govulncheck` pinned as a `tool` directive (`vukem/docs/security.md`, `[SITE PAGE MISS... | no
docs/sdd/13-security.md:270 | SITE PAGE MISSING |   - `osv-scanner` on `go.mod` and `bun.lock`, pinned (vukem uses it, `vukem/docs/maintenan... | no
docs/sdd/13-security.md:345 | NOT IN VUKEM DOCS | 7. Role-scoped 2FA, session lifetime and IP allowlist `[NOT IN VUKEM DOCS]`; file a VF ent... | 13-7
docs/sdd/13-security.md:346 | NOT IN VUKEM DOCS | 8. No documented re-seal command after key rotation `[NOT IN VUKEM DOCS]`; VF candidate. O... | 13-8
docs/sdd/14-user-library.md:169 | NOT IN VUKEM DOCS | `vukem make:admin-resource Report` and `ContactMessage`: list/filter by `status`/`handled`... | VF-008 (text ref)
docs/sdd/14-user-library.md:202 | NOT IN VUKEM DOCS | 3. CSV cells get the leading-`'` formula guard of `export.CSV` (<https://vukem.dev/docs/di... | no
docs/sdd/14-user-library.md:245 | NOT IN VUKEM DOCS | 6. Writer-only CSV/JSON helpers for jobs are [NOT IN VUKEM DOCS]. Owner: vukem. Closes wit... | 14-6
docs/sdd/README.md:65 | NOT IN VUKEM DOCS | - Facts about vukem come from the public documentation site **<https://vukem.dev/docs>** (... | no
docs/spike-report.md:29 | NÃO VERIFICADO | Session env names (open issue 03-7): `SESSION_LIFETIME` (absolute), `SESSION_IDLE_TIMEOUT`... | no
ROADMAP.md:13 | NOT IN VUKEM DOCS | - Resolved conflicts between phase files: the trending API (P3.A.1.4) follows SDD 08 (top ... | no
ROADMAP.md:168 | NOT IN VUKEM DOCS |     - `vukem openapi:export -prefix /api/v1 / jq -r '.servers[0].url'` → the API host URL,... | no
ROADMAP.md:228 | NOT IN VUKEM DOCS |   - Notes: `vukem codegen:types` (cited by SDD 09 §6.1) is not on the reference-runtime pa... | no
ROADMAP.md:240 | NOT IN VUKEM DOCS |   - Notes: the scaffold writes `STRIPE_PRICE_BASIC`/`STRIPE_PRICE_PRO`; record how BRL/USD... | no
ROADMAP.md:1240 | SITE PAGE MISSING | - vukem pages read for this phase (2026-10-28): `/docs/deployment/docker-and-railway`, `/d... | VF-023 (text ref)
ROADMAP.md:1301 | NOT IN VUKEM DOCS |   - Notes: the starter uses `oven/bun:1.4.3-alpine` for the assets stage with `--pm bun`; ... | VF-027 (text ref)
ROADMAP.md:1456 | NOT IN VUKEM DOCS |   - Notes: a `version` field inside the `/up` body is [NOT IN VUKEM DOCS] (VF-027): use th... | VF-027 (text ref)
ROADMAP.md:1566 | NOT IN VUKEM DOCS |   - Notes: Ingestion and Moderation dashboards are added by P6.A and P4.B. Above 6,000 ser... | no
ROADMAP.md:1588 | SITE PAGE MISSING |   - Notes: R3, R5 refer to systems built in P6/P4; write them now with a "applies from P4/... | no
ROADMAP.md:1659 | NOT IN VUKEM DOCS |   - Notes: [NOT IN VUKEM DOCS] (issue 7): app middleware, plus a VF entry if none exists. ... | no
ROADMAP.md:1684 | SITE PAGE MISSING |   - Notes: `osv-scanner` is cited from `vukem/docs/maintenance.md` [SITE PAGE MISSING]; it... | no
ROADMAP.md:1752 | SITE PAGE MISSING |   - Notes: `pg_dump -Fc -Z 6 --no-owner` through `contrib/engine` `RunStream` ([SITE PAGE ... | no
ROADMAP.md:2211 | NOT IN VUKEM DOCS |   - Notes: retries per SDD 05 (8 from 30 s) are a sender option; if `SenderOptions` lacks ... | no
ROADMAP.md:2586 | NOT IN VUKEM DOCS |   - Notes: VF-008 workaround: the panel has no custom pages, so app routes under `/admin/m... | VF-008 (text ref)
ROADMAP.md:3150 | NOT IN VUKEM DOCS |   - Notes: queue `ingest`, `queue.Timeout(10*time.Minute)`, `queue.Retries(3)` (that is fo... | VF-018 (text ref)
ROADMAP.md:3316 | NOT IN VUKEM DOCS |   - Notes: `contrib/engine` is [NOT IN VUKEM DOCS]. <https://vukem.dev/docs/deployment/sin... | VF-004 (text ref)
ROADMAP.md:3394 | NOT IN VUKEM DOCS |   - Notes: the triggers are app routes because custom panel actions are [NOT IN VUKEM DOCS... | VF-008 (text ref)
vukem-feedback.md:5 | SITE PAGE MISSING | Documentation reference (owner rule, 2026-10-28): every entry cites the public docs site <... | no
vukem-feedback.md:223 | SITE PAGE MISSING | - **vukem version**: v0.32.0. **Docs consulted**: `https://vukem.dev/docs/digging-deeper/o... | no
vukem-feedback.md:232 | SITE PAGE MISSING | - **Context**: facts that exist only in the repository docs, tagged `[SITE PAGE MISSING]` ... | no
