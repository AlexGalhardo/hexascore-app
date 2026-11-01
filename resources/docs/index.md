# Quickstart

Community scores for games, movies, TV series and books, from 0 to 10 in six categories per item, over a read-only JSON API.

## 1. Create an API key

Open your [API dashboard](/profile/api) and create a key. It starts with `hx_live_` and is shown only once, so store it in an environment variable and never in client-side code.

The Free plan includes 100 requests in total, enough to build and test an integration.

## 2. Make your first request

The base URL is `https://api.hexascore.app/v1`. Every endpoint is a `GET`. Here are the five best-rated games:

```sh
curl -H "Authorization: Bearer $HEXASCORE_API_KEY" \
  "https://api.hexascore.app/v1/games?sort=score&limit=5"
```

## 3. Read the response

Lists return `data` plus `meta.next_cursor` for the next page. Field names are snake_case and dates are ISO 8601. `score` and `categories` are `null` until an item has 5 ratings.

## Next steps

- Fetch one item with `GET /v1/games/{id}`.
- Map ids you already have with `GET /v1/lookup`.
- Handle [errors](/docs/errors) and [rate limits](/docs/rate-limits), then browse the [API reference](/docs/reference).
