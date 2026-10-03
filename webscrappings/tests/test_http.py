import json

import httpx
import pytest

from common.cli import parse_since
from common.http import HttpClient, retry_after_seconds, user_agent
from revalidate import revalidate


def test_user_agent_identifies_bot(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("INGEST_CONTACT_URL", "https://example.org/contact")
    assert user_agent() == "HexascoreBot/0.1 (+https://example.org/contact)"


def test_retry_after_parsing() -> None:
    assert retry_after_seconds("7", now=0) == 7
    assert retry_after_seconds(None, now=0) is None
    assert retry_after_seconds("Thu, 01 Jan 1970 00:00:30 GMT", now=10) == 20
    assert retry_after_seconds("garbage", now=0) is None


def test_retries_429_honoring_retry_after_then_succeeds() -> None:
    calls: list[int] = []
    sleeps: list[float] = []

    def handler(request: httpx.Request) -> httpx.Response:
        calls.append(1)
        if len(calls) == 1:
            return httpx.Response(429, headers={"Retry-After": "12"})
        if len(calls) == 2:
            return httpx.Response(503)
        return httpx.Response(200, json={"ok": True})

    client = HttpClient(transport=httpx.MockTransport(handler), sleep=sleeps.append, clock=lambda: 1e9)
    assert client.get_json("https://example.org/x") == {"ok": True}
    assert len(calls) == 3
    assert 12 in sleeps
    assert 4 in sleeps  # exponential backoff on the second attempt


def test_gives_up_after_max_attempts() -> None:
    client = HttpClient(
        transport=httpx.MockTransport(lambda r: httpx.Response(500)), sleep=lambda _: None, max_attempts=2
    )
    with pytest.raises(httpx.HTTPStatusError):
        client.get_json("https://example.org/x")


def test_does_not_retry_client_errors() -> None:
    calls: list[int] = []

    def handler(request: httpx.Request) -> httpx.Response:
        calls.append(1)
        return httpx.Response(404)

    client = HttpClient(transport=httpx.MockTransport(handler), sleep=lambda _: None)
    assert client.request("GET", "https://example.org/x").status_code == 404
    assert len(calls) == 1


def test_rate_limit_spaces_requests_per_host() -> None:
    now = [0.0]
    sleeps: list[float] = []

    def sleep(seconds: float) -> None:
        sleeps.append(seconds)
        now[0] += seconds

    client = HttpClient(
        transport=httpx.MockTransport(lambda r: httpx.Response(200, json={})), sleep=sleep, clock=lambda: now[0]
    )
    client.get_json("https://query.wikidata.org/a")
    client.get_json("https://query.wikidata.org/b")
    client.get_json("https://other.example.org/c")
    assert sleeps == [1.0]


def test_parse_since() -> None:
    assert parse_since("2026-01-01").isoformat() == "2026-01-01T00:00:00+00:00"
    delta = parse_since("1d") - parse_since("25h")
    assert round(delta.total_seconds()) == 3600


def test_revalidate_posts_tags_with_bearer(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("APP_URL", "https://app.example.org/")
    monkeypatch.setenv("CRON_SECRET", "s3cret")
    seen: list[httpx.Request] = []

    def handler(request: httpx.Request) -> httpx.Response:
        seen.append(request)
        return httpx.Response(200)

    client = HttpClient(transport=httpx.MockTransport(handler), sleep=lambda _: None)
    assert revalidate(["catalog"], client) is True
    assert str(seen[0].url) == "https://app.example.org/api/cron/revalidate"
    assert seen[0].headers["Authorization"] == "Bearer s3cret"
    assert json.loads(seen[0].content) == {"tags": ["catalog"]}


def test_revalidate_failures_are_not_fatal(monkeypatch: pytest.MonkeyPatch) -> None:
    monkeypatch.setenv("APP_URL", "https://app.example.org")
    monkeypatch.setenv("CRON_SECRET", "s3cret")

    def down(request: httpx.Request) -> httpx.Response:
        raise httpx.ConnectError("down")

    unreachable = HttpClient(transport=httpx.MockTransport(down), sleep=lambda _: None, max_attempts=1)
    assert revalidate(["catalog"], unreachable) is False
    unauthorized = HttpClient(transport=httpx.MockTransport(lambda r: httpx.Response(401)), sleep=lambda _: None)
    assert revalidate(["catalog"], unauthorized) is False
    monkeypatch.delenv("APP_URL")
    assert revalidate(["catalog"], unauthorized) is False
