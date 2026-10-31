"""Smoke tests for the generated client against a mocked transport (no network)."""

import httpx
from hexascore.api.catalog import get_item, list_items

from hexascore import AuthenticatedClient
from hexascore.models import ItemPageJSON, Problem

BASE_URL = "https://api.hexascore.app/v1"


def client_with(handler) -> AuthenticatedClient:
    client = AuthenticatedClient(base_url=BASE_URL, token="hx_live_test")
    client.set_httpx_client(
        httpx.Client(
            base_url=BASE_URL,
            headers={"Authorization": "Bearer hx_live_test"},
            transport=httpx.MockTransport(handler),
        )
    )
    return client


def test_list_items_sends_auth_and_query_and_parses_the_page() -> None:
    seen: list[httpx.Request] = []

    def handler(request: httpx.Request) -> httpx.Response:
        seen.append(request)
        return httpx.Response(200, json={"data": [], "next_cursor": None})

    page = list_items.sync("games", client=client_with(handler), sort="-score", limit="5", min_score="8")

    assert isinstance(page, ItemPageJSON)
    assert page.data == []
    request = seen[0]
    assert request.url.path == "/v1/games"
    assert dict(request.url.params) == {"sort": "-score", "limit": "5", "min_score": "8"}
    assert request.headers["authorization"] == "Bearer hx_live_test"


def test_errors_parse_into_problem() -> None:
    problem = {
        "type": "https://hexascore.app/docs/errors#not_found",
        "title": "Not Found",
        "status": 404,
        "code": "not_found",
        "detail": "No such game.",
        "request_id": "req_1",
    }
    result = get_item.sync_detailed(
        "games",
        client=client_with(lambda request: httpx.Response(404, json=problem)),
        id="019a0000-0000-7000-8000-000000000000",
    )

    assert result.status_code == 404
    assert isinstance(result.parsed, Problem)
    assert result.parsed.detail == "No such game."
