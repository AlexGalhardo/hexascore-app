# hexascore

Typed Python client for the [Hexascore API](https://hexascore.app/docs): community scores (0–100, six categories
per item) for games, movies, TV series and books.

- Every endpoint of the API as a typed function, in sync and async variants.
- Python 3.11+, built on [`httpx`](https://www.python-httpx.org/) and [`attrs`](https://www.attrs.org/); ships
  type information (`py.typed`).
- Generated from the API's [OpenAPI document](https://hexascore.app/openapi.json) with
  [`openapi-python-client`](https://github.com/openapi-generators/openapi-python-client).

## Install

```sh
pip install hexascore      # or: uv add hexascore / poetry add hexascore
```

## Usage

Create an API key at <https://hexascore.app/profile/api> and keep it on the server.

```python
import os

from hexascore import AuthenticatedClient
from hexascore.api.games import get_game, list_games
from hexascore.api.search import lookup_item
from hexascore.models import ListGamesSort, LookupItemSource, Problem

client = AuthenticatedClient(base_url="https://api.hexascore.app/v1", token=os.environ["HEXASCORE_API_KEY"])

with client as c:
    page = list_games.sync(client=c, sort=ListGamesSort.SCORE_DESC, limit=10, min_gameplay=80)
    if isinstance(page, Problem):
        raise RuntimeError(f"{page.code}: {page.detail} (request {page.request_id})")
    for game in page.data:
        print(game.score, game.title)

    detail = get_game.sync(client=c, id=page.data[0].id)
    found = lookup_item.sync_detailed(client=c, source=LookupItemSource.IMDB, id="tt0111161")
    print(found.status_code, found.headers.get("ratelimit-remaining"))
```

### Operations

Each module has four functions: `sync`, `sync_detailed`, `asyncio` and `asyncio_detailed`. The plain ones return
the parsed body; the `_detailed` ones return a `Response` with `status_code`, `headers`, `content` and `parsed`.

| Package | Modules |
| --- | --- |
| `hexascore.api.games` | `list_games`, `get_game`, `list_game_genres` |
| `hexascore.api.movies` | `list_movies`, `get_movie`, `list_movie_genres` |
| `hexascore.api.series` | `list_series`, `get_series`, `list_series_genres` |
| `hexascore.api.books` | `list_books`, `get_book`, `list_book_genres` |
| `hexascore.api.search` | `search_items`, `lookup_item` |
| `hexascore.api.account` | `get_usage` |

Models (`ItemSummary`, `ItemDetail`, `ItemList`, `Problem`, `Usage`, the `…Sort` and `…Locale` enums) are in
`hexascore.models`.

### Async

```python
import asyncio

from hexascore import AuthenticatedClient
from hexascore.api.movies import list_movies


async def main() -> None:
    async with AuthenticatedClient(base_url="https://api.hexascore.app/v1", token="hx_live_…") as client:
        page = await list_movies.asyncio(client=client, limit=5)
        print(page)


asyncio.run(main())
```

### Pagination

List endpoints are cursor based: pass `next_cursor` of one page as `cursor` of the next until it is `None`.

```python
from hexascore.types import UNSET

cursor = UNSET
while True:
    page = list_games.sync(client=c, limit=100, cursor=cursor)
    if isinstance(page, Problem):
        raise RuntimeError(page.detail)
    for game in page.data:
        print(game.title)
    if page.next_cursor is None:
        break
    cursor = page.next_cursor
```

### Errors

4xx and 5xx responses parse into `Problem` (RFC 9457: `status`, `code`, `detail`, `request_id`); they are returned,
not raised. Network failures raise `httpx` exceptions. Pass `raise_on_unexpected_status=True` to the client to raise
`hexascore.errors.UnexpectedStatus` on a status the API does not document.

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

`AuthenticatedClient` accepts `timeout` (an `httpx.Timeout`), `headers`, `cookies`, `verify_ssl`,
`follow_redirects` and `httpx_args`. Use `base_url="https://<host>/api/v1"` to target another deployment.

## Development

The package in [`hexascore/`](hexascore/) is generated: never edit it. `pyproject.toml`, this README, the LICENSE
and `tests/` are written by hand and survive regeneration.

```sh
bun run sdk:generate:python   # from the repo root; needs uv. Rewrites sdks/python/hexascore
cd sdks/python
uv run pytest -q              # smoke tests against a mocked transport
uv build                      # dist/hexascore-<version>.tar.gz and .whl
```

Regenerate whenever `openapi.json` changes (`bun run openapi`). The generator is pinned
(`openapi-python-client` 0.29.1, run through `uvx`) and formats its output with `ruff` 0.16.10. Publishing:
`docs/sdk-publish-pypi.md` in the repository.

## License

MIT
