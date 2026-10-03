"""Full catalog load: top items by Wikidata sitelinks per type, enrichment, normalization, covers, snapshot.

The snapshot is a gzipped psql script (COPY ... FROM stdin) built with psycopg's COPY, so pg_dump is not
needed. Restore with: gunzip -c snapshot.sql.gz | psql "$DATABASE_URL"
"""

import argparse
import gzip
import logging
from pathlib import Path

import psycopg
from psycopg import sql

import covers
from common.cli import setup_logging
from common.db import connect, store_records
from connectors import gog, openlibrary, steam, wikidata, wikipedia
from normalize import normalize
from revalidate import revalidate

log = logging.getLogger("bootstrap")

ENRICHMENT = {"wikipedia": wikipedia, "openlibrary": openlibrary, "steam": steam, "gog": gog}
# FK order: items first.
SNAPSHOT_TABLES = ("items", "item_translations", "external_scores")


def export_snapshot(conn: psycopg.Connection, path: Path) -> dict[str, int]:
    counts: dict[str, int] = {}
    with gzip.open(path, "wb") as out:
        out.write(b"BEGIN;\n")
        for table in SNAPSHOT_TABLES:
            columns = [
                r[0]
                for r in conn.execute(
                    """SELECT column_name FROM information_schema.columns
                    WHERE table_schema = 'public' AND table_name = %s AND is_generated = 'NEVER'
                    ORDER BY ordinal_position""",
                    (table,),
                )
            ]
            cols = sql.SQL(", ").join(map(sql.Identifier, columns))
            header = sql.SQL("COPY {} ({}) FROM stdin;\n").format(sql.Identifier(table), cols).as_string(conn)
            out.write(header.encode())
            with conn.cursor() as cur:
                copy_sql = sql.SQL("COPY {} ({}) TO STDOUT").format(sql.Identifier(table), cols)
                with cur.copy(copy_sql) as copy:
                    for chunk in copy:
                        out.write(chunk)
            out.write(b"\\.\n")
            counts[table] = conn.execute(sql.SQL("SELECT count(*) FROM {}").format(sql.Identifier(table))).fetchone()[0]
        out.write(b"COMMIT;\n")
    return counts


def main() -> None:
    setup_logging()
    p = argparse.ArgumentParser(description="Bootstrap the catalog from Wikidata and enrichment sources")
    p.add_argument("--types", default=",".join(wikidata.TYPES), help="comma-separated item types")
    p.add_argument("--limit", type=int, help="items per type (default: ~35k movies, 15k series, 25k games, 25k books)")
    p.add_argument("--sources", default=",".join(ENRICHMENT), help="enrichment sources to run")
    p.add_argument("--skip-covers", action="store_true")
    p.add_argument("--export", type=Path, help="write a gzipped SQL snapshot of the catalog tables")
    p.add_argument("--export-only", action="store_true", help="skip ingestion, only export")
    args = p.parse_args()

    with connect() as conn:
        if not args.export_only:
            types = [t for t in args.types.split(",") if t in wikidata.TYPES]
            seen, changed = store_records(conn, wikidata.fetch(None, types=types, limit=args.limit))
            log.info("wikidata fetched=%d changed=%d", seen, changed)
            normalize(conn)
            for name in args.sources.split(","):
                if name in ENRICHMENT:
                    seen, changed = store_records(conn, ENRICHMENT[name].fetch(None, conn, only_missing=True))
                    log.info("%s fetched=%d changed=%d", name, seen, changed)
            normalize(conn)
            if not args.skip_covers:
                log.info("covers updated=%d", covers.run(conn))
            revalidate(["catalog"])
        if args.export:
            log.info("snapshot %s: %s", args.export, export_snapshot(conn, args.export))


if __name__ == "__main__":
    main()
