# Pagination

List endpoints use opaque cursors: stable under inserts, fast at any depth.

## How it works

`GET /v1/games`, `/v1/movies`, `/v1/series` and `/v1/books` return `{ "data": [...], "meta": { "next_cursor": "..." } }`. To get the next page, repeat the same request with `cursor=<meta.next_cursor>`. When `meta.next_cursor` is absent you reached the end.

- Keep every other parameter identical. A cursor from another `sort` returns `400`.
- Cursors are opaque: do not parse or build them, their format may change.
- Ordering is stable: ties are broken by item id, and unscored or undated items always come last.

## Page size

`limit` defaults to 10. The maximum depends on your plan: 10 on Free, 100 on Pro and Business. A larger value returns `400 invalid-query`; the API never silently truncates. Each page counts as one request.

## Example: iterate over every page

```js
let cursor = null;
do {
  const url = new URL("https://api.hexascore.app/v1/games");
  url.searchParams.set("limit", "100");
  if (cursor) url.searchParams.set("cursor", cursor);
  const res = await fetch(url, { headers: { Authorization: `Bearer ${process.env.HEXASCORE_API_KEY}` } });
  const page = await res.json();
  for (const game of page.data) console.log(game.title);
  cursor = page.meta.next_cursor ?? null;
} while (cursor);
```

## Search

`GET /v1/search` is not paginated: it returns the best matches (up to 50) for a query, best first.
