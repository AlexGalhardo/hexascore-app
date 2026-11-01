"""Writes enrich_payloads.json: the payloads the TMDB, IGDB, Steam and Open Library connectors store, produced by the
legacy parsers over the legacy fixtures (the Go connectors land in P6.B.1). Commons is the raw API response.

Usage (from this directory):  python -X utf8 enrich_reference.py <path to legacy/webscrappings> > enrich_payloads.json
"""

import json
import sys
import types

sys.modules["psycopg"] = types.ModuleType("psycopg")
sys.modules["psycopg.types"] = types.ModuleType("psycopg.types")
jsonb = types.ModuleType("psycopg.types.json")
jsonb.Jsonb = object
sys.modules["psycopg.types.json"] = jsonb
root = sys.argv[1]
sys.path.insert(0, root)

from connectors import igdb, openlibrary, steam, tmdb  # noqa: E402


def load(name: str):
    return json.load(open(f"{root}/tests/fixtures/{name}", encoding="utf-8"))


out = {
    "tmdb:Q83495": tmdb.parse_details(load("tmdb_movie.json"), "movie", "Q83495"),
    "igdb:Q4267401": igdb.parse_games(load("igdb_games.json"), {"the-witcher-3-wild-hunt": "Q4267401"})[0],
    "steam:Q4267401": steam.parse_appdetails(load("steam_appdetails.json"), "292030", "Q4267401"),
    "openlibrary:Q208460": openlibrary.parse_doc(load("openlibrary_isbn.json"), "Q208460") | {"pages": 328},
    "commons": load("commons_imageinfo.json"),
}
print(json.dumps(out, indent=1, ensure_ascii=False))
