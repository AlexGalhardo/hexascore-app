import httpx
import pytest

import normalize
import trailers
from tests.conftest import mock_client

GOOD_A, GOOD_B, GOOD_C = "aaaaaaaaaaa", "bbbbbbbbbbb", "ccccccccccc"
GONE, PRIVATE = "ddddddddddd", "eeeeeeeeeee"
STATUS = {GONE: 404, PRIVATE: 401}


def claim(value: str, rank: str = "normal") -> dict:
    return {"value": value, "rank": rank, "qualifiers": {}}


def oembed(calls: list[str]):
    def handler(request: httpx.Request) -> httpx.Response:
        assert request.url.host == "www.youtube.com"
        youtube_id = request.url.params["url"].rsplit("=", 1)[1]
        calls.append(youtube_id)
        code = STATUS.get(youtube_id, 200)
        return httpx.Response(code, json={"title": "Trailer"} if code == 200 else None)

    return handler


def test_candidates_order_wikidata_first_and_drop_malformed_ids() -> None:
    records = {
        "wikidata": [claim(GOOD_B), claim(GOOD_A, "preferred"), claim("not-an-id")],
        "tmdb": [GOOD_C, GOOD_A, "x" * 12],
    }
    assert trailers.candidates(records) == [GOOD_A, GOOD_B, GOOD_C]
    assert trailers.candidates({"igdb": None, "wikidata": None}) == []


def test_check_keeps_two_available_and_stops_asking() -> None:
    calls: list[str] = []
    result = trailers.check(mock_client(oembed(calls)), [GONE, GOOD_A, PRIVATE, GOOD_B, GOOD_C], trailers.Checked())
    assert result.available == [GOOD_A, GOOD_B]
    assert result.unavailable == [GONE, PRIVATE]
    assert calls == [GONE, GOOD_A, PRIVATE, GOOD_B]  # the third good one is never requested


def test_check_reuses_known_verdicts() -> None:
    calls: list[str] = []
    known = trailers.Checked(available=[GOOD_A], unavailable=[GONE])
    result = trailers.check(mock_client(oembed(calls)), [GONE, GOOD_A, GOOD_B], known)
    assert result.available == [GOOD_A, GOOD_B]
    assert calls == [GOOD_B]


def test_rate_limit_is_not_a_verdict() -> None:
    client = mock_client(lambda request: httpx.Response(429))
    with pytest.raises(httpx.HTTPStatusError):
        trailers.is_available(client, GOOD_A)


def test_fetch_writes_one_record_per_item_and_skips_failures(monkeypatch: pytest.MonkeyPatch) -> None:
    rows = [
        ("Q1", {"wikidata": [claim(GONE)], "tmdb": [GOOD_A]}),
        ("Q2", {"wikidata": None, "igdb": None}),  # no candidates, never checked: nothing to write
        ("Q3", {"tmdb": [GONE], "youtube": {"available": [GONE], "unavailable": []}}),
        ("Q4", {"tmdb": ["fffffffffff"]}),
    ]
    monkeypatch.setattr(trailers, "targets", lambda conn, since: rows)
    monkeypatch.setitem(STATUS, "fffffffffff", 500)
    calls: list[str] = []
    client = mock_client(oembed(calls))

    records = list(trailers.fetch(None, conn=None, client=client, workers=1))  # type: ignore[arg-type]
    assert [(r.source, r.external_id) for r in records] == [("youtube", "Q1"), ("youtube", "Q3")]
    assert records[0].payload == {"qid": "Q1", "available": [GOOD_A], "unavailable": [GONE]}
    # A full run asks again, so a video that went away since the last check is dropped.
    assert records[1].payload == {"qid": "Q3", "available": [], "unavailable": [GONE]}

    calls.clear()
    reused = list(trailers.fetch(None, conn=None, client=client, only_missing=True, workers=1))  # type: ignore[arg-type]
    assert reused[1].payload["available"] == [GONE]
    assert GONE not in calls[1:]  # Q3's stored verdict is reused; Q1's unchecked ids are asked


def test_normalize_shows_only_checked_trailers_and_at_most_two(wikidata_records: dict[str, dict]) -> None:
    wd = wikidata_records["Q83495"]
    wd = wd | {"claims": wd["claims"] | {"P1651": [claim(GONE), claim(GOOD_A)]}}
    tmdb = {"trailers": [GOOD_B, GOOD_C]}

    unchecked = normalize.build_item("Q83495", {"wikidata": wd, "tmdb": tmdb})
    assert unchecked is not None
    assert unchecked.meta["trailers"] == [{"youtubeId": GONE}, {"youtubeId": GOOD_A}]

    youtube = {"available": [GOOD_A, GOOD_B], "unavailable": [GONE]}
    checked = normalize.build_item("Q83495", {"wikidata": wd, "tmdb": tmdb, "youtube": youtube})
    assert checked is not None
    assert checked.meta["trailers"] == [{"youtubeId": GOOD_A}, {"youtubeId": GOOD_B}]

    none_left = normalize.build_item("Q83495", {"wikidata": wd, "youtube": {"available": [], "unavailable": [GONE]}})
    assert none_left is not None
    assert "trailers" not in none_left.meta
