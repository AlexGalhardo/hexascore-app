# SDD 11: Infrastructure, deployment and data operations

Status: reviewed (2026-10-28 consistency pass). Decisions: D-07, D-08, D-18, D-19, D-20, D-21, A-03, A-04. ADRs: 0002, 0003, 0004, 0007, 0013, 0014. Related SDDs: 03, 05, 06, 07, 12, 13.

Sources, accessed 2026-10-28. Railway (`https://docs.railway.com/…`): [pricing] `reference/pricing/plans`, [buckets] `guides/storage-buckets`, [postgres] `guides/postgresql`, [health] `reference/healthchecks`, [network] `reference/private-networking`, [regions] `reference/deployment-regions`, [cac] `reference/config-as-code`, [cli] `reference/cli-api`, [predeploy] `guides/pre-deploy-command`, [cron] `reference/cron-jobs`. vukem (`https://vukem.dev/docs/…`): [vk-deploy] `deployment/docker-and-railway`, [vk-bin] `deployment/single-binary`, [vk-config] `configuration`, [vk-cli] `cli/reference-runtime`, [vk-db] `database/overview`, [vk-mig] `database/migrations`, [vk-seed] `database/factories-seeders`, [vk-queue] `digging-deeper/queue`, [vk-sched] `digging-deeper/scheduler`, [vk-storage] `digging-deeper/storage`, [vk-flags] `digging-deeper/feature-flags`, [vk-otel] `digging-deeper/opentelemetry`, [vk-crypt] `security/encryption`, [vk-inertia] `frontend/inertia`. No site page exists for release tooling, git hooks, Windows caveats and `contrib/engine`: cited as `vukem/docs/tooling.md` and `vukem/docs/subsystems.md` "External binaries" [SITE PAGE MISSING].

## 1. Objective

One container plus one managed Postgres on Railway for ≤ US$30/month (D-08): deploys that drop no requests, migrations that never block the catalogue, backups restored monthly (D-18), vukem-style releases (D-20) and a cutover from Vercel/Neon reversible for 30 days (D-21), served from the region nearest Brazil (A-03).

## 2. Scope

In: topology, image, env/secrets, migrations, backups, cutover, CI/CD, local development, costs, SDD 12 hooks. Out (deferred to the first budget increase, D-08): replicas, separate worker, Redis (ADR 0003), WAL archiving (ADR 0014), multi-region, CDN.

## 3. Functional requirements

Each is verified by the matching §9 item.

- FR-11-1 One project, environments `production` and `staging`, services `app` and `postgres`, one bucket per environment.
- FR-11-2 Image from pinned bases (tag + digest), non-root, assets embedded.
- FR-11-3 Traffic switches only after `/up?checks=1` is 200 (db, ssr, storage).
- FR-11-4 Migrations run once per deploy, before traffic, under an advisory lock; failure aborts the deploy.
- FR-11-5 Daily encrypted, verified `pg_dump`; retention 30 daily + 12 monthly.
- FR-11-6 Monthly restore drill on staging ≤ 60 min, asserts green.
- FR-11-7 Cutover imports legacy data; every verification query passes.
- FR-11-8 Branch model (amended 2026-10-28 for the orchestration prompt): `dev` is the integration branch and deploys staging on every merge; `main` receives `dev` by PR when a task group closes and a tag `vX.Y.Z` on `main` deploys production; responses carry the version.
- FR-11-9 `docker compose up` + `vukem serve` work with a deterministic seed on Windows 11 and Linux (CI on `windows-2025`, `ubuntu-24.04`).

## 4. Non-functional requirements

- NFR-11-1 Cost ≤ US$30/month at launch (D-27): Railway e-mail alert at US$24, hard usage limit US$30; SDD 12 adds a "projected spend at day 20" alert so the owner can raise the limit before a legitimate spike takes the site down (§6.5).
- NFR-11-2 0 failed requests in a k6 run (20 rps) across a deploy; RPO 24 h, RTO 1 h (ADR 0014).
- NFR-11-3 Image ≤ 400 MB compressed; boot to ready ≤ 30 s.
- NFR-11-4 HTTP priority via SDD 06 NFR-06-1/2 (≤ 2 ingest jobs, pressure gate).
- NFR-11-5 No secret in git, image or logs (`.dockerignore` [vk-deploy], `config.Secret` [vk-config]).
- NFR-11-6 Cutover read-only window ≤ 60 min; DNS TTL 60 s from T-24 h.

## 5. Data model

Postgres major **18** (vukem Compose pins `postgres:18.6-alpine3.24` [vk-deploy], as `legacy/infra/docker-compose.yaml`), minimum 16 (SDD 07). Railway's image is based on the official one [postgres]; its major, `pg_trgm`/`unaccent` (legacy `0000_extensions.sql`) and `ALTER SYSTEM` are [NOT VERIFIED]. Tuning: `shared_buffers=256MB`, `work_mem=8MB`, `maintenance_work_mem=256MB`, `max_connections=60`.

