# Phase 0 spike report

Findings from the throwaway spike app (`C:\dev\hexascore-app\hexascore-spike`, local git, never pushed). Each microtask appends a section; P0.F.11.1 consolidates them. Verdict values: works, workaround, blocked.

## Summary (P0.F.11.1, 2026-10-29)

Every Phase 0 capability was proven on the spike (`C:\dev\hexascore-app\hexascore-spike`, vukem v0.32.0). No capability is blocked; the workarounds are app code that adopts native vukem support as releases land (vukem `main` already merged VF-002, VF-005, VF-015, VF-024 and VF-027, unreleased because the v0.33.0 release is waiting on the owner, `questions-to-dev.md` Q-D-006).

| Capability | Verdict | Decisive command | Decisive output | VFs |
| --- | --- | --- | --- | --- |
| P0.F.1 auth scaffold, Better Auth hashes, compat hasher | works (es catalogues by hand) | `go test ./app/accounts -run TestSecurity_A07_LegacyScryptVerifiesAndRehashes` | PASS; stored hash becomes `$argon2id$` | VF-032, VF-033 |
| P0.F.2 Inertia SSR, single container, locale prefix | workaround | `docker exec <id> pkill node` then `GET /up?checks=1` | `ssr: ok` again after 1.3 s | VF-002, VF-024, VF-027, VF-028, VF-038, VF-041 |
| P0.F.3 signed keyset cursor with `Scroll` | workaround | `go test ./app/catalog/... -run TestCursor` | PASS; 13 pages, 300 unique rows | VF-005, VF-039 |
| P0.F.4 host routing for `api.<host>/v1` | workaround | `curl -H 'Host: api.localhost' /v1/ping` | same body as `/api/v1/ping` | VF-015, VF-034 |
| P0.F.5 migrations, jsonb, CHECK, lock, concurrent index | workaround | two parallel `migrate:locked` | one migrates, one prints `Nothing to migrate.` | VF-006, VF-009, VF-017, VF-026, VF-035, VF-036, VF-040 |
| P0.F.6 queue and scheduler inside `serve` | workaround | 200 jobs, SIGTERM and SIGKILL mid-run | `effects=200 distinct_keys=200`, `r2_effects=40`, `failed=0` | VF-018, VF-021 (withdrawn) |
| P0.F.7 OpenAPI documents, problems, clients | workaround | `go test ./tests -run TestAPIContract` | ok on `/api/v1` and `/app/v1` | VF-014, VF-016, VF-042, VF-044 |
| P0.F.8 Stripe billing scaffold | workaround | `go test ./... -run 'TestBilling'` | 10 tests PASS | VF-037 |
| P0.F.9 S3 storage, covers, export ZIP | works | `go test ./tests -run 'TestCoversServe\|TestPresignExpires\|TestExportJob\|TestZipCSV'` | 4 PASS | VF-029, VF-045 |
| P0.F.10 Lighthouse CI budgets | works (after gzip) | `bunx lhci autorun` | LCP 1.37 s, CLS 0, TBT ≤ 11 ms, JS 115 kB | VF-043 |

## P0.F.1.1 Auth scaffold

Environment: vukem CLI and framework v0.32.0 (CLI installed with `go install github.com/AlexGalhardo/vukem/cmd/vukem@v0.32.0`), Go 1.27.3, Bun 1.4.4, Postgres `postgres:18.6-alpine3.24` (tag exists, pulled). Spike branch `agent/senior/P0.F.1.1` (commits 0674406, 185d098). go.mod requires the tag `v0.32.0` with no `replace` and `-framework` was not needed.

Commands run:

- `vukem new -inertia react -pm bun -docker hexascore-spike` (flags go before the name; `-docker` gives `compose.yaml` with the Postgres service). Postgres driver: `vukem.WithDatabaseDrivers(postgres.Driver())` in `bootstrap/app.go`; the driver is a separate module (`go get github.com/AlexGalhardo/vukem/drivers/postgres@v0.5.1`).
- `vukem install:auth -social google -2fa -locales en,pt-BR,es` exits 2: `no auth message catalog for locale "es" (one of en, pt-BR)` (VF-032). Rerun with `-locales en,pt-BR` exits 0 and `app/accounts/accounts.go` exists.
- `vukem migrate` ran `create_posts_table`, `create_users_table`, `create_auth_tables`, `create_social_accounts_table`, `add_two_factor_columns`; `vukem migrate:status --json` lists all five `"ran":true,"batch":1` on Postgres 18.
- `go test ./tests/...` with `TEST_DATABASE_URL` set: `ok  hexascore-spike/tests 1.298s`. These scaffold tests run on in-memory SQLite, not Postgres (VF-033); an added `tests/postgres_test.go::TestRegisterOnPostgres` (pgtest + `Options.Env DATABASE_URL`) passes on Postgres 18 and the whole package stays `ok`.

Tables and columns:

- `users`: id, name, email (unique), password, email_verified_at (null), session_version (bigint, default 0, bumped to revoke all sessions), timestamps; `add_two_factor_columns` adds two_factor_secret (text, orm.Encrypted), two_factor_recovery_codes (text, encrypted SHA-256 hashes), two_factor_confirmed_at.
- `remember_tokens`: id (unique), user_id (index), hash, expires_at.
- `access_tokens`: id (unique), user_id (index), name, hash, abilities (text), last_used_at, expires_at, revoked_at, created_at.
- `password_resets`: email (unique), hash, created_at.
- `social_accounts`: id, user_id (FK users, cascade), provider, provider_id (unique together), timestamps.
- There is no sessions table in the scaffold (`SESSION_DRIVER` default is not `database`).

Pages: React `.tsx` (TypeScript) under `resources/js/pages/Auth/`: Login, Register, CheckEmail, ForgotPassword, ResetPassword, VerifyEmail, TwoFactor, TwoFactorChallenge; `Dashboard.tsx` updated. The pages do not set a layout; they reuse the starter `layouts/AppLayout.tsx` through the default layout in `resources/js/app.tsx`. No Google-specific page: social is route-only.

