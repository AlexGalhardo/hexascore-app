import hashlib
import json
import os
from collections.abc import Iterable
from dataclasses import dataclass
from typing import Any

import psycopg
from psycopg.types.json import Jsonb


@dataclass(frozen=True, slots=True)
class Record:
    """Raw payload from one source. `external_id` is always the Wikidata QID, so every source
    shares the (source, external_id) primary key as its match key."""

    source: str
    external_id: str
    payload: dict[str, Any]


def canonical_hash(payload: Any) -> str:
    canonical = json.dumps(payload, sort_keys=True, separators=(",", ":"), ensure_ascii=False)
    return hashlib.sha256(canonical.encode()).hexdigest()


def connect() -> psycopg.Connection:
    url = os.environ.get("DATABASE_URL")
    if not url:
        raise SystemExit("DATABASE_URL is not set")
    return psycopg.connect(url)


UPSERT_SQL = """
INSERT INTO source_records (source, external_id, payload, hash, fetched_at)
VALUES (%s, %s, %s, %s, now())
ON CONFLICT (source, external_id) DO UPDATE
SET payload = excluded.payload, hash = excluded.hash, fetched_at = excluded.fetched_at
WHERE source_records.hash <> excluded.hash
RETURNING 1
"""


def upsert_source_record(conn: psycopg.Connection, source: str, external_id: str, payload: dict[str, Any]) -> bool:
    """Returns True when the row was inserted or its payload changed; unchanged payloads are a no-op."""
    row = conn.execute(UPSERT_SQL, (source, external_id, Jsonb(payload), canonical_hash(payload))).fetchone()
    return row is not None


def existing_ids(conn: psycopg.Connection, source: str) -> set[str]:
    return {r[0] for r in conn.execute("SELECT external_id FROM source_records WHERE source = %s", (source,))}


def store_records(conn: psycopg.Connection, records: Iterable[Record], commit_every: int = 200) -> tuple[int, int]:
    seen = changed = 0
    for record in records:
        seen += 1
        changed += upsert_source_record(conn, record.source, record.external_id, record.payload)
        if seen % commit_every == 0:
            conn.commit()
    conn.commit()
    return seen, changed