Sizing: 2 GB data + 600 MB indexes (SDD 07) + 1 GB WAL ≈ 3.6 GB; alert at 70 % of the volume. Pool: `DB_MAX_OPEN=20`, `DB_MAX_IDLE=10` (defaults 10/5 [vk-db]; ingestion needs `DB_MAX_OPEN` >= 2, because a step transaction and the `cover_gc` reservation hold two connections, and the ingest runner refuses to build below that) for ~14 HTTP, 3 workers, scheduler; pre-deploy and `pg_dump` add ≤ 5.

```sql
CREATE TABLE backup_runs (
  id uuid PRIMARY KEY,                 -- UUID v7
  kind text NOT NULL CHECK (kind IN ('daily','monthly','drill','offsite')),
  object_key text,                     -- backups/postgres/daily/2026/10/28/hexascore-production-20261028T0900Z.dump.enc
  bytes bigint, sha256 text,
  status text NOT NULL CHECK (status IN ('running','uploaded','verified','failed')),
  duration_s integer, error text,
  started_at timestamptz NOT NULL, finished_at timestamptz
);
```

Bucket prefixes: `covers/<id>-<hash16>.webp` (SDD 06 §7.5), `exports/` (SDD 05), `backups/postgres/{daily,monthly}/`, `backups/legacy/`.

Not backed up: Railway logs (SDD 12), staging data (re-copied), covers (re-derivable by ingestion, issue 1), Railway variables (owner keeps an encrypted export). Sessions and cache are dumped but worthless after a restore.

## 6. Contracts

### 6.1 Topology

| Service | Source | Cap | Notes |
| --- | --- | --- | --- |
| `app` | Dockerfile, 1 replica | 2 vCPU / 2 GB | `serve` = HTTP + queue + scheduler (ADR 0007) + SSR child |
| `postgres` | Railway Postgres + volume | 1 vCPU / 1 GB | private only, no TCP proxy |
| bucket `hexascore-<env>` | Railway bucket, app region | — | private, S3 API [buckets] |

Plan **Pro**: it includes US$20 of usage (Hobby US$5) [pricing], so at ~US$27 usage both bills match, and Hobby caps volumes at 5 GB per service [pricing]. Region: Railway has US West, US East (`us-east4-eqdc4a`), EU West, Southeast Asia, nothing in South America [regions] → US East (A-03). Measurement: throwaway `/up` services in `us-east4-eqdc4a` and `us-west2`; 50 `curl -w '%{time_connect} %{time_starttransfer}'` from the owner's line and 3 Brazilian probes; p50/p95 in an ADR 0002 addendum.

Networking: private network ("encrypted Wireguard tunnels using internal DNS" [network]), `DATABASE_URL=${{postgres.DATABASE_URL}}` [vk-deploy]. Domains `hexascore.app`, `www` (301 in app), `api`; DNS on Cloudflare (D-28: zone hosted or moved there before cutover) with CNAME flattening at the apex pointing to the Railway domain, proxy off at launch (available later as the cover CDN); TLS by Railway; host routing in middleware (VF-015); allowed hosts include `healthcheck.railway.app` [health].

### 6.2 Railway configuration

vukem states `railway.json`/`railway.toml` stop being read on 2026-12-01 and generates `.railway/railway.ts` (`vukem new --deploy railway`) [vk-deploy]; Railway's CLI has `railway config init|plan|apply` for it [cli]. HexaScore commits:

```ts
const db = postgres("postgres");
const app = service("app", {
  healthcheck: "/up?checks=1",
  preDeploy: "/app/hexascore hexascore:predeploy",
  env: { APP_KEY: preserve(), DATABASE_URL: db.env.DATABASE_URL, TRUSTED_PROXIES: "100.64.0.0/24",
         CLIENT_IP_HEADER: "X-Real-IP", TRUST_FORWARDED_PROTO: "true", OTEL_METRICS_EXPORTER: "none" /* … §6.4 */ },
});
```

Overlap 20 s, draining 35 s, restart `ALWAYS` max 5 are documented settings [cac] (IaC keys [NOT VERIFIED]; service settings meanwhile). Commands are the binary path plus arguments, no shell [vk-deploy]. The healthcheck only gates a deploy, "not used for continuous monitoring" [health], so using readiness there cannot restart a healthy process (the concern in [vk-bin]); SDD 12 adds an external probe.

### 6.3 Container

