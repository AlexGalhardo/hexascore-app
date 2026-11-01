"""Runs the legacy `build_item` over the frozen parity records and writes canonical JSON (SDD 06 section 7.9).

    uv run --project ../legacy/webscrappings python scripts/parity/dump_python.py \
        --legacy ../legacy/webscrappings --data testdata/parity

The outputs in <data>/python/<QID>.json are committed, so CI needs no Python.
"""

import argparse
import json
import sys
from dataclasses import asdict, is_dataclass
from datetime import date
from pathlib import Path


def plain(value):
    if is_dataclass(value):
        return plain(asdict(value))
    if isinstance(value, date):
        return value.isoformat()
    if isinstance(value, dict):
        return {k: plain(v) for k, v in value.items()}
    if isinstance(value, (list, tuple)):
        return [plain(v) for v in value]
    return value


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("--legacy", required=True, type=Path)
    ap.add_argument("--data", required=True, type=Path)
    args = ap.parse_args()

    sys.path.insert(0, str(args.legacy.resolve()))
    import normalize  # noqa: PLC0415 (needs the path above)

    out = args.data / "python"
    out.mkdir(exist_ok=True)

    built = skipped = 0
    for path in sorted((args.data / "records").glob("*.json")):
        doc = json.loads(path.read_text(encoding="utf-8"))
        item = normalize.build_item(doc["qid"], doc["records"])
        if item is None:
            skipped += 1
            result = {"qid": doc["qid"], "item": None}
        else:
            built += 1
            year = item.release_date.year if item.release_date else None
            result = {
                "qid": doc["qid"],
                "item": plain(item),
                # First candidate: the slug a fresh database would give (no collisions in the corpus).
                "slug": normalize.slug_candidates(item.title_en, year, item.qid)[0],
            }
        (out / f"{doc['qid']}.json").write_text(
            json.dumps(result, ensure_ascii=False, indent=2, sort_keys=True) + "\n", encoding="utf-8"
        )
    print(f"built={built} skipped={skipped}")


if __name__ == "__main__":
    main()
