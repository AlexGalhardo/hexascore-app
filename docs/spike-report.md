# Phase 0 spike report

Findings from the throwaway spike app (`C:\dev\hexascore-app\hexascore-spike`, local git, never pushed). Each microtask appends a section; P0.F.11.1 consolidates them. Verdict values: works, workaround, blocked.

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

## P0.F.4.1 Host routing for api.<host>/v1

Spike branch `agent/senior/P0.F.4.1` (commit 6570235). Files: `app/http/middleware/api_host.go` (+ `_test.go`), `routes/api.go` (`/api/v1/ping` group and `r.UseServer(middleware.APIHost(API_HOST))`), `routes/registry.go`, `bootstrap/app.go` (`vukem.WithOpenAPI(openapi.Info{Servers: ...})` from `API_URL`), `tests/api_host_test.go`. The server middleware runs before the mux (`router.build`: `r.server` wraps `dispatch`), so it rewrites `/v1/` to `/api/v1/` before route matching; on the API host any other path (`/`, `/login`, `/api/v1/*`, `/v10/*`) is 404.

Env vars the tests need: none for `go test ./app/http/middleware/` and `go test ./tests/ -run APIHost` (in-memory SQLite, `API_HOST` is set through `vukemtest.Options.Env`). The full `go test ./...` also needs `TEST_DATABASE_URL` (postgres admin URL) for `TestRegisterOnPostgres`. Runtime env: `API_HOST=api.localhost`, `API_URL=http://api.localhost:3001/v1` (process env, for OpenAPI servers), `PORT=3001`.

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

- `bootstrap/app.go`: `inertia.SSRConfig{Bundle: "bootstrap/ssr/ssr.js", OnFailure: ssrFailures.Observe}`; `app.AddHealthCheck("ssr", app.Inertia.SSRHealthy)` adds `ssr` to `GET /up?checks=1` (https://vukem.dev/docs/deployment/single-binary "Health checks"); `GET /up/ssr-fallbacks` exposes the counter for the spike only (production exports a metric, SDD 12).
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

Spike branch `agent/orchestrator/P0.F.1.3`. vukem's `auth.NewSessionGuard` takes any `hash.Hasher` (`Make`, `Verify`, `NeedsRehash`), and the guard stores a fresh hash through the provider's `PasswordUpdater` when `NeedsRehash` is true after a successful login (https://vukem.dev/docs/security/authentication "Password hashing"). So no VF is needed.

- `app/accounts/compat_hasher.go`: `CompatHasher` wraps the scaffold's hasher (`hash.Default()`, or `hash.Fake()` in tests). Values matching `^[0-9a-f]{32}:[0-9a-f]{128}$` are verified with scrypt N=16384, r=16, p=1, dkLen=64 over `norm.NFKC` of the password, with the hex salt text as the salt and `subtle.ConstantTimeCompare`; `NeedsRehash` is always true for them. Everything else is delegated, and `Make` always produces argon2id.
- `app/accounts/accounts.go`: `hasher = NewCompatHasher(hasher)`.
- Dependencies: `golang.org/x/crypto v0.58.0` is now a direct dependency (Q-001, approved 2026-10-29). `golang.org/x/text v0.43.0` was already in the module graph.

Acceptance (2026-10-29):

- `go test ./app/accounts -run TestSecurity_A07_LegacyScryptVerifiesAndRehashes` → PASS. The test runs real HTTP sign-up and login. A wrong password redirects back to `/login` and leaves the stored Better Auth hash untouched. The fixture password logs in, and the stored value then starts with `$argon2id$`.
- `TestCompatHasherBetterAuthFixtures` → 3/3 fixtures verify (including the NFKC-equivalent input), and a wrong password or a malformed value fails.
- `go test ./...` on the spike → all packages `ok`.

## P0.F.2.3 URL-prefix locales, i18n Once prop and theme cookie under SSR

Verdict: workaround (VF-024 and VF-028 stay open; refined in VF-NEW-P0.F.2.3).

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