Routes: GET/POST `/login`, `/register`, GET `/register/check-email`, POST `/logout`, GET `/forgot-password`, POST `/forgot-password`, GET `/reset-password/{token}`, POST `/reset-password`, GET `/email/verify`, `/email/verify/{id}`, POST `/email/verification-notification`, GET `/dashboard`, GET `/auth/{provider}/redirect`, GET/POST `/auth/{provider}/callback`, GET `/user/two-factor`, POST `/user/two-factor`, `/user/two-factor/confirm`, `/user/two-factor/recovery-codes`, `/user/two-factor/disable`, GET/POST `/two-factor-challenge`. Google needs `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET` (dummy in spike `.env`).

Session env names (open issue 03-7): `SESSION_LIFETIME` (absolute), `SESSION_IDLE_TIMEOUT`, plus `SESSION_DRIVER`, `SESSION_COOKIE`, `SESSION_SECURE`, `SESSION_RECORD_CLIENT` (found in framework source `session/`; durations are time.Duration fields, string format [NÃO VERIFICADO]).

Verdict: works, with one workaround: generate `en,pt-BR` and hand-translate `es` catalogs (VF-032).

## P0.F.1.2 Better Auth hash format

Verdict: works (format pinned, fixtures produced, cross-verified outside Node).

Source read: `better-auth` 1.7.7 (the version `legacy/apps/web/package.json:32` pins) installed with `bun install --ignore-scripts` in a temporary folder outside `legacy/`. Better Auth delegates hashing to its dependency `@better-auth/utils` 0.4.2 (`better-auth/package.json:466`); `better-auth/dist/crypto/password.mjs:1-13` re-exports `hashPassword` / `verifyPassword` from `@better-auth/utils/password`, which resolves to `node:crypto` scrypt on Node and Bun and to `@noble/hashes` scrypt elsewhere. The legacy app does not override the hasher (`legacy/apps/web/lib/auth.ts` has no `password.hash` / `password.verify` option).

Parameters (`@better-auth/utils/dist/password.node.mjs:3-34`):

| Parameter | Value | Line |
| --- | --- | --- |
| Algorithm | scrypt (`node:crypto`) | 1, 11 |
| N | 16384 | 4 |
| r | 16 | 5 |
| p | 1 | 6 |
| dkLen | 64 bytes | 7 |
| maxmem | `128 * N * r * 2` (64 MiB) | 19 |
| Salt | 16 random bytes, hex-encoded to a 32-character string | 31 |
| Password input | `password.normalize("NFKC")` | 12 |
| Stored format | `<salt hex>:<derived key hex>` (32 + 1 + 128 characters) | 33 |

Pitfall for the Go verifier: the scrypt salt is the 32-character hex **text** (its ASCII bytes), not the 16 decoded bytes. `verifyPassword` splits on `:` and passes the hex string straight to scrypt (lines 35-41). Better Auth compares the hex strings with `===`; the Go verifier must use a constant-time compare instead.

Fixtures: `docs/spike/testdata/betterauth_hashes.json` (to be copied into the spike as `testdata/betterauth_hashes.json` by P0.F.1.3), generated from throwaway passwords with `hashPassword` and self-checked with `verifyPassword` (correct password verifies, password + `x` fails):

- `ascii`: 28-character ASCII password.
- `accented-decomposed-nfkc`: decomposed accents (`e` + U+0301, `c` + U+0327, `a` + U+0303) and the ligature U+FB01; its NFKC-composed equivalent (`nfkc_equivalent`) also verifies.
- `32-chars`: 32-character password with symbols.

Independent check with Python `hashlib.scrypt` (stdlib) over the same fixtures: salt as hex text matches 3/3, salt as decoded bytes matches 0/3, and skipping NFKC breaks the accented case only. This is the behaviour the compatibility hasher in P0.F.1.3 must reproduce.

2FA secrets (open issue 03-2, export not done here): the TOTP plugin decrypts `twoFactor.secret` with `symmetricDecrypt({ key: ctx.context.secretConfig, data })` (`better-auth/dist/plugins/two-factor/totp/index.mjs:122-125`, defined in `better-auth/dist/crypto/index.mjs:48`); backup codes use the same function when `storeBackupCodes` is `"encrypted"` (`two-factor/backup-codes/index.mjs:45`). The key is the Better Auth secret configuration, so the export script needs `BETTER_AUTH_SECRET` at export time.

Acceptance:

- `rg -n "scrypt|NFKC" <copy>/node_modules/better-auth/dist <copy>/node_modules/@better-auth` → `@better-auth/utils/dist/password.node.mjs:1` (`import { randomBytes, scrypt } from 'node:crypto'`), `:12` (`password.normalize("NFKC")`); constants quoted above with lines.
- `jq '.cases | length' docs/spike/testdata/betterauth_hashes.json` → `3`.

## P0.F.4.1 Host routing for `api.<host>/v1`

Spike branch `agent/senior/P0.F.4.1` (commit 6570235). Files: `app/http/middleware/api_host.go` (+ `_test.go`), `routes/api.go` (`/api/v1/ping` group and `r.UseServer(middleware.APIHost(API_HOST))`), `routes/registry.go`, `bootstrap/app.go` (`vukem.WithOpenAPI(openapi.Info{Servers: ...})` from `API_URL`), `tests/api_host_test.go`. The server middleware runs before the mux (`router.build`: `r.server` wraps `dispatch`), so it rewrites `/v1/` to `/api/v1/` before route matching; on the API host any other path (`/`, `/login`, `/api/v1/*`, `/v10/*`) is 404.

