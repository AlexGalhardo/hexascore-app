# webscrappings — catalog ingestion

Python 3.14 + uv. Loads games, movies, series and books from open APIs into the app's Postgres
(`source_records` → `items`, `item_translations`, `external_scores`). No HTML scraping: only Wikidata
(SPARQL + `wbgetentities`), Wikipedia REST `page/summary`, Open Library API, Steam `appdetails` and the
GOG public catalog API. Details and decisions: [`docs/ingestion.md`](../docs/ingestion.md).

## Setup

```sh
cd webscrappings
uv sync
export DATABASE_URL=postgresql://hexascore:hexascore@localhost:5432/hexascore   # a migrated database
```

| Env var | Required | Use |
| --- | --- | --- |
| `DATABASE_URL` | yes | Postgres with the app migrations applied (`apps/web/lib/db/migrations`) |
| `INGEST_CONTACT_URL` | no | contact URL in the `HexascoreBot/0.1 (+url)` User-Agent (default: the GitHub repo) |
| `APP_URL`, `CRON_SECRET` | no | `POST $APP_URL/api/cron/revalidate` after changes; skipped when unset |
| `BLOB_READ_WRITE_TOKEN` | no | copy covers to Vercel Blob; without it `meta.coverUrl` is the source URL |
| `TEST_DATABASE_URL` | no | enables the DB tests (`tests/test_db.py`), which roll back everything |

## Run

```sh
uv run python -m bootstrap --limit 20                  # small local load: top 20 per type, enrich, normalize, covers
uv run python -m bootstrap --export snapshot.sql.gz    # full targets (~100k items) + snapshot
uv run python -m bootstrap --export-only --export snapshot.sql.gz
gunzip -c snapshot.sql.gz | psql "$DATABASE_URL"       # restore into a freshly migrated database

uv run python -m connectors.wikidata --since 36h       # items edited since (also: --types, --limit, --refresh-known)
uv run python -m connectors.wikipedia --since 8d       # also: --langs en,pt,es, --missing-only
uv run python -m connectors.openlibrary --since 8d     # every enrichment connector takes --since / --missing-only
uv run python -m connectors.steam --since 36h
uv run python -m connectors.gog --since 8d
uv run python -m normalize                             # pending source_records -> catalog, then revalidate
uv run python -m covers                                # fill meta.coverUrl (--refresh, --limit)
uv run python -m revalidate
```

`--since` accepts `36h`, `8d` or an ISO timestamp.

## Quality

```sh
uv run ruff check && uv run ruff format --check && uv run pytest
```

Tests use recorded API responses in `tests/fixtures/` (no network).

## Schedule (GitHub Actions)

| Workflow | When | What |
| --- | --- | --- |
| `ingest-bootstrap.yml` | manual (`limit`, `types`) | top-N Wikidata → enrichment jobs in parallel (resumable) → normalize, covers, snapshot release |
| `ingest-incremental.yml` | daily 04:00 UTC | Wikidata edits (36h) → Wikipedia + Steam for what changed → normalize, covers |
| `ingest-weekly.yml` | Sunday 05:00 UTC | re-read every known entity (scores), synopses, Open Library, GOG → normalize, covers |
| `python-ci.yml` | PRs touching `webscrappings/**` | ruff + pytest (DB tests against Postgres 18.6) |

## Dependencies

Runtime: `httpx`, `psycopg[binary]`, `pydantic` (validates every external payload). Dev: `pytest`, `ruff`.
All pinned to exact versions in `pyproject.toml`; nothing else is needed.
