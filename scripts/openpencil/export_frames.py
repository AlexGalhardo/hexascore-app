# Calls OpenPencil MCP set_image_fill over local HTTP; token read from mcp.json, never printed.
import json, sys, urllib.request
cfg = json.load(open(r'C:\Users\aleex\AppData\Local\OpenPencil\mcp.json'))
URL = f"http://127.0.0.1:{cfg['httpPort']}/mcp"
H = {"Content-Type": "application/json", "Accept": "application/json, text/event-stream", "Authorization": "Bearer " + cfg['authToken']}
sid = None
def post(body):
    global sid
    h = dict(H)
    if sid: h["Mcp-Session-Id"] = sid
    r = urllib.request.urlopen(urllib.request.Request(URL, json.dumps(body).encode(), h), timeout=120)
    sid = r.headers.get("Mcp-Session-Id") or sid
    t = r.read().decode()
    if "data:" in t[:50]:
        t = "\n".join(l[5:] for l in t.splitlines() if l.startswith("data:"))
    return json.loads(t) if t.strip() else None
post({"jsonrpc":"2.0","id":0,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"designer-script","version":"1"}}})
post({"jsonrpc":"2.0","method":"notifications/initialized"})
n = 1
# usage: opexport2.py <doc> <root-dir> <pageId>... ; exports every top-level frame as <platform>__<page>__<state>.png
doc, root = sys.argv[1], sys.argv[2]
n = 1
def call(name, args):
    global n; n += 1
    r = post({"jsonrpc":"2.0","id":n,"method":"tools/call","params":{"name":name,"arguments":args}})
    return r["result"]["content"][0]["text"]
call("activate_document", {"document_id": doc})
for page in sys.argv[3:]:
    call("activate_document", {"document_id": doc, "page_id": page})
    tree = json.loads(call("get_page_tree", {"document_id": doc, "page_id": page, "depth": 1}))
    for ch in tree["children"]:
        if ch["name"].count("/") != 2: continue
        fn = ch["name"].replace("/", "__") + ".png"
        s = call("export_image", {"document_id": doc, "page_id": page, "ids": [ch["id"]], "maxEdge": 4096, "path": root + "\\" + fn})
        print(fn, "ok" if "written" in s else "FAIL " + s[:120])