Env vars the tests need: none for `go test ./app/http/middleware/` and `go test ./tests/ -run APIHost` (in-memory SQLite, `API_HOST` is set through `vukemtest.Options.Env`). The full `go test ./...` also needs `TEST_DATABASE_URL` (postgres admin URL) for `TestRegisterOnPostgres`. Runtime env: `API_HOST=api.localhost`, `API_URL=<http://api.localhost:3001/v1>` (process env, for OpenAPI servers), `PORT=3001`.

Decisive output (server on port 3001, spike Postgres):

- `curl -s localhost:3001/api/v1/ping` and `curl -s -H 'Host: api.localhost' localhost:3001/v1/ping` both return `{"pong":"ok"}`.
- `curl -s -o /dev/null -w '%{http_code}' -H 'Host: api.localhost' localhost:3001/` returns `404`.
- `MSYS_NO_PATHCONV=1 vukem openapi:export -prefix /api/v1 | jq -r '.servers[0].url'` with `API_URL` set returns `http://api.localhost:3001/v1` (without it: `http://localhost:3000`, the `APP_URL` default). Paths stay `/api/v1/ping`, so the post-processing step `jq '.paths |= with_entries(.key |= sub("^/api/v1";""))'` is needed (VF-034). `//api/v1` in Git Bash failed (see VF-034).
- `go test ./app/http/middleware/` `ok`; `go test ./tests/ -run APIHost` `ok`; full `go test ./...` `ok hexascore-spike/tests`.

Verdict: works (workaround: OpenAPI servers via `bootstrap` process env, paths post-processed, VF-034; host routing itself is app middleware, VF-015).

## P0.F.5.1 jsonb, CHECK enums and a generated column

Spike branch `agent/senior/P0.F.5.1` (eba1939). Files: `database/migrations/20261029010000_create_items_table.go`, `20261029010001_create_ratings_table.go` (registered in `registry.go`), `tests/orm_generated_test.go`. Both migrations use `s.Raw` for the whole table (uuid primary key, composite key, smallint and generated column are not in the builder: VF-035) and return early when `s.Dialect() != "postgres"` so the SQLite-based scaffold tests keep passing (VF-033). `items_type_probe` is a builder-made table that records the type mapping.

Env the test needs: `TEST_DATABASE_URL` (postgres admin URL; `pgtest.Database` creates and drops a private database). Migrations: `DATABASE_URL` (DB `spike_senior`), loaded from the spike `.env`.

Decisive output:

- `vukem migrate; vukem migrate:rollback; vukem migrate` each exit 0 (`Rolled back: ..._create_ratings_table`, `..._create_items_table`, then both `Migrated:` again); `migrate:status --json` shows both `"ran":true`.
- `go test ./tests -run TestGeneratedOverallColumn -v` PASS: insert via `orm.Create` succeeds; after `Update(orm.Set{"c1": 100})` through the query builder `overall` goes from 525/6 to 545/6 and `round(overall)` equals `round(sum/6)`; `go test ./...` `ok hexascore-spike/tests`.
- Type mapping on PostgreSQL 18 (information_schema): `t.JSON` is `jsonb`, `t.Enum` is `text` plus a CHECK (a value outside the list is rejected), `t.UUID` is `uuid`, `ratings.overall` is `numeric` with `is_generated = ALWAYS`.
- ORM: `orm.JSON[T]` round-trips `jsonb`; CHECK violations (`type = 'podcast'`) surface as errors; the ORM writes every mapped column, so a model mapping `overall` fails with `cannot insert a non-DEFAULT value into column "overall"` (SQLSTATE 428C9): leave the column out of the write model (or `db:"-"`) and read it through a separate read model. `orm.Update`/`orm.Delete` filter by the single `PrimaryKey()` column, so composite-key tables must use `orm.QueryOf[T](db).Where(...).Where(...).Update(...)` (VF-035). Acronym fields need a `db` tag (VF-036). Fallback raw SQL was not needed.

Verdict: workaround (raw DDL for these tables; ORM works with the models above; VF-009 updated, VF-035, VF-036).

## P0.F.8.1 Billing scaffold, pro/business and per-currency prices

Spike branch `agent/senior/P0.F.8.1` (c316d45, from spike `main` 01bdac9). `vukem install:billing` added `app/billing`, `app/controllers/account/billing*.go`, `app/models/billing.go`, `config/billing.go`, `routes/billing.go`, `tests/billing_test.go`, `tests/billing_page_test.go`, `resources/js/pages/Billing.tsx`, migration `20261029011004_add_billing_tables.go`, and `contrib/stripe v0.4.1` in `go.mod` (latest tag, no `replace`). Beyond the listed files I also edited `app/billing/billing.go` and `app/controllers/account/billing.go` (plan names in comments, `Checkout` takes a currency, plan view exposes `amounts`), because the currency has to reach `CreateCheckoutSession`.

Env the tests need: `TEST_DATABASE_URL` is not required for these tests (they use `vukemtest` on in-memory SQLite with a fake Stripe server and the env set in `billingApp`: `STRIPE_KEY=rk_test_vukem`, `STRIPE_WEBHOOK_SECRET`, `STRIPE_API_BASE`, `STRIPE_PRICE_PRO`, `STRIPE_PRICE_BUSINESS`; all fake). Real runs set the same names in `.env`; `.env.example` now lists `STRIPE_PRICE_PRO` and `STRIPE_PRICE_BUSINESS` instead of `..._BASIC`/`..._PRO`.

Decisive output:

