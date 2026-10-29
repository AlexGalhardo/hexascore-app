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

- `users`: id, name, email (unique), password, email_verified_at (null), timestamps (a line between email_verified_at and timestamps was not inspected: [NÃO VERIFICADO]); `add_two_factor_columns` adds two_factor_secret (text, orm.Encrypted), two_factor_recovery_codes (text, encrypted SHA-256 hashes), two_factor_confirmed_at.
- `remember_tokens`: id (unique), user_id (index), hash, expires_at.
- `access_tokens`: id (unique), user_id (index), name, hash, abilities (text), last_used_at, expires_at, [column at line 28 not inspected], revoked_at, created_at.
- `password_resets`: email (unique), hash, created_at.
- `social_accounts`: id, user_id (FK users, cascade), provider, provider_id (unique together), timestamps.
- There is no sessions table in the scaffold (`SESSION_DRIVER` default is not `database`).

Pages: React `.tsx` (TypeScript) under `resources/js/pages/Auth/`: Login, Register, CheckEmail, ForgotPassword, ResetPassword, VerifyEmail, TwoFactor, TwoFactorChallenge; `Dashboard.tsx` updated. The pages do not set a layout; they reuse the starter `layouts/AppLayout.tsx` through the default layout in `resources/js/app.tsx`. No Google-specific page: social is route-only.

Routes: GET/POST `/login`, `/register`, GET `/register/check-email`, POST `/logout`, GET `/forgot-password`, POST `/forgot-password`, GET `/reset-password/{token}`, POST `/reset-password`, GET `/email/verify`, `/email/verify/{id}`, POST `/email/verification-notification`, GET `/dashboard`, GET `/auth/{provider}/redirect`, GET/POST `/auth/{provider}/callback`, GET `/user/two-factor`, POST `/user/two-factor`, `/user/two-factor/confirm`, `/user/two-factor/recovery-codes`, `/user/two-factor/disable`, GET/POST `/two-factor-challenge`. Google needs `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET` (dummy in spike `.env`).

Session env names (open issue 03-7): `SESSION_LIFETIME` (absolute), `SESSION_IDLE_TIMEOUT`, plus `SESSION_DRIVER`, `SESSION_COOKIE`, `SESSION_SECURE`, `SESSION_RECORD_CLIENT` (found in framework source `session/`; value formats [NÃO VERIFICADO]).

Verdict: works, with one workaround: generate `en,pt-BR` and hand-translate `es` catalogs (VF-032).
