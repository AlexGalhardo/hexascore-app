import httpx
import pytest
from pydantic import ValidationError

from connectors import gog, openlibrary, steam, wikidata, wikipedia
from tests.conftest import load, mock_client, wikidata_handler


def test_select_qids_parses_sparql_bindings() -> None:
    qids = wikidata.select_qids(mock_client(wikidata_handler), "movie", None, 3)
    assert len(qids) == 3
    assert all(q.startswith("Q") for q in qids)


def test_queries_target_type_classes() -> None:
    assert "wd:Q7889" in wikidata.top_query("game", 10)
    assert "LIMIT 10" in wikidata.top_query("game", 10)


def test_trim_entity_keeps_only_used_data(wikidata_records: dict[str, dict]) -> None:
    matrix = wikidata_records["Q83495"]
    assert matrix["type"] == "movie"
    assert matrix["labels"]["en"] == "The Matrix"
    assert matrix["sitelinks"]["enwiki"] == "The Matrix"
    assert matrix["sitelinks_count"] > 50
    assert set(matrix["claims"]) <= wikidata.KEPT_PROPS
    assert matrix["claims"]["P345"][0]["value"] == "tt0133093"
    assert matrix["claims"]["P2047"][0]["value"] == {"amount": "+136", "unit": "Q7727"}
    assert {"Q9544977", "Q9545711"} <= set(matrix["refs"])  # Wachowskis: one only has a "mul" label


def test_score_statements_keep_their_qualifiers(wikidata_records: dict[str, dict]) -> None:
    scores = wikidata_records["Q83495"]["claims"]["P444"]
    metascore = next(s for s in scores if s["value"] == "73/100")
    assert metascore["qualifiers"]["P447"] == ["Q150248"]


def test_missing_entities_are_skipped() -> None:
    records = list(wikidata.fetch_entities(mock_client(wikidata_handler), [("Q999999999", "movie")]))
    assert records == []


def test_invalid_entity_payload_is_rejected() -> None:
    with pytest.raises(ValidationError):
        wikidata.RawEntity.model_validate({"id": "Q1", "claims": {"P31": [{"rank": "weird"}]}})


def test_wikipedia_extracts() -> None:
    payloads = wikipedia.parse_extracts(load("wikipedia_extracts_en.json"), {"Q83495", "Q189600"}, "en")
    assert [p["qid"] for p in payloads] == ["Q83495", "Q189600"]
    assert payloads[0]["url"] == "https://en.wikipedia.org/wiki/The_Matrix"
    assert payloads[0]["extract"].startswith("The Matrix")


def test_wikipedia_rejects_disambiguation_missing_and_other_entities() -> None:
    # Q190069 is the "Matrix" disambiguation page; Q1 matches nothing; the missing page has no pageprops.
    assert wikipedia.parse_extracts(load("wikipedia_extracts_en.json"), {"Q190069", "Q1"}, "en") == []


def test_openlibrary_paths_prefer_editions() -> None:
    assert openlibrary.document_path({"isbn13": "978-0-451-52493-5"}) == "/isbn/9780451524935.json"
    assert openlibrary.document_path({"openlibrary": "OL1M"}) == "/books/OL1M.json"
    assert openlibrary.document_path({"openlibrary": "OL1168083W"}) == "/works/OL1168083W.json"
    assert openlibrary.document_path({"openlibrary": "OL23A"}) is None


def test_openlibrary_edition() -> None:
    payload = openlibrary.parse_doc(load("openlibrary_isbn.json"), "Q208460")
    assert payload == {"qid": "Q208460", "key": "/books/OL62592264M", "pages": None, "covers": [15257377]}


def test_openlibrary_drops_placeholder_covers() -> None:
    payload = openlibrary.parse_doc({"key": "/works/OL1W", "covers": [-1, 5], "number_of_pages": 320}, "Q1")
    assert payload["covers"] == [5]
    assert payload["pages"] == 320


def test_steam_appdetails() -> None:
    payload = steam.parse_appdetails(load("steam_appdetails.json"), "292030", "Q4267401")
    assert payload is not None
    assert payload["platforms"] == ["Windows"]
    assert payload["header_image"].startswith("https://")
    assert steam.parse_appdetails(load("steam_appdetails_fail.json"), "1", "Q1") is None


def test_gog_picks_product_by_slug_prefix() -> None:
    payload = gog.parse_catalog(
        load("gog_catalog.json"), "game/the_witcher_3_wild_hunt", "The Witcher 3: Wild Hunt", "Q4267401"
    )
    assert payload is not None
    assert payload["slug"] == "the_witcher_3_wild_hunt_game_of_the_year_edition"
    assert payload["store_url"].startswith("https://www.gog.com/")
    assert payload["platforms"] == ["Windows"]


def test_gog_no_match() -> None:
    assert gog.parse_catalog(load("gog_catalog.json"), "game/portal", "Portal", "Q1") is None


def test_enrichment_skips_failed_targets() -> None:
    class FakeConn:
        def execute(self, *_: object) -> FakeConn:
            return self

        def fetchall(self) -> list[tuple[str, str]]:
            return [("Q1", "10"), ("Q2", "292030")]

    def handler(request: httpx.Request) -> httpx.Response:
        if request.url.params["appids"] == "10":
            return httpx.Response(403)
        return httpx.Response(200, json=load("steam_appdetails.json"))

    records = list(steam.fetch(None, FakeConn(), mock_client(handler)))
    assert [r.external_id for r in records] == ["Q2"]


def test_wikipedia_fetch_runs_requests_concurrently(monkeypatch) -> None:
    import threading
    import time

    import httpx

    from common.http import HttpClient

    active = {"now": 0, "max": 0}
    lock = threading.Lock()
    batch_sizes: list[int] = []

    def handler(request: httpx.Request) -> httpx.Response:
        titles = request.url.params["titles"].split("|")
        with lock:
            batch_sizes.append(len(titles))
            active["now"] += 1
            active["max"] = max(active["max"], active["now"])
        time.sleep(0.02)
        with lock:
            active["now"] -= 1
        pages = [
            {
                "title": t,
                "extract": f"{t} intro.",
                "fullurl": f"https://en.wikipedia.org/wiki/{t}",
                "pageprops": {"wikibase_item": f"Q{t[1:]}"},
            }
            for t in titles
        ]
        return httpx.Response(200, json={"query": {"pages": pages}})

    # Title "T<n>" belongs to Q<n>; 85 targets -> batches of 20, 20, 20, 20, 5.
    rows = [(f"Q{n}", {"enwiki": f"T{n}"}) for n in range(85)]
    monkeypatch.setattr(wikipedia, "targets", lambda conn, since: rows)
    monkeypatch.setattr(wikipedia, "existing_ids", lambda conn, source: {"Q0"} if source == "wikipedia:en" else set())
    client = HttpClient(transport=httpx.MockTransport(handler), sleep=lambda _: None)
    monkeypatch.setitem(__import__("common.http").http.HOST_INTERVALS, "en.wikipedia.org", 0.0)

    records = list(
        wikipedia.fetch(None, conn=None, client=client, only_missing=True, langs=["en"], workers=4)  # type: ignore[arg-type]
    )
    assert sorted(r.external_id for r in records) == sorted(f"Q{n}" for n in range(1, 85))
    assert sorted(batch_sizes) == [4, 20, 20, 20, 20]
    assert active["max"] > 1