- `vukem install:billing` exit 0. `vukem route:list` shows `GET /billing`, `POST /billing/checkout/{plan}`, `POST /billing/portal`; `POST /stripe/webhook` is not listed because it is mounted with `r.UseServer` (documented), so it is proven by the webhook tests below.
- `go test ./tests/... -run 'TestBilling|TestStripeWebhook' -v`: all PASS, including `TestBillingWebhookProcessedOnce` (duplicate event processed once), `TestBillingWebhookRejectsBadSignature` (forged signature 400), `TestBillingWebhookSyncsSubscription` (a `past_due` subscription loses access, free), and the new `TestBillingCheckoutCurrency` (BRL on `business` and default USD on `pro` reach Stripe with `currency` and the right price; `gbp` returns 422 and never reaches Stripe). There is no test named `TestStripeWebhook*`; the webhook tests are named `TestBillingWebhook*` and match the `TestBilling` alternative. Full `go test ./...` is `ok`. `vukem migrate` applied `add_billing_tables`.
- Fit of BRL/USD/EUR: one Stripe price per plan (`STRIPE_PRICE_PRO`, `STRIPE_PRICE_BUSINESS`) with `currency_options`; the checkout sends `currency` in `stripe.Params`; `config.Plan.Amounts` holds the SDD amounts (pro 2900/600/600, business 14900/2900/2700 minor units, BRL/USD/EUR) for display only. Locale-to-currency default (pt-BR to BRL, es to EUR, en to USD) is not implemented: the default is `usd`. Live Stripe acceptance of `currency` with a `currency_options` price was not run (test keys only, human gate): [NÃO VERIFICADO].

Verdict: workaround (scaffold code edited for plan names and currency; VF-037).

## P0.F.2.1 Inertia SSR with failure counter and health check

Verdict: workaround (works; two project-side fixes and VF-038).

Spike branch `agent/orchestrator/P0.F.2.1` (merged into spike `main` as 13b4c0a). Port 3000, database `spike_orch`.

