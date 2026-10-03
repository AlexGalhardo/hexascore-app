import hashlib
import json
import logging
import os
import threading
from collections.abc import Iterable
from dataclasses import dataclass
from itertools import batched
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


KEEPALIVE_SECONDS = 60
log = logging.getLogger(__name__)
_BUSY: dict[int, threading.Lock] = {}


def busy(conn: psycopg.Connection) -> threading.Lock:
    """Held around multi-statement work (pipeline mode, executemany). psycopg's own lock only covers single calls,
    so a keep-alive ping sent between two pipelined statements fails the batch ("another command is already in
    progress")."""
    return _BUSY.setdefault(id(conn), threading.Lock())


def keep_alive(conn: psycopg.Connection, interval: float = KEEPALIVE_SECONDS) -> threading.Thread:
    """Pings the connection while it is open. Fetchers spend minutes on HTTP between writes (a SPARQL
    selection, rate-limited enrichment) and Neon drops idle connections, so the next INSERT would fail.
    A connection held in `busy` is in use, so that ping is skipped."""

    def ping() -> None:
        lock = busy(conn)
        while not conn.closed:
            threading.Event().wait(interval)
            if conn.closed:
                return
            if not lock.acquire(blocking=False):
                continue
            try:
                conn.execute("SELECT 1")
            except psycopg.Error as error:
                log.warning("keep-alive ping failed: %s", error)
                return
            finally:
                lock.release()

    thread = threading.Thread(target=ping, name="db-keepalive", daemon=True)
    thread.start()
    return thread


def connect() -> psycopg.Connection:
    url = os.environ.get("DATABASE_URL")
    if not url:
        raise SystemExit("DATABASE_URL is not set")
    conn = psycopg.connect(url, keepalives=1, keepalives_idle=30)
    keep_alive(conn)
    return conn


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
    # executemany pipelines each chunk in one round trip; one upsert per record capped a remote DB at ~6 rows/s.
    for chunk in batched(records, commit_every, strict=False):
        params = [(r.source, r.external_id, Jsonb(r.payload), canonical_hash(r.payload)) for r in chunk]
        with busy(conn), conn.cursor() as cur:
            cur.executemany(UPSERT_SQL, params, returning=True)
            while True:
                changed += cur.fetchone() is not None
                if not cur.nextset():
                    break
        conn.commit()
        seen += len(chunk)
    return seen, changed
