# Ingestion worker (second Railway service)

RFC-001 section 3, task CW-01, owner decision OQ-03 / D-40: the ingest queue runs in its own Railway service built from the same image. **Status checked 2026-11-03 via Railway MCP: staging worker deployed; no worker deployment or configuration in production.** Creating it, setting its variables and turning ingestion on are owner actions (agents cannot write production variables or deploy). Do staging first, then production.

The staging worker's latest deployment was `ff53eda9-2e88-461f-9914-354a37b5ca0d` (SUCCESS, 14:01 UTC). Its logs show `queue worker started` with queue `ingest`, concurrency 3 and limit 2; they do not establish completed ingestion batches or seven healthy days. The web follows branch `staging`, while the worker follows `main`: align their tested revisions before collecting the M0 evidence. Production worker promotion remains pending that evidence and owner approval.

## What changes

| | `app` (web) | `worker` (new) |
| --- | --- | --- |
| Command | image entrypoint (`/app/app serve`) | `/app/app queue:work -queue ingest` (a Railway start command replaces the entrypoint, so it names the binary) |
| Queues | `default` | `ingest` |
| Scheduler | yes, dispatches the chain starts to `ingest` | none |
| HTTP | yes | no (Railway cannot health-check it; the web reports it, see "Check") |
| Source | `AlexGalhardo/hexascore-app`, environment's release branch | the same repo and tested revision as the web; verify both source branches before promotion |

The scheduler stays in the web process: a chain start is already a queue job, so the worker only needs the queue. The worker never serves traffic, so it carries no domain.

## Create the service (staging first)

In the Railway dashboard, project `hexascore`, environment `staging`:

1. New service, from the GitHub repo `AlexGalhardo/hexascore-app`, branch `main`. Name it `worker`.
2. Settings, Deploy: custom start command `/app/app queue:work -queue ingest`. No healthcheck path, no public domain, no pre-deploy command (the web runs the migrations).
3. Variables: first reference the web's database and keys, then the worker-only ones (table below).
4. Deploy, then follow "Check".

The CLI sets variables the same way for both environments; a secret goes through stdin, never through chat or the shell history:

```text
railway variable set QUEUE_NAMES=ingest --service worker --environment staging
printf "%s" "$VALUE" | railway variable set THEMOVIEDATABASE_API_READ_ACCESS_TOKEN --stdin --service worker --environment staging
```

## Variables

Same values as `app` (Railway reference variables or a copy): `APP_ENV=production`, `APP_KEY`, `DATABASE_URL`, `S3_*` (covers), `COVERS_BASE_URL`, `INGEST_CONTACT_URL`.

Worker only:

| Variable | Value | Why |
| --- | --- | --- |
| `QUEUE_CONNECTION` | `database` | required: `sync` runs jobs inside dispatch and `queue:work` exits without consuming anything |
| `QUEUE_NAMES` | `ingest` | it consumes only the ingest queue |
| `QUEUE_CONCURRENCY` | `3` | RFC-001 section 3 |
| `QUEUE_LIMITS` | `ingest=2` | the running-job cap that already applies to ingestion |
| `SCHEDULE_IN_SERVE` | `false` | no scheduler here |
| `GOMEMLIMIT` | `400MiB` | the batch target is under 400 MB |
| `DB_MAX_OPEN` | `4` | at least 2 are needed (a step transaction plus the cover reservation) |
| `FLAG_INGEST_ENABLED` | `true` | master switch; stays `false` until the staging run below is green |
| `FLAG_INGEST_<SOURCE>` | `true` one by one | `WIKIDATA`, `WIKIPEDIA`, `OPENLIBRARY`, `TMDB`, `IGDB`, `STEAM`, `GOG`, `COVERS`, `YOUTUBE` |
| `THEMOVIEDATABASE_API_READ_ACCESS_TOKEN` | secret | TMDB (also the watch providers of RFC-003) |
| `IGDB_TWITCH_CLIENT_ID`, `IGDB_TWITCH_CLIENT_SECRET` | secrets | IGDB |
| `FLAG_OFFERS_VERIFY` | `true` later | the daily offer-link check (off until the owner wants it) |
| `FLAG_INGEST_LLM` | `true` later | the model normaliser of RFC-001 section 5 F (off until the owner wants it; check the provider's current prices first) |
| `ANTHROPIC_API_KEY` | secret | only read when `FLAG_INGEST_LLM` is on |
| `LLM_MONTHLY_BUDGET_USD` | `3` | hard ceiling per UTC month; at the ceiling the job stops and `alert=A18` pages once |

On `app` (web):

| Variable | Value | Why |
| --- | --- | --- |
| `QUEUE_CONNECTION` | `database` | required on the producer too, so dispatch persists jobs for the separate worker |
| `QUEUE_NAMES` | `default` | the web stops consuming `ingest` (or set `QUEUE_LIMITS=ingest=0`) |
| `FLAG_INGEST_ENABLED` | `true` | **required**: the web task `ingest:pressure_signal` writes the load signal only while the master switch is on in the web too, and the worker reads that signal before every batch |
| `FLAG_INGEST_<SOURCE>` | same enabled sources as the worker | **required**: `routes/schedule.go` checks each chain's source flag before enqueueing its first stage; flags only on the worker leave scheduled source chains idle. Persisted `/admin/_flags` overrides also apply |
| `WORKER_HEARTBEAT_CHECK` | `true` | set it only after the worker has beaten once; it adds the `worker` entry to `/up?checks=1` |

Never put a production secret in a variable of the staging environment or the reverse.

## Check

1. `railway logs --service worker --environment staging` shows jobs being processed (`offers verify`, `ingest.worker_heartbeat`, run logs) a minute after the deploy.
2. `curl -s "https://staging.hexascore.app/up?checks=1"` lists `worker: ok` once `WORKER_HEARTBEAT_CHECK=true` is set on the web. A worker that stopped consuming the queue shows `worker: fail` within three minutes (the heartbeat is a queued job, so a running process that is stuck is caught too).
3. Run a small chain by hand from `/admin/ingest` with `limit` 100, watch it finish, then let the schedule take over. Seven green days on staging come before production (P6.A.4.2).
4. The 30-minute Telegram digest (`ingest:digest`, silent ops chat) shows the worker's items, failures and heartbeat age.

Record the seven-day window, failed queue count, refresh run totals (at least 1,000 items), cover coverage over 14 days, accuracy sample (at least 97%) and spend (at most US$12) before M0 promotion (RFC-001 section 10). A successful deployment or a fresh heartbeat alone does not prove source ingestion. New discovery towards 250k, 500k and 1M follows the later milestone gates; M0 is refresh only.

## Roll back

Turn the flags off on the worker (`FLAG_INGEST_ENABLED=false`) and set the web's `QUEUE_NAMES` back to `default,ingest`: the web consumes the ingest queue again, exactly as before the worker existed. Deleting the `worker` service loses nothing: runs resume from their cursors (`job_runs.cursor`) and the queue lives in Postgres.

Plan B (RFC-001 section 3), if the owner keeps one service: `QUEUE_LIMITS=ingest=1` on `app` and the heavy chains only between 00:00 and 06:00 America/Sao_Paulo, which caps growth near 1,500 items a day.

## Cost

About US$2.5 a month at 0.1 vCPU average and 0.25 GB (RFC-001 section 6). The `projected-spend.yml` alert at US$18 and the US$20 hard limit (D-27) stay the guard. Raising the cap is decision OQ-03.