- `bootstrap/app.go`: `inertia.SSRConfig{Bundle: "bootstrap/ssr/ssr.js", OnFailure: ssrFailures.Observe}`; `app.AddHealthCheck("ssr", app.Inertia.SSRHealthy)` adds `ssr` to `GET /up?checks=1` (<https://vukem.dev/docs/deployment/single-binary> "Health checks"); `GET /up/ssr-fallbacks` exposes the counter for the spike only (production exports a metric, SDD 12).
- `app/ssr/failures.go`: atomic counter plus a warning log per fallback (VF-002 workaround); `TestFailuresCountsEachFallback` passes.
- `resources/js/ssr.tsx`: the `@inertiajs/vite` 3.9.1 plugin wraps a top-level `createInertiaApp({...})` **expression statement** with `createServer` and `renderToString`. An `export default createInertiaApp(...)` entry builds without errors, but it produces a bundle that never listens, and `ssr:start` then exits 0 with no log.
- `resources/vite/vukem.js` (project-owned starter plugin): during the SSR build it forced `outDir` to `resources/public/build` with `emptyOutDir: true`. The server bundle then landed in the public directory and the client build was wiped. Fixed: when `isSsrBuild` is set, the output goes to `bootstrap/ssr`. `package.json` gets `build:ssr` = `vite build --ssr resources/js/ssr.tsx`.

Acceptance (run on 2026-10-29):

- `bun run build:ssr && <app> ssr:start`, then `<app> ssr:check` → `SSR server is healthy`, exit 0.
- `GET /` → exactly one `data-page` attribute, and `#app` holds server-rendered markup (`<header ...><a ...>HexaScoreSpike</a><nav ...>`).
- `GET /up?checks=1` → `200 {"status":"ok","checks":{"db":"ok","ssr":"ok"}}`. After the SSR process is stopped: `503 {"status":"fail","checks":{"db":"ok","ssr":"fail"}}`, two page loads render an empty `#app` (client fallback), and `/up/ssr-fallbacks` → `{"fallbacks":2}`.
- `go test ./...` on spike `main` → all packages `ok`.

## P0.F.3.1 Signed keyset cursor with inertia.Scroll

Spike branch `agent/senior/P0.F.3.1` (8522ef6, from spike `main` a7c27a1). Files: `app/catalog/cursor.go` (+ `cursor_test.go`), `app/catalog/list.go`, `app/catalog/seed.go`, `database/seeders/items_seeder.go` (300 items, Postgres only), `app/controllers/items.go`, `routes/web.go` (`GET /items`), `resources/js/pages/Items/Index.tsx` (`InfiniteScroll`), `tests/items_cursor_test.go`.

Design: cursor is `base64url(json{s,v,id}).base64url(HMAC-SHA256)`, key from `App.DeriveKey("catalog-cursor", 32)`, compared with `hmac.Equal`; the sort key (with direction) is inside the signed payload, so a cursor from another sort fails. Query: `ORDER BY col, id` plus `WhereRaw("((col > ?) OR (col = ? AND id > ?))")` (`<` for descending), fetching `perPage+1` rows to know whether a next page exists. `col` comes only from the `catalog.Sorts` allowlist (`slug`, `-slug`, `type`, `-type`). Nulls-last is not covered (both columns are NOT NULL).

Env the tests need: `TEST_DATABASE_URL` (postgres admin URL, `pgtest.Database` creates a private database) for `tests/items_cursor_test.go`; `go test ./app/catalog/...` needs none.

Decisive output:

- `go test ./app/catalog/... -run TestCursor -v`: `TestCursorRoundTrip`, `TestCursorRejected` PASS (tampered payload, missing or empty signature, garbage, other sort, other direction and other key all give `ErrInvalidCursor`).
- `go test ./tests -run 'TestCursor|TestSecurity_A08' -v` PASS: `TestCursorPagesWithoutDuplicates` walks `/items` by Inertia partial visits (`X-Inertia-Partial-Component: Items/Index`, `X-Inertia-Partial-Data: items`) for the four sorts: 13 pages (12 of 24, then 12), 300 unique rows, ordered by `(sort value, id)` including the 75-way ties of `type`; `TestSecurity_A08_CursorTamperRejected` gets 400 `application/problem+json` with `code: invalid-cursor` for a tampered, cross-sort, reversed and garbage cursor. `go test ./...` `ok`; `bun run typecheck` clean.
- `ScrollMeta` with an opaque cursor is accepted: the page object has `scrollProps.items = {pageName: "cursor", previousPage: null, currentPage: null, nextPage: "<opaque token>", reset: false}` and `mergeProps: ["items.data"]`, built with `inertia.ScrollCursor(orm.CursorPage{...}, "cursor")`. This closes 07-4 and 09-4 with `Scroll`; the `WhenVisible` fallback is not needed. `previousPage` stays null (append-only scroll). The browser behaviour of `<InfiniteScroll data="items">` was not run (typecheck only): [NÃO VERIFICADO].
- Note: the items seeder returns early on non-Postgres dialects so the scaffold's SQLite seeder test keeps passing.

Verdict: works (workaround: app-owned HMAC cursor and hand-written problem body, VF-039; VF-005).

## P0.F.2.2 Go and Node SSR in one container

Verdict: works (Go supervisor in the app; no extra init system).

Spike branch `agent/orchestrator/P0.F.2.2`.

- `app/ssr/supervisor.go`: `Supervise(ctx, delay, "node", "bootstrap/ssr/ssr.js")` restarts Node one second after every exit and stops on SIGTERM/SIGINT; `cmd/app/main.go` starts it only for the server process (`len(os.Args) == 1`) when `SSR_SUPERVISE=true`, so `migrate` and other commands never spawn Node. `TestSuperviseRestartsUntilCancelled` passes. A small supervisor was chosen over `ssr:start` as a child because `ssr:start` exits 0 silently when the bundle does not listen (VF-038) and would add a second restart loop.
- SSR bundle made self-contained (`ssr.noExternal: true` in the starter plugin's SSR branch): it imports only Node built-ins, so the runtime needs no `node_modules`. Page components stay code-split under `bootstrap/ssr/assets/`; the whole `bootstrap/ssr` directory must be copied (copying only `ssr.js` produced `component-resolution` fallbacks, caught by the P0.F.2.1 counter).
- `Dockerfile`: assets stage runs `bun run build && bun run build:ssr`; runtime base `node:26.13.0-alpine3.24` (exact tag, latest LTS on 2026-10-29 per E-002) instead of `alpine:3.24.2`; `NODE_ENV=production SSR_SUPERVISE=true`; non-root user kept.

Acceptance (2026-10-29):

- `docker build -t hexascore-spike .` → exit 0; image size 298 MB (Node runtime is most of it).
- `docker run -p 3010:3000 ... hexascore-spike`, then `GET /up?checks=1` → `200 {"status":"ok","checks":{"db":"ok","ssr":"ok"}}`; `GET /` → server-rendered `#app` markup.
- `docker exec <id> pkill node` → `ssr: ok` again after 1.3 s (≤ 5 s), fallback counter 0.
- Boot needs the billing variables (`STRIPE_KEY`, `STRIPE_WEBHOOK_SECRET`, `STRIPE_PRICE_PRO`, `STRIPE_PRICE_BUSINESS`) because the P0.F.8.1 routes validate them at startup; test-mode values were used.

## P0.F.5.2 Migration advisory lock and concurrent index script

Spike branch `agent/senior/P0.F.5.2` (376aeb1). Files: `app/commands/migrate_locked.go` (+ `_test.go`), `bootstrap/commands.go` (registers `migrate:locked`), `scripts/concurrent-index.sh`, `database/concurrent/items_title_trgm.sql`. The runner exposes no lock hook, so the command wraps `Migrator().Up` in a session-level `pg_advisory_lock(hashtext('hexascore:migrate'))` taken on a dedicated connection (`db.SQL().Conn`), released on return (VF-040, VF-026). The index is `items_title_trgm` on `items.slug` because the spike `items` table has no title column.

Env: `go test ./app/commands/` needs `TEST_DATABASE_URL` (postgres admin URL, `pgtest.Database`). The CLI run needs `DATABASE_URL`. The script needs `DATABASE_URL` with `psql` on PATH, or `PSQL` (for example `PSQL='docker exec -i hexascore-spike-postgres-1 psql -U app -d spike_lock'`); optional `WATCH_LOCKS=1`.

Decisive output (fresh database `spike_lock`, two `migrate:locked` processes started together while a psql session held the same advisory lock for 6 s, so both visibly waited):

- `select count(*) from pg_locks where locktype='advisory' and not granted` printed `2` while both waited; after release: `exit1=0`, `exit2=0`; one process printed only `Nothing to migrate.`, the other printed the eight `Migrated:` lines (posts, users, auth tables, social accounts, two-factor, items, ratings, billing).
- `go test ./app/commands/ -v`: `TestMigrationLockSerializesRunners` (second runner does not enter for 500 ms while the first holds the lock, then does) and `TestMigrationLockReleasedOnError` PASS.
- With 400,000 extra rows in `items`: `WATCH_LOCKS=1 scripts/concurrent-index.sh items_title_trgm` printed `lock samples=5853 AccessExclusiveLock_samples=0 modes_seen={ShareUpdateExclusiveLock}` and `index items_title_trgm indisvalid=t`, exit 0. A server-side sampler (DO block, 20 ms interval) was needed because a `docker exec` per sample missed the build.
- INVALID leftover: after `update pg_index set indisvalid=false`, a rerun printed `dropping INVALID leftover items_title_trgm` and ended `indisvalid=t`. `go test ./...` `ok`.

Limits: the lock protects only runs through `migrate:locked`; plain `migrate` still races (the pre-deploy step must call the locked command); `migrate:status` does not list the concurrent index; the script was run through `docker exec`, not a local `psql` (not installed here); `shellcheck` was not available.

Verdict: workaround (lock and index script are app code; VF-040, VF-026, VF-017).

## P0.F.1.3 Better Auth compatibility hasher

Verdict: works (no framework change needed).

Spike branch `agent/orchestrator/P0.F.1.3`. vukem's `auth.NewSessionGuard` takes any `hash.Hasher` (`Make`, `Verify`, `NeedsRehash`), and the guard stores a fresh hash through the provider's `PasswordUpdater` when `NeedsRehash` is true after a successful login (<https://vukem.dev/docs/security/authentication> "Password hashing"). So no VF is needed.

- `app/accounts/compat_hasher.go`: `CompatHasher` wraps the scaffold's hasher (`hash.Default()`, or `hash.Fake()` in tests). Values matching `^[0-9a-f]{32}:[0-9a-f]{128}$` are verified with scrypt N=16384, r=16, p=1, dkLen=64 over `norm.NFKC` of the password, with the hex salt text as the salt and `subtle.ConstantTimeCompare`; `NeedsRehash` is always true for them. Everything else is delegated, and `Make` always produces argon2id.
- `app/accounts/accounts.go`: `hasher = NewCompatHasher(hasher)`.
- Dependencies: `golang.org/x/crypto v0.58.0` is now a direct dependency (Q-001, approved 2026-10-29). `golang.org/x/text v0.43.0` was already in the module graph.

Acceptance (2026-10-29):

- `go test ./app/accounts -run TestSecurity_A07_LegacyScryptVerifiesAndRehashes` → PASS. The test runs real HTTP sign-up and login. A wrong password redirects back to `/login` and leaves the stored Better Auth hash untouched. The fixture password logs in, and the stored value then starts with `$argon2id$`.
- `TestCompatHasherBetterAuthFixtures` → 3/3 fixtures verify (including the NFKC-equivalent input), and a wrong password or a malformed value fails.
- `go test ./...` on the spike → all packages `ok`.

## P0.F.6.1 Queue workers and scheduler inside serve

Verdict: workaround (works; per-family semaphore and idempotency table are app code, VF-018).

Spike branch `agent/orchestrator/P0.F.6.1`. `QUEUE_CONNECTION=database`, `QUEUE_NAMES=heavy,default`, `QUEUE_CONCURRENCY=8`, `JOBS_HEAVY_CONCURRENCY=2`, `SCHEDULE_TIMEZONE=America/Sao_Paulo`, `SHUTDOWN_TIMEOUT=15s`. The `jobs` and `failed_jobs` tables are created by vukem at boot (<https://vukem.dev/docs/digging-deeper/queue> "database").

- `app/jobs/recompute.go`: `spike.recompute` (heavy family, `queue.OnQueue("heavy")`, holds a semaphore slot for `Work`) and `spike.touch` (light). Both apply their effect through `applyOnce`, which inserts the idempotency key into `job_effects` (`ON CONFLICT DO NOTHING`) and runs the effect in the same transaction, so redelivery under at-least-once semantics changes nothing.
- `app/jobs/semaphore.go`: in-process per-family cap (buffered channel) with a `job_family_stats` row (`active`, `max_active`) as evidence. vukem queues are strict priority with one concurrency for all kinds (VF-018).
- `app/jobs/counter.go`: `BufferedCounter` for view counts. Each light job adds a hit; the `views:flush` task (`Cron("* * * * *")`, `WithoutOverlapping()`) writes one upsert per item. A failed write merges the counts back. Counts buffered at a crash are lost (accepted for views). Note: `HourlyAt` is documented on <https://vukem.dev/docs/digging-deeper/scheduler> "Frequencies" as of 2026-10-29, so VF-021 may be stale; cron is used anyway.
- Wiring: `bootstrap/app.go` builds one `jobs.Deps` and passes it to `WithJobsFunc` and `WithSchedule` through closures (no package globals). `jobs:demo -n -run -work` dispatches the test load.

Acceptance (2026-10-29, in the Linux container so `docker stop` delivers SIGTERM):

- Container log: `scheduler running inside serve; set SCHEDULE_IN_SERVE=false if you run schedule:work` (`"tasks":1`).
- 200 jobs (100 heavy, 100 light). `docker stop` after 4 s gave `shutting down cause="terminated signal received"`, with 46 effects applied and 154 jobs left. After restart: `jobs_left=0 failed=0 effects=200 distinct_keys=200 heavy=100 light=100 heavy_peak=2`.
- Hard kill: 40 more jobs with 3 s heavy work; `docker kill` (SIGKILL) after 5 s left 2 effects and 38 jobs. After restart, the expired reservations were redelivered in about 105 s, ending with `jobs_left=0 failed=0 r2_effects=40`. Every key was applied once. The stats row read `heavy_peak=4` because the two jobs killed mid-run never decremented `active`. That is a measurement artifact: the cap is enforced in process, and the graceful run plus `TestSemaphoreCapsConcurrency` show it holds at 2. The real app records concurrency as a metric (SDD 12), not as a row.
- `go test ./app/jobs/... -run TestBufferedCounterFlush` → PASS (with `-race`). `go test ./...` on spike `main` → all `ok`.

## P0.F.2.3 URL-prefix locales, i18n Once prop and theme cookie under SSR

Verdict: workaround (VF-024 and VF-028 stay open; refined in VF-041).

Spike branch `agent/senior/P0.F.2.3`. `vukem.WithTranslations(resources.Lang())` loads `resources/lang/en.json` and `pt-BR.json` (the roadmap says `lang/`; `resources/lang` is where the `WithTranslations` docs put catalogs). Design:

- `app/http/middleware/locale_prefix.go`: `LocalePrefix` (server-level, before routing) strips `/pt-br`, drops any client `_locale` and passes the locale as that query parameter; `i18n.Middleware` is configured with `Query: "_locale"`, `Order: [SourceQuery]`, so the URL is the only locale source (Accept-Language and the `lang` cookie no longer matter). A GET/HEAD without prefix whose `locale` cookie is `pt-BR` gets a 307 to the prefixed URL (not for Inertia visits, `/up`, `/api/`, `/build/`, file-like paths).
- `RestoreURL` (route-level, first in `r.Use`) puts the original URL back so Inertia's `url` keeps `/pt-br/` (Inertia uses `r.URL.RequestURI()`).
- `inertia.I18n("ui.")` shares `locale`, `locales` and the once prop `messages` keyed `i18n:<locale>`; `server.*` keys stay out of the page. Messages are flat dotted keys (`"ui.nav.home"`).
- Theme: view functions have no request access, so the cookie (`dark`|`light` only) is shared as prop `theme` and the root view renders `{{with .Page.Props.theme}} data-theme="{{.}}"{{end}}`.

Acceptance (spike built with `bun run build && bun run build:ssr`, `app serve` plus `app ssr:start`, port 3011 because 3001 was taken by another process):

- `/up?checks=1` → `{"status":"ok","checks":{"db":"ok","ssr":"ok"}}`.
- `curl -s localhost:3011/pt-br/ | grep -o '<html lang="[^"]*"'` → `<html lang="pt-BR"`; the SSR body contains the page and `"url":"\/pt-br\/"`.
- `curl -sI -H 'Cookie: locale=pt-BR' localhost:3011/` → `HTTP/1.1 307 Temporary Redirect`, `Location: /pt-br/`.
- `curl -s -H 'Cookie: theme=dark' localhost:3011/ | grep -c 'data-theme="dark"'` → `1`.
- Inertia visit `/pt-br/` returns `messages`; with `X-Inertia-Except-Once-Props: i18n:pt-BR` the props hold only `appName, auth, errors, locale, locales` (no `messages`), `onceProps` still lists `i18n:pt-BR`.
- `go test ./app/http/middleware/` ok (16 cases incl. open-redirect and client `_locale` spoof); `go test ./tests/ -run TestLocalePrefix` 4 tests PASS (prefix, cookie redirect, Once omission, theme incl. attribute-injection cookie). `go test ./...`: only the Postgres tests fail locally (no `TEST_DATABASE_URL` exported in this shell).

Limits: in-app links and `router.visit` must add the prefix themselves (no prefix-aware `route()`/URLFor, VF-024); `es` is not enabled (no app catalog yet); locale switching must set the `locale` cookie and navigate (not built); the 307 is per request, so a CDN must `Vary: Cookie` (set on the redirect only).

## P0.F.7.1 Two OpenAPI documents with RFC 9457 problem codes

Verdict: workaround (VF-014 and VF-016 confirmed; VF-042).

Spike branch `agent/senior/P0.F.7.1`. Typed handlers via `api.Register`: `routes/api.go` (`getItem`, plus the existing `ping`) and `routes/app_api.go` (`getAppItem`, `reportClientError` with 202). `app/http/problems/problem.go` adds `code` and extension members; `app/catalog/api.go` holds the item lookup (unknown type or malformed id is a 404 `not-found`).

- VF-014: payload fields are camelCase through `json` tags (`createdAt`), but the problem document keeps `request_id` and the envelope keys snake_case: `web.Problem` is fixed and has no naming option. Accepted as documented in the SDD README; no custom envelope written.
- VF-016: extension members are not supported by `web.Problem`/`DefaultErrorHandler`. `problems.Install` wraps the router-wide `Services.Errors` (group middleware cannot: handler errors do not pass through services swapped by middleware). The OpenAPI `Problem` schema does not list `code`, so generated clients will not see it until the framework adds extensions.
- Each document contains only its own prefix: `-prefix` keeps page routes out (no duplicate `operationId`).

Acceptance (2026-10-29):

- `app openapi:export -prefix /api/v1 > openapi/api-v1.json` and `-prefix /app/v1 > openapi/app-v1.json` → exit 0 / exit 0; both `openapi: 3.1.0`, operations `ping, getItem` and `reportClientError, getAppItem`.
- `curl -si -H 'Host: api.localhost' localhost:3011/v1/items/missing` → `404`, `Content-Type: application/problem+json`, body `{"code":"not-found",...,"request_id":"KKAZPBGJH6CHN62WIHAFU6SRYF",...}`.
- `go test ./tests -run TestAPIContract` → `ok` (both prefixes: 200, 404 `not-found`, 422 `validation-failed` with field errors, 202; every response passes `AssertMatchesOpenAPI()`); `go test ./app/http/problems/` ok (extension member, reserved-member guard, 5xx detail hidden, other paths untouched); `go test ./...` all `ok` with `TEST_DATABASE_URL` set.

Limits: the exported documents are checked in from the spike only as samples (no CI diff yet); `servers` shows `http://localhost:3000` because `API_URL` was unset; auth, quota and rate-limit headers are not wired (other tasks); `listItems` and the rest of SDD 05 section 6.2 are not implemented; unmatched routes (router 404) are not under the problem handler when the path is outside the installed prefixes.

## P0.F.10.1 Lighthouse CI budgets on the SSR container

Verdict: works, after one app change (gzip) and one environment mitigation.

Spike branch `agent/senior/P0.F.10.1`. `lighthouserc.json` (mobile emulation, Lighthouse default Moto G Power with slow 4G simulation, 3 runs, median) asserts on `http://localhost:3010/`: `largest-contentful-paint` <= 2000 ms, `cumulative-layout-shift` <= 0.05, `total-blocking-time` <= 200 ms, `resource-summary:script:size` <= 153,600 B (150 kB transfer). `@lhci/cli` 0.15.1 (npm `latest` on 2026-10-29), exact, dev-only; Chrome from the host (`CHROME_PATH`), reports in `.lighthouseci/` (gitignored). The target is the Docker image built from spike main (`hexascore-spike`, 298 MB) run as in P0.F.2.2.

Acceptance (2026-10-29):

- First run, container without compression: `bunx lhci autorun` exit 1: `resource-summary.script.size` found 568,467 (expected <= 153,600) and LCP 4,589 ms. Cause 1: `app-*.js` served raw (353,925 B). Fix: `r.UseServer(middleware.Gzip())` (`GET` with gzip: 112,898 B), VF-043.
- Cause 2 (environment): the host's Kaspersky web shield injects a 212 kB script, a stylesheet and polling XHRs from `gc.kis.v2.scr.kaspersky-labs.com` into every page, which inflated script size and delayed FCP/LCP. `blockedUrlPatterns: ["*kaspersky-labs.com*"]` in `lighthouserc.json` removes it (harmless on CI runners).
- Final run: `bunx lhci autorun` -> `All results processed!`, `Done running autorun.`, exit 0, 78 s for 3 runs after the container is up. Per run (LCP ms / TBT ms / CLS / script bytes): 1367 / 0 / 0 / 115,585; 1373 / 11 / 0 / 115,585; 1380 / 0 / 0 / 115,585. Performance score 1.

Sizing for P1.I.5.1: about 25 s per run, so ~80 s for 3 runs of one URL, plus image build (68 s cold in this run) and container start (~6 s). The page measured is the starter Home, not an item page: budgets on the real item page (images, fonts) are untested.

## P0.F.7.2 First-party client, typed routes and legacy SDK generators

Verdict: works for the first-party client and routes; the legacy SDK suites need a decision on operation names (below).

Spike branch `agent/senior/P0.F.7.2` (on spike main with P0.F.7.1). `app codegen:client -prefix /app/v1 -o packages/app-client/src/api.ts` and `app codegen:routes -o resources/js/routes.ts` both exit 0. `packages/app-client` has a `package.json` (private `@hexascore/app-client`) and a strict `tsconfig.json`. `codegen:types` (SDD 09 section 6.1) was not run: `[NOT IN VUKEM DOCS]` on the reference-runtime page; the `codegen:client` output already types props-shaped DTOs (`Item`, `Problem`), but Inertia page props are not covered by any generator.

Acceptance (2026-10-29):

- `bunx tsc --noEmit -p packages/app-client` -> exit 0, no errors; `bun run typecheck` (root, includes `resources/js/routes.ts`) -> exit 0.
- `routes.ts` lists named routes with method and path (`"app.v1.getAppItem": { method: "GET", path: "/app/v1/{type}/{id}" }`), including page routes.
- Legacy generators on temporary copies of `legacy/sdks/ts` and `legacy/sdks/python` (legacy untouched; copies in the session scratchpad, input pointed at `openapi/api-v1.json`): `@hey-api/openapi-ts` 0.99.0 `bun run sdk:generate` -> exit 0, `sdk.gen.ts` exports `ping` and `getItem`; `tsc --noEmit` passes. `openapi-python-client` 0.29.1 (installed with pip in a venv because `uv` is not on this machine; `uvx ruff` hook replaced by `ruff` 0.16.10) -> client generated with a warning for `ping` (`*/*` response).
- Legacy test suites fail as predicted: `bun test` -> `SyntaxError: Export named 'getGame' not found`; `pytest` -> `ModuleNotFoundError: No module named 'hexascore.api.games'`. Cause: the legacy contract names operations per type (`getGame`, `listGames`, `getBook`, `listBookGenres`, tags per type) while SDD 05 section 6.2 defines `getItem` and `listItems` over `/{type}`.

Decision needed (for the orchestrator): keep the legacy SDK surface by registering per-type typed routes (`GET /api/v1/games/{id}` named `getGame`, literal beats `{type}` in the mux) or accept a breaking SDK rename at v2 of the packages. Parity (D-20) points to the first.

Limits: only two public operations exist in the spike document, so the SDK regeneration is not a full compatibility check; response properties are all optional in generated types (VF-044); the Python run used pip, not `uvx`.

Orchestrator decision on P0.F.7.2 (2026-10-29): SDD 05 §6.2 is final and names the operations `listItems`/`getItem` over `/v1/{type}`, so the legacy per-type SDK names (`getGame`, `listGames`, ...) are not kept. The SDKs get a new major version; their tests are rewritten against the new document in P1.H.4.

## P0.F.9.1 Storage bucket

Spike branch `agent/senior/P0.F.9.1` (b2113ce, on spike main 783ad2f). RustFS `rustfs/rustfs:1.0.4` added to the spike `compose.yaml`; `drivers/s3` v0.3.1 (vukem module) via `config.Storage(app)` (`S3_*` env, path style, insecure HTTP for local only); `app/covers/serve.go` serves `GET /covers/{key}`; `app/exports/zip_job.go` is the queued `exports.build` job (database queue, `QUEUE_CONNECTION=database`).

Acceptance (2026-10-29), `go test ./tests -run "TestCoversServe|TestPresignExpires|TestExportJob|TestZipCSV" -v` against the running RustFS and Postgres:

- `TestCoversServe` PASS: `/covers/42-0123456789abcdef.webp` -> 200, `Cache-Control: public, max-age=31536000, immutable`, `X-Content-Type-Options: nosniff`; missing key, non-hash name and `.png` -> 404 without Cache-Control.
- `TestExportJobWritesZip` PASS: dispatch `exports.Build`, `q.Work(StopWhenEmpty)` writes `exports/<user>/<id>.zip`; a presigned GET downloads a valid ZIP whose `profile.csv` holds the guarded cell `'=cmd|calc`.
- `TestPresignExpires` PASS: presigned GET (2 s) -> 200 with the body; after 3 s -> 403. Put, Exists and Open also exercised.
- `TestZipCSVMatchesRecorder` PASS: the VF-029 choice. The job uses `encoding/csv` plus exported `export.SafeCell`; its bytes equal `export.CSV` through an `httptest.ResponseRecorder`. The re-implementation is smaller (no request, no recorder) and needs no extra code beyond the guard.

Limits: the literal `curl -sI localhost:3000/covers/...` was proven through the in-process test client, not a running server; the covers route resolves the disk lazily so other routes and tests boot without `S3_*` (production must fail at boot or in a readiness check, to do in P1); the export reads only name and email (the real job per SDD 14 reads the full data set in a repeatable-read transaction); `storage.Handler` gaps filed as `VF-045`.
