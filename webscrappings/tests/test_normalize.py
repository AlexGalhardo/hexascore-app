from datetime import date

import covers
import normalize
from common.db import canonical_hash
from connectors import igdb, openlibrary, steam, tmdb, wikipedia
from tests.conftest import load


def test_slugify_is_ascii_lowercase() -> None:
    assert normalize.slugify("The Witcher 3: Wild Hunt") == "the-witcher-3-wild-hunt"
    assert normalize.slugify("Amélie — Le Fabuleux Destin") == "amelie-le-fabuleux-destin"
    assert normalize.slugify("千と千尋の神隠し") == ""


def test_slug_candidates_add_year_then_qid() -> None:
    assert normalize.slug_candidates("Dune", 2021, "Q1") == ["dune", "dune-2021", "dune-q1"]
    assert normalize.slug_candidates("千と千尋", None, "Q155653") == ["q155653"]


def test_canonical_hash_ignores_key_order() -> None:
    assert canonical_hash({"a": 1, "b": [1, 2]}) == canonical_hash({"b": [1, 2], "a": 1})
    assert canonical_hash({"a": 1}) != canonical_hash({"a": 2})


def test_parse_time_precisions() -> None:
    assert normalize.parse_time({"time": "+1999-03-31T00:00:00Z", "precision": 11}) == date(1999, 3, 31)
    assert normalize.parse_time({"time": "+1999-03-00T00:00:00Z", "precision": 10}) == date(1999, 3, 1)
    assert normalize.parse_time({"time": "+1949-00-00T00:00:00Z", "precision": 9}) == date(1949, 1, 1)
    assert normalize.parse_time({"time": "+1900-00-00T00:00:00Z", "precision": 7}) is None
    assert normalize.parse_time({"time": "-0500-00-00T00:00:00Z", "precision": 9}) is None


def test_movie_mapping(wikidata_records: dict[str, dict]) -> None:
    summary = wikipedia.parse_extracts(load("wikipedia_extracts_en.json"), {"Q83495"}, "en")[0]
    item = normalize.build_item("Q83495", {"wikidata": wikidata_records["Q83495"], "wikipedia:en": summary})
    assert item is not None
    assert item.type == "movie"
    assert item.release_date == date(1999, 3, 31)
    assert item.popularity == wikidata_records["Q83495"]["sitelinks_count"]
    assert item.external_ids["imdb"] == "tt0133093"
    assert item.external_ids["tmdb"] == "603"
    assert item.meta["runtimeMinutes"] == 136
    assert {"Lana Wachowski", "Lilly Wachowski"} <= set(item.meta["creators"])
    assert "science fiction film" in item.meta["genres"]
    assert 1 <= len(item.meta["trailers"]) <= 3
    assert "youtubeId" in item.meta["trailers"][0]
    assert "platforms" not in item.meta

    en = next(t for t in item.translations if t.locale == "en")
    assert en.title == "The Matrix"
    assert en.synopsis_license == "CC BY-SA 4.0"
    assert en.synopsis_source_url == "https://en.wikipedia.org/wiki/The_Matrix"
    assert {t.locale for t in item.translations} == {"en", "pt-BR", "es"}

    by_source = {s.source: s for s in item.scores}
    assert by_source["metacritic"].value == "73/100"
    assert by_source["metacritic"].url == "https://www.metacritic.com/movie/the-matrix"
    assert by_source["rottentomatoes"].value == "83%"


def test_series_mapping(wikidata_records: dict[str, dict]) -> None:
    item = normalize.build_item("Q1079", {"wikidata": wikidata_records["Q1079"]})
    assert item is not None
    assert item.release_date == date(2008, 1, 20)
    assert item.meta["seasons"] == 5
    assert item.meta["episodes"] == 62
    assert item.external_ids["tmdb"] == "1396"


def test_game_mapping_with_steam_enrichment(wikidata_records: dict[str, dict]) -> None:
    steam_payload = steam.parse_appdetails(load("steam_appdetails.json"), "292030", "Q4267401")
    records = {"wikidata": wikidata_records["Q4267401"], "steam": steam_payload}
    item = normalize.build_item("Q4267401", records)
    assert item is not None
    assert item.external_ids["steam"] == "292030"
    assert item.external_ids["gog"] == "game/the_witcher_3_wild_hunt"
    assert "Microsoft Windows" in item.meta["platforms"]  # Wikidata wins over Steam
    assert item.meta["creators"] == ["CD Projekt RED"]
    assert "runtimeMinutes" not in item.meta
    assert covers.pick_source_cover("game", records) == covers.Cover(steam_payload["header_image"])


