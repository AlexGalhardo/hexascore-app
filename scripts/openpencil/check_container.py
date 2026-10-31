# Desktop container check (owner item 9): every node inside a 1440 frame must sit between the
# navbar edges (x 144 to 1296). Full-width bands (x 0, width 1440: header, footer, page wrapper)
# and the insides of instances are exempt. Run: op.py py <doc> check_container.py  (env PAGES=0:7,...)
import json, os

L, R = 144, 1296
pages = os.environ.get("PAGES", "0:7").split(",")


def walk(nd, ax, path, bad):
    x = ax + (nd.get("x") or 0)
    w = nd.get("width") or 0
    band = abs(x) < 0.5 and abs(w - 1440) < 0.5
    if not band and nd.get("type") not in ("VECTOR",) and (x < L - 0.5 or x + w > R + 0.5):
        bad.append("%s [%s] x=%d..%d" % (path + "/" + nd.get("name", "?"), nd.get("type"), x, x + w))
        return
    if nd.get("type") == "INSTANCE":
        return
    for c in nd.get("children", []) or []:
        if isinstance(c, dict):
            walk(c, x, path + "/" + nd.get("name", "?"), bad)


for p in pages:
    t = json.loads(call("get_page_tree", {"document_id": doc, "page_id": p, "depth": 1}))
    for fr in t["children"]:
        if not fr["name"].startswith("web-desktop"):
            continue
        root = json.loads(call("get_node", {"document_id": doc, "page_id": p, "id": fr["id"], "depth": 12}))
        bad = []
        for c in root.get("children", []):
            if isinstance(c, dict):
                walk(c, 0, "", bad)
        print(("OK   " if not bad else "FAIL ") + fr["name"] + ("" if not bad else "  " + "; ".join(bad[:3])))
