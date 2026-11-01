# @hexascore/sdk

Typed TypeScript client for the [Hexascore API](https://hexascore.app/docs): community scores (0–100, six
categories per item) for games, movies, TV series and books.

- Every endpoint of the API as a typed function, with request and response types.
- No runtime dependencies: it uses the platform `fetch`. Works in Node 18+, Bun, Deno, browsers and edge runtimes.
- ESM only. Generated from the API's [OpenAPI document](https://hexascore.app/openapi.json) with
  [`@hey-api/openapi-ts`](https://heyapi.dev).

## Install

```sh
npm install @hexascore/sdk      # or: bun add / pnpm add / yarn add
```

From [JSR](https://jsr.io/@hexascore/sdk) (Deno, or any runtime through the JSR CLI):

```sh
deno add jsr:@hexascore/sdk
npx jsr add @hexascore/sdk      # or: bunx jsr add / pnpm dlx jsr add
```

## Usage

Create an API key at <https://hexascore.app/profile/api> and keep it on the server.

```ts
import { createHexascoreClient, getItem, listItems, lookupItem } from "@hexascore/sdk";

const client = createHexascoreClient({ apiKey: process.env.HEXASCORE_API_KEY! });

const { data, error, response } = await listItems({
	client,
	path: { type: "games" },
	query: { sort: "-score", limit: "10", min_score: "8" },
});
if (error) throw new Error(`${error.code}: ${error.detail} (request ${error.request_id})`);
for (const game of data.data) console.log(game.score, game.title);

const first = await getItem({ client, path: { type: "games", id: data.data[0].id } });
const shawshank = await lookupItem({ client, query: { source: "imdb", id: "tt0111161" } });
console.log(response.headers.get("x-quota-remaining"));
```

### Operations

| Resource | Functions |
| --- | --- |
| Items (`type` = `games`, `movies`, `series` or `books`) | `listItems`, `getItem`, `listEditions` (books) |
| Genres | `listGameGenres`, `listMovieGenres`, `listSeriesGenres`, `listBookGenres` |
| Entities | `getEntity`, `listEntityItems`, `getFranchise`, `getOrganization`, `getPerson`, `getPlatform` |
| Search | `searchItems`, `searchEntities`, `lookupItem` |
| Changes | `listChanges` (`since` cursor over catalogue edits) |
| Webhooks | `listWebhooks`, `createWebhook`, `deleteWebhook`, `testWebhook` |
| Account | `getUsage` |

Types such as `ItemSummary`, `ItemDetail`, `ItemList`, `Problem` and `Usage` are exported too.

### Pagination

List endpoints are cursor based: pass `meta.next_cursor` of one page as `cursor` of the next until it is absent.

```ts
let cursor: string | undefined;
do {
	const { data, error } = await listMovies({ client, query: { sort: "-score", limit: 100, cursor } });
	if (error) throw new Error(error.detail);
	for (const movie of data.data) console.log(movie.title);
	cursor = data.meta?.next_cursor;
} while (cursor);
```

### Errors

Errors are returned, not thrown: `error` is the API's RFC 9457 problem (`status`, `code`, `detail`, `request_id`).
Pass `throwOnError: true` to an operation to throw instead. Network failures always throw.

| Status | `code` | Meaning |
| --- | --- | --- |
| 400 | `bad_request` | Invalid parameter (`errors` lists each one) |
| 401 | `unauthorized` | Missing or invalid API key |
| 402 | `payment_required` | Plan quota used up and no credits left |
| 403 | `forbidden` | The account is suspended |
| 404 | `not_found` | No such item |
| 429 | `rate_limited` | Too many requests; see the `retry-after` header |
| 500 / 503 | `internal_error` / `service_unavailable` | Server-side failure; retry later |

### Options

```ts
createHexascoreClient({
	apiKey: "hx_live_…",
	baseUrl: "https://<host>/v1", // another deployment; defaults to https://api.hexascore.app/v1
	fetch: myFetch, // custom fetch (tests, proxies)
});
```

## Development

```sh
cd sdks/ts
bun run sdk:generate      # regenerate src/gen from openapi/api-v1.json
bun run typecheck && bun test && bun run build
```

`src/gen/` is generated: never edit it. The only hand-written code is `src/index.ts`. Regenerate whenever
`openapi/api-v1.json` changes (`vukem openapi:export`). Publishing: `docs/runbooks/sdk-release.md` in
the repository.

## License

MIT