def test_book_mapping_uses_open_library_pages_as_fallback(wikidata_records: dict[str, dict]) -> None:
    ol = openlibrary.parse_doc(load("openlibrary_isbn.json"), "Q208460") | {"pages": 328}
    records = {"wikidata": wikidata_records["Q208460"], "openlibrary": ol}
    item = normalize.build_item("Q208460", records)
    assert item is not None
    assert item.release_date is not None and item.release_date.year == 1949
    assert item.meta["creators"] == ["George Orwell"]
    assert item.meta["pages"] == 328
    assert covers.pick_source_cover("book", records) == covers.Cover(
        "https://covers.openlibrary.org/b/id/15257377-L.jpg"
    )


def test_tmdb_fills_only_what_wikidata_lacks(wikidata_records: dict[str, dict]) -> None:
    tmdb_payload = tmdb.parse_details(load("tmdb_movie.json"), "movie", "Q83495")
    wd = wikidata_records["Q83495"]
    item = normalize.build_item("Q83495", {"wikidata": wd, "tmdb": tmdb_payload})
    assert item is not None
    assert "science fiction film" in item.meta["genres"]  # Wikidata wins
    assert "Science Fiction" not in item.meta["genres"]

    bare = wd | {"claims": {k: v for k, v in wd["claims"].items() if k not in ("P136", "P2047", "P577", "P4947")}}
    item = normalize.build_item("Q83495", {"wikidata": bare, "tmdb": tmdb_payload})
    assert item is not None
    assert item.meta["genres"] == tmdb_payload["genres"]
    assert item.meta["runtimeMinutes"] == 136
    assert item.release_date == date(1999, 3, 31)
    assert item.external_ids["tmdb"] == "603"
    record = {"tmdb": tmdb_payload}
    assert covers.pick_source_cover("movie", record) == covers.Cover(
        tmdb_payload["cover"], {"source": "TMDB", "url": "https://www.themoviedb.org/movie/603"}
    )
    assert covers.pick_source_cover("game", record) is None


def test_igdb_cover_wins_over_steam(wikidata_records: dict[str, dict]) -> None:
    igdb_payload = igdb.parse_games(load("igdb_games.json"), {"the-witcher-3-wild-hunt": "Q4267401"})[0]
    steam_payload = steam.parse_appdetails(load("steam_appdetails.json"), "292030", "Q4267401")
    records = {"wikidata": wikidata_records["Q4267401"], "steam": steam_payload, "igdb": igdb_payload}
    cover = covers.pick_source_cover("game", records)
    assert cover is not None
    assert cover.url == igdb_payload["cover"]
    assert cover.attribution == {"source": "IGDB", "url": igdb_payload["url"]}

    bare = records["wikidata"] | {"claims": {}}
    item = normalize.build_item("Q4267401", records | {"wikidata": bare})
    assert item is not None
    assert "PlayStation 4" in item.meta["platforms"]  # IGDB before Steam's OS list
    assert item.meta["creators"] == ["CD Projekt RED"]
    assert item.release_date == date(2015, 5, 19)


def test_own_synopsis_only_fills_locales_without_wikipedia(wikidata_records: dict[str, dict]) -> None:
    summary = wikipedia.parse_extracts(load("wikipedia_extracts_en.json"), {"Q83495"}, "en")[0]
    own = {"en": "Own English text.", "pt-BR": "Texto próprio.", "es": ""}
    records = {"wikidata": wikidata_records["Q83495"], "wikipedia:en": summary, "synopsis": own}
    item = normalize.build_item("Q83495", records)
    assert item is not None
    by_locale = {t.locale: t for t in item.translations}
    assert by_locale["en"].synopsis == summary["extract"]
    assert by_locale["pt-BR"].synopsis == "Texto próprio."
    assert by_locale["pt-BR"].synopsis_license is None
    assert by_locale["pt-BR"].synopsis_source_url is None
    assert by_locale["es"].synopsis is None


def test_items_without_wikidata_or_english_title_are_skipped(wikidata_records: dict[str, dict]) -> None:
    assert normalize.build_item("Q1", {"steam": {"platforms": []}}) is None
    payload = wikidata_records["Q83495"] | {"labels": {"pt": "Matrix"}, "sitelinks": {}}
    assert normalize.build_item("Q83495", {"wikidata": payload}) is None


def test_commons_cover_attribution() -> None:
    cover = covers.parse_commons(load("commons_imageinfo.json"))
    assert cover is not None
    assert cover.url.startswith("https://")
    assert cover.attribution is not None
    assert cover.attribution["source"] == "Wikimedia Commons"
    assert cover.attribution["license"]
    assert "<" not in cover.attribution["author"]
