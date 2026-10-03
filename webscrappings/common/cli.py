import argparse
import logging
import re
from collections.abc import Callable, Iterable
from datetime import UTC, datetime, timedelta

from common.db import Record, connect, store_records


def parse_since(value: str) -> datetime:
    """Accepts a relative window (`36h`, `8d`) or an ISO 8601 timestamp (UTC when naive)."""
    if match := re.fullmatch(r"(\d+)([hd])", value):
        amount, unit = int(match[1]), match[2]
        return datetime.now(UTC) - timedelta(hours=amount if unit == "h" else amount * 24)
    parsed = datetime.fromisoformat(value)
    return parsed if parsed.tzinfo else parsed.replace(tzinfo=UTC)


def setup_logging() -> None:
    logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(name)s: %(message)s")
    logging.getLogger("httpx").setLevel(logging.WARNING)


def parser(description: str) -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(description=description)
    p.add_argument("--since", type=parse_since, help="only refresh what changed since (e.g. 36h, 8d, 2026-01-01)")
    p.add_argument("--missing-only", action="store_true", help="skip targets that already have a record (resume)")
    return p


def run_enrichment(
    description: str,
    fetch: Callable[..., Iterable[Record]],
) -> None:
    setup_logging()
    args = parser(description).parse_args()
    with connect() as conn:
        seen, changed = store_records(conn, fetch(args.since, conn, only_missing=args.missing_only))
    logging.getLogger(description).info("records fetched=%d changed=%d", seen, changed)
