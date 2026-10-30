# k6 baseline

Measured 2026-10-30 on a developer laptop (Docker Desktop, Windows 11), app container built from `main` plus P1.I.3.1, compose profile `app` (`COMPOSE_PROJECT_NAME=hx-senior`), demo catalogue (300 items, `DemoCatalog` seeder). k6 `grafana/k6:2.3.0` with `--network host`, 60 s per scenario. **This is not the NFR dataset** (500k items, staging): it proves the scripts and the thresholds' plumbing, and the numbers are a floor, not a capacity statement. Re-run on staging (P2.F) and replace this table.

## Setup

```sh
export APP_KEY="base64:$(openssl rand -base64 32)"
docker compose --profile app up -d --build --wait app
PSQL="docker exec -i <project>-postgres-1 psql -U hexascore -d hexascore" qa/k6/seed.sh > /tmp/k6.env
. /tmp/k6.env   # API_KEYS (10 Business keys), LIST_PATHS (10 public lists of 25 games)
docker run --rm --network host -e BASE_URL=http://127.0.0.1:3000 -e API_KEYS -e API_KEY="${API_KEYS%%,*}" -e LIST_PATHS \
  -v "$PWD/qa:/qa" grafana/k6:2.3.0 run /qa/k6/api-v1.js   # or public-lists.js, pages-ttfb.js
```

`seed.sh` writes only to the database named by `PSQL`; never point it at production. The scripts send a distinct `X-Forwarded-For` per request (the compose stack trusts `172.16.0.0/12`) so the 120/min per-IP limit does not mask latency, and spread requests across the keys round-robin: a Business key allows 50 rps, so skewed key choice produces genuine 429s (seen once while writing the script).

## Results

| Script | Load | Threshold | Result |
| --- | --- | --- | --- |
| `api-v1.js` | 25 rps list + 25 rps detail, 10 keys | list p95 < 300 ms, detail p95 < 150 ms, failed < 0.1 % | list p95 16.4 ms, detail p95 14.7 ms, 0 of 3006 failed |
| `public-lists.js` | 30 rps | p95 < 400 ms, failed < 0.5 % | p95 9.9 ms, 0 of 1800 failed |
| `pages-ttfb.js` | 10 rps each on home, listing, item, entity | TTFB p75 < 400 ms | home 20.8, entity 21.0, item 12.8, listing 11.6 ms; 0 of 2405 failed |

## Not covered here

- `search.js` and `rating-hot-item.js` (existing, unchanged): `rating-hot-item.js` needs 50 `/app/v1` tokens and an item with ~10,000 ratings; `search.js` needs `SEARCH_CACHE=off`. Both belong to the staging run.
- `DURATION=10m` for the full NFR-05-1 run; the 500k-item dataset.
- Open issue 09-3 (TTFB without an SSR output cache): passes on the demo catalogue, so the VF-025 anonymous HTML cache workaround is not opened as a microtask; confirm on staging data.
