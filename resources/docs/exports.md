# Exports

Download the whole list of item ids once a day, then fetch only what you need.

## What is exported

Each day, before 08:00 UTC, we write one file per item type: games, movies, series and books. Files are kept for 90 days. Exports are available on the Pro and Business plans; the Free plan receives [`plan-required`](/docs/errors#plan-required).

A file is gzip-compressed NDJSON, one JSON object per line:

```json
{"id":"0192f3a4-7b1c-7e21-9c55-3f1e2d4a6b71","slug":"hades","wikidata":"Q62640075","updatedAt":"2026-11-02T06:12:44Z"}
```

| Field | Meaning |
| --- | --- |
| `id` | The HexaScore item id, as used by `GET /v1/{type}/{id}`. |
| `slug` | The URL slug of the item page. |
| `wikidata` | The Wikidata entity id. |
| `updatedAt` | When the item last changed. |

The files carry ids only. Item data, scores and licence-restricted fields come from the regular endpoints, so the plan's rules always apply.

## Download

`GET /v1/exports` lists the available files. `GET /v1/exports/{type}/{day}` (for example `/v1/exports/games/2026-11-02`) answers `302` to a download URL that is valid for 5 minutes. Follow the redirect straight away and do not store the URL.

```sh
curl -L -H "Authorization: Bearer $HEXASCORE_API_KEY" \
  "https://api.hexascore.app/v1/exports/games/2026-11-02" -o games_ids.json.gz
gunzip -c games_ids.json.gz | head -n 3
```

An unknown day, or one older than 90 days, returns [`not-found`](/docs/errors#not-found).

## A good routine

- Download the files once a day and compare with your copy: new ids are creations, missing ids are deletions.
- Fetch the new ids with `GET /v1/{type}/{id}`, respecting your [rate limits](/docs/rate-limits).
- Between exports, follow [Changes](/docs/changes) for the minute-by-minute view.
