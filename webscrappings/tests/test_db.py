"""DB-level tests. Opt-in: set TEST_DATABASE_URL to a migrated database. Everything runs in one
transaction that is rolled back, so the database is left untouched."""

import os
from collections.abc import Iterator

import psycopg
import pytest

import normalize
from common.db import upsert_source_record

URL = os.environ.get("TEST_DATABASE_URL")
pytestmark = pytest.mark.skipif(not URL, reason="TEST_DATABASE_URL not set")


@pytest.fixture
def conn() -> Iterator[psycopg.Connection]:
    assert URL
    with psycopg.connect(URL) as connection:
        yield connection
        connection.rollback()


def test_upsert_is_noop_when_payload_unchanged(conn: psycopg.Connection) -> None:
    assert upsert_source_record(conn, "test", "Q1", {"a": 1, "b": 2}) is True
    assert upsert_source_record(conn, "test", "Q1", {"b": 2, "a": 1}) is False
    assert upsert_source_record(conn, "test", "Q1", {"a": 2}) is True


def test_normalize_writes_item_translations_and_scores(
    conn: psycopg.Connection, wikidata_records: dict[str, dict]
) -> None:
    for qid, payload in wikidata_records.items():
        upsert_source_record(conn, "wikidata", qid, payload)
    qids = sorted(wikidata_records)
    assert normalize.normalize_batch(conn, qids) == (4, 0)

    row = conn.execute(
        "SELECT id, slug, type::text, meta->>'runtimeMinutes' FROM items WHERE wikidata_qid = 'Q83495'"
    ).fetchone()
    assert row is not None
    item_id, slug, item_type, runtime = row
    assert (slug, item_type, runtime) == ("the-matrix", "movie", "136")
    assert item_id.version == 7
    titles = dict(conn.execute("SELECT locale::text, title FROM item_translations WHERE item_id = %s", (item_id,)))
    assert titles["en"] == "The Matrix"
    sources = {r[0] for r in conn.execute("SELECT source::text FROM external_scores WHERE item_id = %s", (item_id,))}
    assert sources == {"metacritic", "rottentomatoes"}
    pending = conn.execute(
        "SELECT count(*) FROM source_records WHERE external_id = ANY(%s) AND normalized_at IS NULL", (qids,)
    ).fetchone()
    assert pending == (0,)

    # Re-normalizing keeps the id and slug stable.
    assert normalize.normalize_batch(conn, qids) == (4, 0)
    again = conn.execute("SELECT id, slug FROM items WHERE wikidata_qid = 'Q83495'").fetchone()
    assert again == (item_id, "the-matrix")


def test_slug_collision_appends_year(conn: psycopg.Connection, wikidata_records: dict[str, dict]) -> None:
    remake = wikidata_records["Q83495"] | {"id": "Q999000001"}
    remake["claims"] = remake["claims"] | {
        "P577": [{"value": {"time": "+2031-01-01T00:00:00Z", "precision": 11}, "rank": "normal"}]
    }
    upsert_source_record(conn, "wikidata", "Q83495", wikidata_records["Q83495"])
    upsert_source_record(conn, "wikidata", "Q999000001", remake)
    normalize.normalize_batch(conn, ["Q83495", "Q999000001"])
    slug = conn.execute("SELECT slug FROM items WHERE wikidata_qid = 'Q999000001'").fetchone()
    assert slug == ("the-matrix-2031",)
