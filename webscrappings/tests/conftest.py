import json
from collections.abc import Callable
from pathlib import Path
from typing import Any

import httpx
import pytest

from common.http import HttpClient
from connectors import wikidata

FIXTURES = Path(__file__).parent / "fixtures"


def load(name: str) -> Any:
    return json.loads((FIXTURES / name).read_text(encoding="utf-8"))


def mock_client(handler: Callable[[httpx.Request], httpx.Response]) -> HttpClient:
    return HttpClient(transport=httpx.MockTransport(handler), sleep=lambda _: None)


def wikidata_handler(request: httpx.Request) -> httpx.Response:
    """Serves recorded wbgetentities/SPARQL responses, filtered to the requested ids."""
    if request.url.host == "query.wikidata.org":
        return httpx.Response(200, json=load("sparql_top_movies.json"))
    ids = request.url.params["ids"].split("|")
    name = "wikidata_labels.json" if request.url.params["props"] == "labels" else "wikidata_entities.json"
    entities = load(name)["entities"]
    return httpx.Response(200, json={"entities": {i: entities.get(i, {"id": i, "missing": ""}) for i in ids}})


@pytest.fixture
def wikidata_records() -> dict[str, dict]:
    """Trimmed payloads (as stored in source_records) for the recorded entities, keyed by QID."""
    types = {"Q83495": "movie", "Q4267401": "game", "Q1079": "series", "Q208460": "book"}
    client = mock_client(wikidata_handler)
    return {r.external_id: r.payload for r in wikidata.fetch_entities(client, types.items())}