Image as built (P2.A.3.1, 2026-10-29): assets stage `oven/bun:1.4.3-alpine` (the starter's, kept instead of `node` plus a pinned Bun binary), build stage `golang:1.27.3-alpine3.24`, runtime `node:26.13.0-alpine3.24` (latest LTS on the day, replacing the planned 24.22.0), every `FROM` pinned by digest; runtime packages `postgresql18-client=18.6-r0` and `libwebp-tools=1.6.0-r0` (both present on Alpine 3.24); non-root user `app`; `scripts/infra/check-image.sh` enforces ≤ 400 MB compressed (73 MB measured), user `app`, pinned bases, `pg_dump` 18.x, `cwebp`, and `/app/storage` plus `/tmp` as the only writable paths.

Stages, starting from the pins of the vukem starter Dockerfile [vk-deploy], each with `@sha256:` digest (Dependabot `docker`):

1. `node:24.22.0-alpine3.24` with a pinned Bun binary (the repo uses `bun.lock`, SDD 09 §7.10): `bun install --frozen-lockfile`, `bun x vite build`, then the permanent `bun run build:ssr` step for `bootstrap/ssr/ssr.js`, because `vukem build` does not build the SSR bundle (VF-027, open issue 09-7).
2. `golang:1.27.3-alpine3.24`: `CGO_ENABLED=0 go build -trimpath -tags "$(cat .buildtags)" -ldflags "-s -w -X main.version=$VERSION -X main.commit=$COMMIT" -o /out/hexascore ./cmd/app`, assets embedded (same as the starter; extra `-X` through `vukem build` is [NOT IN VUKEM DOCS]).
3. Runtime `node:24.22.0-alpine3.24` (Node is needed at runtime for SSR, ADR 0004) + `ca-certificates`, `tzdata`, `postgresql18-client`, `libwebp-tools` (`cwebp`, SDD 06) at pinned apk versions [NOT VERIFIED availability]; user `app` uid 10001, writable `/app/storage` and `/tmp` only; `HOST=0.0.0.0`, `LOG_FORMAT=json`, `ENTRYPOINT ["/app/hexascore"]`, `CMD ["serve"]` (starter defaults [vk-deploy]).

Process: `serve` runs the SSR server as a child with `ssr:start` behaviour (restart after `-restart-delay 1s` [vk-cli]); a built-in `serve` runner for it is [NOT IN VUKEM DOCS] in v0.32.0 (VF-027). As built (P2.A.3.2): `cmd/app` supervises Node (`app/ssr/supervisor.go`) only for the bare server command with `SSR_SUPERVISE=true`; SIGTERM reaches the Go process only, the HTTP drain runs within `SHUTDOWN_TIMEOUT`, then the child gets SIGTERM and SIGKILL after 5 s, so requests in flight still render. `GOMEMLIMIT` caps Go; Node inherits `NODE_OPTIONS`.

Health: `/up` liveness, `/up?checks=1` readiness (503 with per-check `ok`/`fail` [vk-bin]), each check 2 s (`bootstrap.ReadyCheckTimeout`): `db` (framework Ping re-registered with 2 s), `ssr` (`/health`, as `ssr:check` [vk-cli]), `storage` (`storage.Check` round trip [vk-storage]; on the local exports disk until the bucket of P2.A.2.1 exists), `scheduler` (in-memory heartbeat beaten by `s.Func("heartbeat").EveryMinute()`, age < 180 s, seeded at boot; registered only when `SCHEDULE_IN_SERVE=true`). Queue order across `QUEUE_NAMES` observed on v0.32.0's database driver: strict priority by position (`TestQueue_ObservedPriorityOrder`); the app does not depend on it. Shutdown: `SHUTDOWN_TIMEOUT=30s` (default 15 s [vk-deploy]) for requests, jobs, tasks; interrupted jobs re-run, so handlers are idempotent [vk-queue]; draining 35 s > 30 s.

### 6.4 Environment variables

Railway service variables per environment; `.env` locally; `.env.example` lists every key with empty secrets. Invalid config stops the boot with every problem listed [vk-config], [vk-deploy] "What stops the boot".

| Name | Value / example | Notes |
| --- | --- | --- |
| `APP_NAME`, `APP_ENV`, `APP_URL` | `HexaScore`, `production`, `https://hexascore.app` | staging: `APP_ENV=production`, own URL |
| `APP_KEY`, `APP_PREVIOUS_KEYS` | `base64:` ≥ 32 bytes | secret; `docker run --rm hx key:generate`; per environment; HKDF sub-keys, rotation [vk-crypt], [vk-config] |
| `PORT`, `HOST` | injected, `0.0.0.0` | [vk-deploy] |
| `DATABASE_URL`, `DB_MAX_OPEN`, `DB_MAX_IDLE` | reference, `20`, `10` | §5 |
| `TRUSTED_PROXIES`, `CLIENT_IP_HEADER`, `TRUST_FORWARDED_PROTO` | `100.64.0.0/24`, `X-Real-IP`, `true` | observed range, check `peer` in logs [vk-deploy] |
| `SESSION_DRIVER`, `CACHE_STORE`, `RATE_LIMIT_STORE` | `database`, `database`, `cache` | SDD 03 |
| `QUEUE_CONNECTION`, `QUEUE_NAMES`, `QUEUE_CONCURRENCY` | `database`, `default,ingest`, `3` | worker inside `serve` [vk-queue] |
| `SCHEDULE_IN_SERVE`, `SCHEDULE_TIMEZONE` | `true`, `America/Sao_Paulo` | [vk-sched] |
| `SHUTDOWN_TIMEOUT`, `LOG_FORMAT` | `30s`, `json` | |
| `S3_BUCKET`, `S3_REGION`, `S3_ENDPOINT`, `S3_ACCESS_KEY_ID`, `S3_SECRET_ACCESS_KEY`, `S3_USE_PATH_STYLE` | bucket Credentials tab [buckets] | secret; `drivers/s3` [vk-storage] |
| `BACKUP_KEY` | 32 bytes base64 | secret, separate from `APP_KEY`, offline copy with the owner |
| `BACKUP_SOURCE_S3_*` | production bucket credentials | staging only, drill |
| `API_HOST`, `ADMIN_EMAILS` | `api.hexascore.app`, owner | |
| `MAIL_MAILER` + provider key | `resend` | must be set in production [vk-deploy] |
| `GOOGLE_*`, session timeouts | per SDD 03 | names [NOT VERIFIED] (SDD 03 issue 7) |
| `STRIPE_SECRET_KEY`, `STRIPE_WEBHOOK_SECRET` | live / test on staging | SDD 05 |
| `TMDB_READ_ACCESS_TOKEN`, `IGDB_TWITCH_CLIENT_ID`, `IGDB_TWITCH_CLIENT_SECRET` | | secret (legacy names shortened) |
| `OTEL_*` | injected by Railway tracing; `OTEL_METRICS_EXPORTER=none` | Railway receiver is trace-only [vk-deploy], [vk-otel] |
| `FLAG_PIX_CHECKOUT`, `FLAG_PUBLIC_API_SIGNUP`, `FLAG_SEED_DATA` | `false`, `true`, `false` | names must be lowercase + `_` [vk-flags], so `pix-checkout` → `pix_checkout` |
| `GOMEMLIMIT`, `NODE_OPTIONS` | `900MiB`, `--max-old-space-size=384` | |

Dropped (`legacy/apps/web/lib/env.ts`): `BETTER_AUTH_*`, `UPSTASH_*`, `KV_*`, `BLOB_READ_WRITE_TOKEN`, `CRON_SECRET`, `VERCEL_ENV`, `NEXT_PUBLIC_*`. Legacy crons (`legacy/apps/web/vercel.json`: 00:15, 12:00, 03:00) become scheduler tasks of SDDs 05 and 03. Staging: Stripe test keys, `noindex`, mail allowlist, ingestion on demand.

### 6.5 Cost sheet

Prices [pricing], [buckets]: RAM US$10/GB-month, CPU US$20/vCPU-month, egress US$0.05/GB, volume US$0.15/GB-month, bucket US$0.015/GB-month with free egress and operations; builds free. Averages are estimates (issue 13).

| Item | Assumption | US$ |
| --- | --- | --- |
| prod `app` | 0.25 vCPU, 0.45 GB | 9.50 |
| prod `postgres` | 0.10 vCPU, 0.45 GB, 6 GB volume | 7.40 |
| staging `app` | 0.03 vCPU, 0.30 GB | 3.60 |
| staging `postgres` | 0.02 vCPU, 0.25 GB, 6 GB volume | 3.80 |
| buckets | covers 3 GB × 2 + backups 13 GB | 0.29 |
| egress | 60 GB (pages + proxied covers) | 3.00 |
| **Bill (Pro)** | max(20, usage) | **≈ 27.6** |

Alerts and limit (D-27): Railway "Custom email alert" at US$24 and hard limit US$30 [pricing]; the day-20 projected-spend alert is owned by SDD 12 §7.9. Covers are served by the app with `Cache-Control: public, max-age=31536000, immutable` (buckets are private [buckets]); when cover egress exceeds US$5/month (SDD 12 cost review) the CDN decision is triggered (Cloudflare proxy on the cover host, D-28). Above US$24: stop staging `app` between releases (−3.6). Growth (5× traffic, 300k items): app 20 + Postgres 15.8 + egress 300 GB 15 + staging 7.4 ≈ US$58. First purchases: (1) `worker` service (`queue:work`, `schedule:work`; `QUEUE_IN_SERVE=false`, `SCHEDULE_IN_SERVE=false` [vk-sched], ADR 0007); (2) Redis before a second replica (ADR 0003); (3) WAL archiving, RPO ≤ 1 h (ADR 0014).

### 6.6 CI/CD (GitHub Actions on `AlexGalhardo/hexascore-app`)

GitHub Actions is available for `hexascore-app` (D-26; legacy had disabled them 2026-10-03, `legacy/docs/process.md`); the full pipeline runs there. The port of `legacy/scripts/ci-local.sh` on lefthook `pre-push` (`vukem/docs/tooling.md` [SITE PAGE MISSING]) is kept as a convenience mirror.

| Job | Trigger | Content | Required |
| --- | --- | --- | --- |
| `lint` | PR | golangci-lint v2 run + fmt diff, `go vet`, Biome (`biome ci`, D-22) + `tsc --noEmit`, markdownlint, commitlint on PR commits and PR title (D-37; the local `.githooks/commit-msg` stays as the offline check) | yes |
| `test` | PR | `go test ./...` and `-tags=integration` with `postgres:18.6-alpine3.24` service | yes |
| `vuln` | PR, weekly | `govulncheck ./...` pinned | yes |
| `openapi` | PR | `openapi:export` diff vs committed spec; Schemathesis vs built container | yes |
| `build` | PR | image build, ≤ 400 MB, no unpinned `FROM` | yes |
| `e2e` | PR | Playwright + axe vs built container | yes |
| `lighthouse` | PR | ADR 0004 budgets | yes |
| `smoke` | staging deploy | k6 smoke `p(95)<500`, 0 errors | gates release |
| `zap` | weekly | ZAP baseline vs staging | issue on failure |
| `release` | tag `v*`, `workflow_dispatch(version)` | notes from `CHANGELOG.md`, production deploy | — |

`openapi`, `e2e` and `lighthouse` download `image.tar` from `build` and do not run `docker build`. `openapi` exports from that image (`scripts/openapi-check.sh --image`) and fuzzes it (`qa/schemathesis/ci.sh`). `e2e` and `lighthouse` start it with `scripts/ci/qa-up.sh`: compose profile `app`, `--no-build`, DemoCatalog seed. Playwright runs `bun run e2e` (3 locales, desktop and mobile, axe in the specs) and uploads traces on failure. Lighthouse runs `bunx --bun @lhci/cli@0.15.1 autorun` against the five URLs in `lighthouserc.cjs`. Neither job calls staging.

Actions pinned by SHA. Staging: Railway GitHub autodeploy of `main` [vk-deploy]. Production: `release.yml` checks out the tag and runs `railway up --detach` [cli] with a project token (flags [NOT VERIFIED]). Version: `VERSION=$GITHUB_REF_NAME` build arg → `X-Hexascore-Version` header and `about` (inside `/up` [NOT IN VUKEM DOCS]).

## 7. Behaviour and algorithms

### 7.1 Release (ADR 0013; `vukem/docs/tooling.md` "Release process", "Tag pushes" [SITE PAGE MISSING])

1. `main` green, staging smoke green, migration rehearsal done (§7.2).
2. `git-cliff --config .config/cliff.toml --unreleased --tag vX.Y.Z` (preview); `scripts/release.sh vX.Y.Z` (POSIX shell, requires git-cliff; no Task) replaces the body of `[Unreleased]` with the generated section, commits `chore(release): vX.Y.Z`, tags. The range starts after the newest first-parent commit whose tree equals the highest tag's tree (squash releases leave tags off `main`); `--since <rev>` overrides it, and the compare links at the bottom are updated ([release-range runbook](../runbooks/release-range.md)).
3. `git push origin main`; `git push origin vX.Y.Z` (one tag per push); `gh run list --workflow release.yml`.
4. Release body from the literal `awk` heading match; empty body fails.
5. `curl -sI https://hexascore.app/up` shows the version; dashboards clean 15 min.

Rollback: `gh workflow run release.yml -f version=vPREV` (schema stays compatible, §7.2).

### 7.2 Migrations (ADR 0014)

- Forward-only in production (`down` tested, never run there). Expand → migrate (backfill job, SDD 06 §7.8) → contract two releases later. Each migration runs in its own transaction, rolled back on failure on Postgres [vk-db].
- Runner: Railway pre-deploy command `/app/hexascore hexascore:predeploy` (`app/console/predeploy.go`, P2.E.1.1; vukem's Railway setup uses `/app/app migrate` [vk-deploy]); on failure "the deployment will not proceed" [predeploy], so the old version keeps serving and new code never boots on an unmigrated schema (run-at-boot would migrate while the old container still serves during overlap). It drops `INVALID` index leftovers of a failed concurrent build (current schema, not partitioned parents, none being built per `pg_stat_progress_create_index`), then runs `migrate`. Since vukem v0.33.0 the runner holds its own `pg_advisory_lock` for the whole run (VF-026), so overlapping runs serialise and the second finds nothing pending; `migrate:locked` (P1.A.4.3) is redundant and retired by ADOPT-VF-040. Switching the Railway pre-deploy command is the owner's deploy change (runbook `docs/runbooks/migrations.md`).
- Concurrent indexes (VF-017, native since vukem v0.33.0): a one-statement migration with `NoTransaction: true` and `CREATE INDEX CONCURRENTLY IF NOT EXISTS` (Down `DROP INDEX CONCURRENTLY IF EXISTS`), recorded by `migrate:status` like any other. A failed build leaves an `INVALID` index that `IF NOT EXISTS` would skip on retry, which is why `hexascore:predeploy` drops leftovers first (VF-054). Big index builds ship alone in their release (pre-deploy timeout, §8). `database/concurrent/` and `scripts/concurrent-index.sh` (P0.F.5.2 workaround) are not used.
- Rehearsal before every release with a migration: restore the latest production dump into staging, deploy the candidate, record duration and lock waits; > 5 s `ACCESS EXCLUSIVE` blocks the release.

### 7.3 Backups

1. Task `backup.daily` 06:00 America/Sao_Paulo, `WithoutOverlapping()` [vk-sched], after SDD 06 asserts.
2. `pg_dump -Fc -Z 6 --no-owner` via `os/exec` (`app/backup/dump.go`; `contrib/engine` `RunStream` is [NOT IN VUKEM DOCS] on the site, so the documented fallback of `docs/PROGRESS.md` applies); the child gets only `PGHOST`, `PGPORT`, `PGUSER`, `PGPASSWORD`, `PGDATABASE`, `PGSSLMODE` from `DATABASE_URL`, never the app environment.
3. Encrypt in 4 MiB frames with `crypt.New(BACKUP_KEY)` `Seal(frame, "backup:<key>:<n>")` [vk-crypt] (`<key>` is the object key; the last frame's additional data ends in `:last`; one base64url token per line), so frames cannot be reordered, moved to another object, dropped or truncated unnoticed (`app/backup/crypt.go`); SHA-256 over the ciphertext; multipart `Put` through `drivers/s3` [vk-storage]. `backup:run --kind daily|monthly` (P2.E.2.1) records `running`, then `uploaded` or `failed` (scrubbed error), and deletes a partial object on failure. Railway buckets have no server-side encryption [buckets].
4. Verify: size > 50 % of the 7-day median; decrypt → `pg_restore --list` contains `items`, `users`, `ratings`; status `verified`.
5. Day 1: copy to `monthly/YYYY-MM/`. Prune daily > 30, monthly > 12 in app code (no bucket lifecycle rules [buckets]).
6. Off-site copy (D-32): on day 1 the app mails the owner a signed link (`storage.TemporaryURL`, 7 days, <https://vukem.dev/docs/security/signed-urls>) to the encrypted monthly dump; the job records `backup_runs(kind='offsite')` when the link is mailed and the owner downloads it to personal storage. Daily dumps stay on Railway. Replaced by an external bucket when the budget allows. As built (P2.E.2.3): `backup:offsite` (`app/backup/offsite.go`), scheduled `MonthlyOn(1, "07:00")` an hour after `backup.daily`; recipients `ADMIN_EMAILS`; the link is served by the app at `/backups/offsite/` with `storage.Handler` (monthly keys only, 404 otherwise) and signed with an HKDF sub-key of `APP_KEY`; the mail is sent synchronously so the link never sits in the jobs table; status `uploaded` means mailed, and one row per monthly dump makes reruns no-ops.
7. Missed or failed run → SDD 12 alert within 2 h.

Implementation (P2.E.2.2): `backup:run` uploads, then verifies (`app/backup/verify.go`: the median is over `verified` rows of the same kind from the previous 7 days, none means no size check; `pg_restore --list` runs with an empty environment and must list `TABLE public items|users|ratings`), copies the dump on UTC day 1 to `ObjectKey("monthly", ...)` by decrypting and sealing it again for the new key (frames are bound to their object key, so a byte copy would not decrypt), then prunes by count: the newest 30 daily and 12 monthly objects under each prefix stay (`app/backup/prune.go`). The `backup.*` gauges are registered with the other platform metrics (P2.C.2.2) once the OTel meter exists.

### 7.4 Restore drill (staging, first Sunday 07:00)

`backup:drill`: latest production daily via `BACKUP_SOURCE_S3_*` (and `BACKUP_SOURCE_BACKUP_KEY`; each falls back to the unprefixed name) → decrypt → `createdb hexascore_drill` → `pg_restore` on stdin (single job: `-j` needs a seekable file and the plaintext never touches disk; amended 2026-10-29, P2.E.3.1) → row counts vs dump manifest → `catalog:check` asserts (SDD 01 §7.6 plus scoring; SDD 06 §7.7 asserts join in P6) → drop → `backup_runs(kind='drill')`. Pass: < 60 min, 0 assert errors, counts equal. The manifest is written by `backup:run`: row counts of every `public` table taken in a `REPEATABLE READ` transaction whose snapshot `pg_dump --snapshot` reuses, sealed with `BACKUP_KEY` at `backups/manifests/postgres/<kind>/.../<dump name>.json`. Scheduled only where `BACKUP_SOURCE_S3_BUCKET` is set (staging also runs `APP_ENV=production`), weekly Sunday 07:00 with `--first-sunday`. Runbook: `docs/runbooks/restore.md`.

### 7.5 RTO runbook

Consistency note (SDD 13 open issue 11, 2026-10-28): after a restore, re-apply account deletions and anonymisations recorded after the backup timestamp (source: the `user.deleted` audit events kept in the append-only audit log and the Stripe customer-deletion webhooks), so a restored user never reappears.

1. Status page notice (GitHub Pages, SDD 12 §7.7); `down` [vk-cli] via `railway ssh` (Git Bash: `MSYS_NO_PATHCONV=1`, `vukem/docs/tooling.md` [SITE PAGE MISSING]).
2. New Postgres service, same region (~5 min).
3. `/app/app backup:restore backups/postgres/daily/<latest> --target-url "$NEW_DATABASE_URL"` (~20 min at 3 GB); it then replays the deletions recorded in `DATABASE_URL` (or `--replay-from`) after the dump; `--no-replay` only when that database is lost (runbook `docs/runbooks/restore.md`).
4. Point `DATABASE_URL` at the new service, redeploy (pre-deploy migrates).
5. `hexascore:cutover-verify --counts-only`, `dq:asserts`; `up`.
6. Users told the lost window (≤ 24 h, no replay, ADR 0014). Delete the old service after 7 days.

### 7.6 Legacy cutover (A-04, D-21, ADR 0013)

**Seed users (owner decision 2026-11-01).** The 10,000 legacy users (`@example.test` emails, `legacy_id` `seed-user-*`) are fake. They and their ratings, statuses, lists and time to beat stay in production for now, flagged `users.is_seed` (migration `20261101140000`). A seed user can never sign in (password, remember-me, app tokens, Google, passkeys, password reset all fail closed like a wrong password; `TestSecurity_A07_SeedUserCannotSignIn`). `./app seed:mark --email-suffix @example.test --legacy-prefix seed-user- [--dry-run]` sets the flag (a user must match every filter; admins are never marked). `./app seed:purge [--dry-run] [--batch 500] --confirm` later deletes them in batches (one transaction each; the 31 foreign keys to `users.id` cascade or set null, the tables without a FK are deleted explicitly), then queues the score recompute and the search suggest refresh of the affected items and prints counts per table.

- T-14 d: timed staging rehearsal of steps 2–7 on a fresh Neon copy; Google callback added (SDD 03 §7.7).
- T-7 d: `scripts/legacy/blob-export.ts` in this repository (Bun; `legacy/` is read-only, runbook `docs/migration/covers.md`) reads cover URLs **referenced by the DB** (orphans are never copied), uploads to the production bucket under SDD 06 keys, writes `manifest.ndjson {legacyUrl,key,sha256}`; ~79k objects (`legacy/docs/backup-local.md`).
- T-24 h: DNS TTL 60 s; banner on legacy.

Window (target 45 min, max 60):

1. Freeze: on Neon `ALTER DATABASE <db> SET default_transaction_read_only = on;` and terminate app sessions (legacy write-error UX [NOT VERIFIED]); disable the legacy Stripe webhook (Stripe retries).
2. `pg_dump -Fc --no-owner --no-acl` of `user, account, two_factor, verification, items, item_translations, source_records, ratings, item_status, time_to_beat, lists, list_items, api_keys, api_usage_daily, subscriptions, stripe_customers, credit_ledger, pix_charges, contact_messages, reports, affiliate_links, passkey`; excluded `session`, `external_scores` (D-09), `rate_limit`, `follows` (D-06, kept in the final dump), `processed_events` (webhook dedup is `contrib/stripe`'s cache, SDD 05 §5) — ~8 min.
3. `pg_restore` into schema `legacy_import`.
4. Purge synthetic users in the copy only: `user.id LIKE 'seed-user-%'` or `@example.test` emails and dependent rows (marker in `legacy/apps/web/lib/db/seed.ts`); counts logged.
5. `hexascore:import-legacy` (P2.F.2.1: SDD 03 §5 pre-flight first, abort with counts; stages `users`, `twofactor`, `catalog`, `entities`, `social`, `library`, `billing`, `finalise` in that order; checkpoint per stage in `legacy_import_runs`, written in each batch's transaction; `--dry-run` prints planned counts) `--stage users` (SDD 03), `catalog` (SDD 06 §7.9), `social` (ratings, statuses, TTB, lists), `billing` (api_keys unchanged, subscriptions, stripe_customers; no `processed_events`, SDD 05 §5; `credit_ledger`, `pix_charges` kept for release two, issue 9); covers delta; `search:rebuild`; score recompute (SDD 02).
6. `hexascore:cutover-verify`: per-table counts equal; 1 % id-sample `md5(string_agg(<cols>, '|' ORDER BY id))` equal; recomputed scores ±1 on 0–100; every cover key exists; 0 orphan FKs. Failure → R1–R2.

   Implementation (P2.F.4.1): the `finalise` stage rebuilds the score priors, recomputes every item with the nightly SQL path, queues every item and entity in `search_dirty` and drains the suggest sync (tsvectors are trigger-maintained, so this is the `search:rebuild` of this window). `hexascore:cutover-verify [--counts-only] [--sample-pct N]` only reads: counts mirror the stage filters (reports without an imported target, `free` subscriptions; `session`, `verification`, `rate_limit`, `external_scores`, `follows`, `processed_events` and credential accounts are not compared; `passkey` is imported by the users stage (incomplete rows skipped) and is not a count pair); the sample hashes verbatim columns of `users`, `items`, `lists`, `api_keys`, `credit_ledger`, `pix_charges`, `contact_messages` for ids with `abs(hashtext(id)) % 100 < N`; covers are `items.cover_key`/`cover_thumb_key` checked on the `S3_*` bucket; orphans are counted for every public foreign key plus `reports.target_id`; scores compare the stored SQL recompute with `scoring.ComputeItem` (Go) to 1e-4. Legacy `items.score` against the imported ratings' mean beyond ±1 is listed, never failed. Any failed check exits 1 naming the tables.
7. DNS switch (apex, `www`, `api`); Stripe live webhook to the new route and secret.
8. Watch 2 h, then daily for 7 days: 5xx < 0.5 %, SSR fallback 0, login success ≥ 95 % of legacy, Stripe events processed, job failures.

Rollback (until T+7 d, then forward-fix): R1 DNS back to Vercel (`docs/runbooks/dns-legacy.txt`); R2 Neon read-only off; R3 `hexascore:export-since <cutover>` replayed into legacy — ratings, statuses, TTB, lists (upsert by natural key), new API keys; new accounts replayed with a reset link (argon2id unreadable by Better Auth). Accepted lost: sessions, suggestions/moderation (no legacy table), flag changes. Stripe re-syncs once the legacy webhook is back. The command prints one upsert per line (`database/legacyexport`); `scripts/legacy/replay.sh` applies it. A user with no `legacy_id` is inserted as `hs:<id>` with a `verification` row `reset-password:<email>` / `required`. Ratings are copied as stored. Runbook: `docs/runbooks/cutover-rollback.md`.

T+30 d: final Neon dump to `backups/legacy/` (12 months); delete Vercel project (incl. `dev.hexascore.app`), Neon, Upstash, Blob store, legacy webhooks, old Google callback, `BETTER_AUTH_SECRET`; drop `legacy_import`.

### 7.7 Local development

Compose: `postgres:18.6-alpine3.24` [vk-deploy] on `127.0.0.1:5432`, `rustfs/rustfs:1.0.4` (vukem's S3 test server [vk-storage]). `vukem serve` + Vite HMR. `vukem db:seed DemoCatalog` [vk-seed]: `Seed(42)`, 1 admin, 50 users, 300 items, 2,000 ratings (replaces legacy 10k users / 2.1M ratings). Windows 11 first-class: `setups/setup-windows.ps1`, `setup-unix.sh` (port of `legacy/setups/`).

### 7.8 Hooks for SDD 12

JSON logs on stdout, OTel traces (Railway-injected `OTEL_*` [vk-deploy]), Railway CPU/RAM/volume metrics, `backup_runs`, `/up?checks=1`, SSR `OnFailure` counter [vk-inertia], `X-Hexascore-Version`.

## 8. Edge cases

- Pre-deploy times out on a long index build → deploy fails, old version serves; big indexes ship alone.
- SSR crash loop → restarted every 1 s; SDD 12 alerts on CSR fallback.
- `BACKUP_KEY` lost → dumps unreadable; the drill checks the offline copy.
- Tag pushed without a run → `gh workflow run release.yml -f version=vX.Y.Z`.

## 9. Acceptance criteria

- [ ] `docker image inspect hx --format '{{.Config.User}}'` → `app`; compressed size ≤ 400 MB.
- [ ] `grep -nE ':latest|^FROM [^@]+$' Dockerfile` → no output.
- [ ] `curl -s https://staging.hexascore.app/up?checks=1` → 200, `db`/`ssr`/`storage` `ok`.
- [ ] k6 20 rps across a staging redeploy → 0 non-2xx.
- [ ] Failing migration on staging → deployment failed, previous serving, `migrate:status --json` unchanged.
- [ ] `backup:run --kind daily` → `status='verified'`; `backup:drill` → exit 0, `duration_s < 3600`.
- [ ] Cutover rehearsal: `hexascore:cutover-verify` exit 0; window ≤ 60 min.
- [ ] `curl -sI https://hexascore.app/up | grep -i x-hexascore-version` → last tag.
- [ ] `SELECT extname FROM pg_extension` includes `pg_trgm`, `unaccent`; `SHOW server_version` ≥ 16, recorded here.
- [ ] Fresh Windows 11 clone: `setups\setup-windows.ps1; vukem serve` → home renders seed data.

## 10. Test strategy

Unit: frame encryption, prune selection. Integration (Postgres 18): concurrent `hexascore:predeploy`, invalid-index cleanup, legacy import fixture, cutover-verify on a seeded mismatch, export-since replay. Schemathesis, Playwright + axe on the built image; k6 deploy switch; weekly ZAP. Manual: monthly drill, quarterly timed RTO game day.

## 11. Open issues

1. Railway buckets lack object versioning and server-side encryption [buckets]. Decided (11-1): ADR 0014 amended to read "immutable content-hashed cover keys plus a 30-day deferred delete (`covers.gc`) replace bucket versioning; database dumps are encrypted client-side (vukem `crypt`, [vk-crypt]) before upload". Owner: architect; closes with an ADR 0014 amendment.
2. Railway buckets are private [buckets]. Decided (11-2): the app serves covers with `Cache-Control: public, max-age=31536000, immutable`, paying app egress; CDN decision when cover egress > US$5/month (trigger in the SDD 12 cost review, §6.5).
3. Migration lock [NOT IN VUKEM DOCS]. VF-026; app-level advisory lock around the runner (§7.2). Owner: vukem.
4. Non-transactional migrations (VF-017). Closed: vukem v0.33.0 `schema.Migration.NoTransaction`; `INVALID` leftovers handled by `hexascore:predeploy` (P2.E.1.1, VF-054).
5. SSR supervision inside `serve`, version in `/up`, extra ldflags in `vukem build` [NOT IN VUKEM DOCS]. VF-027; `build:ssr` is a permanent pipeline step (§6.3). Owner: vukem.
6. Railway Postgres major, extensions, `ALTER SYSTEM`, private hostname, Alpine `postgresql18-client` [NOT VERIFIED] (SDD 07 issue 9). Owner: architect; closes at provisioning.
7. Queue priority order of `QUEUE_NAMES` for the database driver (SDD 06 issue 5; concurrency now documented [vk-queue]). Owner: vukem.
8. Railway IaC keys for overlap/draining/restart, config-as-code deprecation, `railway up` flags, wait-for-CI [NOT VERIFIED]. Owner: architect; closes on the first staging deploy.
9. `credit_ledger` balances. Decided (D-23): honoured as extra quota (SDD 05 §7.2); the import stage writes `api_credit_balances`.
10. Apex domain DNS. Decided (D-28): Cloudflare with CNAME flattening, proxy off at launch (§6.1). Owner task: move or confirm the zone before T-24 h.
11. Dumps and database on one provider. Decided (D-32): monthly dump downloaded by the owner via signed link (§7.3 step 6); closes at the first budget increase (external bucket).
12. GitHub Actions. Decided (D-26): available; pre-push mirror kept as a convenience.
13. §6.5 usage averages unmeasured. Owner: architect; closes after 30 days of Railway metrics.
