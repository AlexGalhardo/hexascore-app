# @hexascore/sdk

Typed TypeScript client for the [Hexascore API](https://hexascore.com/docs): community scores (0–100, six
categories per item) for games, movies, TV series and books. Generated from
[`openapi.json`](../../apps/web/public/openapi.json) with [`@hey-api/openapi-ts`](https://heyapi.dev); works in
Node 18+, Bun, Deno and edge runtimes (uses `fetch`).

> Not published to npm yet. Until the first release, build it from this folder (`bun run build`) and depend on it
> via a workspace/file path.

## Usage

```ts
import { createHexascoreClient, getGame, listGames, lookupItem } from "@hexascore/sdk";

const client = createHexascoreClient({ apiKey: process.env.HEXASCORE_API_KEY! });

const { data, error, response } = await listGames({ client, query: { sort: "-score", limit: 10, min_gameplay: 80 } });
if (error) throw new Error(`${error.code}: ${error.detail} (request ${error.request_id})`);
for (const game of data.data) console.log(game.score, game.title);

// Next page
const next = data.next_cursor ? await listGames({ client, query: { sort: "-score", cursor: data.next_cursor } }) : null;

const zelda = await getGame({ client, path: { id: data.data[0].id } });
const shawshank = await lookupItem({ client, query: { source: "imdb", id: "tt0111161" } });
console.log(response.headers.get("ratelimit-remaining"));
```

Every operation of the API has a function (`listGames`, `getMovie`, `listBookGenres`, `searchItems`, `lookupItem`,
`getUsage`, …) and request/response types (`ItemSummary`, `ItemDetail`, `Problem`, …). Errors are returned, not
thrown; pass `throwOnError: true` to throw instead. Use `baseUrl` to target another deployment
(`https://<host>/api/v1`).

## Development

```bash
bun run sdk:generate   # from the repo root: regenerate src/gen from apps/web/public/openapi.json
cd sdks/ts && bun run typecheck && bun run build
```

`src/gen/` is generated — never edit it. The only hand-written code is `createHexascoreClient` in `src/index.ts`.
Regenerate whenever `openapi.json` changes (`bun run openapi`).
