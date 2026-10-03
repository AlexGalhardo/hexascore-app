# hexascore (Python SDK)

Typed Python client for the [Hexascore API](https://hexascore.com/docs), generated from
[`openapi.json`](../../apps/web/public/openapi.json) with
[`openapi-python-client`](https://github.com/openapi-generators/openapi-python-client) **0.29.1** (pinned, run via
`uvx`). Python ≥ 3.11, built on `httpx` + `attrs`. The package lives in [`hexascore/`](hexascore/) and is fully
generated — never edit it by hand.

> Not on PyPI yet. Until the first release: `uv add ./sdks/python/hexascore` (or `pip install ./sdks/python/hexascore`).

## Usage

```python
import os

from hexascore import AuthenticatedClient
from hexascore.api.games import get_game, list_games
from hexascore.api.search import lookup_item
from hexascore.models import ItemList, ListGamesSort, LookupItemSource, Problem

client = AuthenticatedClient(base_url="https://api.hexascore.com/v1", token=os.environ["HEXASCORE_API_KEY"])

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

Every endpoint has `sync`, `sync_detailed`, `asyncio` and `asyncio_detailed` variants; errors (4xx/5xx) parse into
`Problem`. Use `base_url="https://<host>/api/v1"` to target another deployment.

## Regenerate

```bash
bun run sdk:generate:python   # from the repo root; needs uv
```

Equivalent to:

```bash
uvx --from openapi-python-client==0.29.1 openapi-python-client generate \
  --path apps/web/public/openapi.json --config sdks/python/config.yml \
  --output-path sdks/python/hexascore --meta uv --overwrite
```

Run it whenever `openapi.json` changes (`bun run openapi`). Post-hooks format the output with `ruff` 0.16.10.
