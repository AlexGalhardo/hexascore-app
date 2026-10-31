# OpenPencil MCP over local HTTP (token from mcp.json, never printed); one session per run.
# render: op.py render <doc> <prelude.js> <file.jsx>...   file header line: //@ page=0:7 x=0 y=4000 [replace=0:12]
#         "/*H*/" in the JSX is replaced by the prelude (shared helpers).
# export: op.py export <doc> <out-dir> <page> <name-substring>...   (every top-level <platform>/<page>/<state> frame matching)
import json, sys, urllib.request, re, time
cfg = json.load(open(r'C:\Users\aleex\AppData\Local\OpenPencil\mcp.json'))
URL = f"http://127.0.0.1:{cfg['httpPort']}/mcp"
H = {"Content-Type": "application/json", "Accept": "application/json, text/event-stream", "Authorization": "Bearer " + cfg['authToken']}
sid, n = None, 1
def post(body):
    global sid
    h = dict(H)
    if sid: h["Mcp-Session-Id"] = sid
    r = urllib.request.urlopen(urllib.request.Request(URL, json.dumps(body).encode(), h), timeout=180)
    sid = r.headers.get("Mcp-Session-Id") or sid
    t = r.read().decode()
    if "data:" in t[:50]:
        t = "\n".join(l[5:] for l in t.splitlines() if l.startswith("data:"))
    return json.loads(t) if t.strip() else None
def call(name, args):
    global n; n += 1
    r = post({"jsonrpc":"2.0","id":n,"method":"tools/call","params":{"name":name,"arguments":args}})
    if "error" in r: return "ERROR " + json.dumps(r["error"])[:300]
    return r["result"]["content"][0].get("text", "")
post({"jsonrpc":"2.0","id":0,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"designer-op","version":"1"}}})
post({"jsonrpc":"2.0","method":"notifications/initialized"})
cmd, doc = sys.argv[1], sys.argv[2]
if cmd == "render":
    prelude = open(sys.argv[3], encoding="utf-8").read()
    for f in sys.argv[4:]:
        src = open(f, encoding="utf-8").read()
        head, body = src.split("\n", 1)
        meta = dict(kv.split("=") for kv in head[3:].split())
        args = {"document_id": doc, "page_id": meta["page"], "jsx": body.replace("/*H*/", prelude)}
        call("activate_document", {"document_id": doc, "page_id": meta["page"]})
        name = re.search(r'name="([^"]+)"', body).group(1)
        for _ in range(24):  # the app stays busy for a while after an RPC timeout
            tree = call("get_page_tree", {"document_id": doc, "page_id": meta["page"], "depth": 1})
            if tree.startswith("{") and '"children"' in tree: break
            time.sleep(5)
        tree = json.loads(tree)
        hit = [c["id"] for c in tree["children"] if c["name"] == name]
        if hit: args["replace_id"] = hit[0]  # re-render in place, keeps position
        else: args["x"], args["y"] = float(meta.get("x", 0)), float(meta.get("y", 0))
        s = call("render", args)
        print(f.split("\\")[-1].split("/")[-1], s[:200].replace("\n", " "))
elif cmd == "export":
    out, page, pats = sys.argv[3], sys.argv[4], sys.argv[5:]
    call("activate_document", {"document_id": doc, "page_id": page})
    tree = json.loads(call("get_page_tree", {"document_id": doc, "page_id": page, "depth": 1}))
    for ch in tree["children"]:
        if ch["name"].count("/") != 2 or not any(p in ch["name"] for p in pats): continue
        fn = ch["name"].replace("/", "__") + ".png"
        s = call("export_image", {"document_id": doc, "page_id": page, "ids": [ch["id"]], "maxEdge": 4096, "path": out + "\\" + fn})
        print(fn, ch["id"], ch["h"], "ok" if "written" in s else "FAIL " + s[:160])
elif cmd == "py":  # op.py py <doc> <script.py>: run a script with call() and doc in scope
    exec(open(sys.argv[3], encoding="utf-8").read())
# Close the MCP session: the app answers 503 to new sessions once too many are left open.
try:
    h = dict(H); h["Mcp-Session-Id"] = sid
    urllib.request.urlopen(urllib.request.Request(URL, None, h, method="DELETE"), timeout=10)
except Exception:
    pass
