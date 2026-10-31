"""Reference hashes for TestTrimHashMatchesPython.

Runs the legacy `trim_entity` (legacy/webscrappings/connectors/wikidata.py) over wikidata_entities.json
with the three deliberate changes of the Go port applied on top:
  - P444 (review scores) is not kept (D-09),
  - aliases of the kept languages are added,
  - P144 and the SDD 06 7.4 role properties are kept as reference properties.
The "qualifiers" key (always empty once P444 is gone) is dropped from claims.

Usage (from this directory):  python -X utf8 trim_reference.py <path to legacy/webscrappings> > trim_hashes.json
"""

import hashlib
import json
import sys
import types

sys.modules["psycopg"] = types.ModuleType("psycopg")
sys.modules["psycopg.types"] = types.ModuleType("psycopg.types")
jsonb = types.ModuleType("psycopg.types.json")
jsonb.Jsonb = object
sys.modules["psycopg.types.json"] = jsonb
sys.path.insert(0, sys.argv[1])

from connectors import wikidata as w  # noqa: E402

EXTRA_REFS = ("P144", "P58", "P161", "P272", "P449", "P404", "P179", "P8345", "P364", "P407", "P123")
REF_PROPS = w.REF_PROPS + EXTRA_REFS
SCALAR_PROPS = tuple(p for p in w.SCALAR_PROPS if p != "P444")
KEPT = frozenset(REF_PROPS + SCALAR_PROPS)

TYPES = {"Q83495": "movie", "Q4267401": "game", "Q1079": "series", "Q208460": "book"}
entities = json.load(open("wikidata_entities.json", encoding="utf-8"))["entities"]
labels = json.load(open("wikidata_labels.json", encoding="utf-8"))["entities"]
everything = {**labels, **entities}


def trim(raw: dict, item_type: str) -> dict:
    entity = w.RawEntity.model_validate(raw)
    claims = {}
    for prop, statements in entity.claims.items():
        if prop not in KEPT:
            continue
        kept = []
        for st in statements:
            value = w.simplify_value(st.mainsnak.datavalue)
            if st.rank == "deprecated" or value is None:
                continue
            kept.append({"value": value, "rank": st.rank})
        if kept:
            claims[prop] = kept
    aliases = {}
    for lang in w.LABEL_LANGS:
        values = [a["value"] for a in raw.get("aliases", {}).get(lang, [])]
        if values:
            aliases[lang] = values
    payload = {
        "type": item_type,
        "id": entity.id,
        "labels": {lang: entity.labels[lang].value for lang in w.LABEL_LANGS if lang in entity.labels},
        "aliases": aliases,
        "sitelinks": {s: entity.sitelinks[s].title for s in w.SITELINKS if s in entity.sitelinks},
        "sitelinks_count": len(entity.sitelinks),
        "claims": claims,
    }
    refs = {}
    for qid in sorted({c["value"] for p in REF_PROPS for c in claims.get(p, []) if isinstance(c["value"], str)}):
        ref = everything.get(qid)
        if ref is None or "missing" in ref:
            continue
        e = w.RawEntity.model_validate(ref)
        label = e.labels.get("en") or e.labels.get("mul")
        if label and label.value:
            refs[qid] = label.value
    payload["refs"] = refs
    return payload


out = {}
for qid, item_type in TYPES.items():
    out[qid] = hashlib.sha256(
        json.dumps(trim(entities[qid], item_type), sort_keys=True, separators=(",", ":"), ensure_ascii=False).encode()
    ).hexdigest()
print(json.dumps(out, indent=2))
