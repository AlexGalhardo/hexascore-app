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
import { createHexascoreClient, getGame, listGames, lookupItem } from "@hexascore/sdk";

const client = createHexascoreClient({ apiKey: process.env.HEXASCORE_API_KEY! });

const { data, error, response } = await listGames({
	client,
	query: { sort: "-score", limit: 10, min_gameplay: 80 },
});
if (error) throw new Error(`${error.code}: ${error.detail} (request ${error.request_id})`);
for (const game of data.data) console.log(game.score, game.title);

const first = await getGame({ client, path: { id: data.data[0].id } });
const shawshank = await lookupItem({ client, query: { source: "imdb", id: "tt0111161" } });
console.log(response.headers.get("ratelimit-remaining"));
```

### Operations

| Resource | Functions |
| --- | --- |
| Games | `listGames`, `getGame`, `listGameGenres` |
| Movies | `listMovies`, `getMovie`, `listMovieGenres` |
| Series | `listSeries`, `getSeries`, `listSeriesGenres` |
| Books | `listBooks`, `getBook`, `listBookGenres` |
| Search | `searchItems`, `lookupItem` |
| Account | `getUsage` |

Types such as `ItemSummary`, `ItemDetail`, `ItemList`, `Problem` and `Usage` are exported too.

### Pagination

List endpoints are cursor based: pass `next_cursor` of one page as `cursor` of the next until it is `null`.

```ts
let cursor: string | undefined;
do {
	const { data, error } = await listMovies({ client, query: { sort: "-score", limit: 100, cursor } });
	if (error) throw new Error(error.detail);
	for (const movie of data.data) console.log(movie.title);
	cursor = data.next_cursor ?? undefined;
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
	baseUrl: "https://<host>/api/v1", // another deployment; defaults to https://api.hexascore.app/v1
	fetch: myFetch, // custom fetch (tests, proxies)
});
```

## Development

```sh
bun run sdk:generate      # from the repo root: regenerate src/gen from apps/web/public/openapi.json
cd sdks/ts
bun run typecheck && bun test && bun run build
```

`src/gen/` is generated: never edit it. The only hand-written code is `src/index.ts`. Regenerate whenever
`openapi.json` changes (`bun run openapi`). Publishing: `docs/sdk-publish-npm.md` and `docs/sdk-publish-jsr.md` in
the repository.

## License

MIT
