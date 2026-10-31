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
doc = sys.argv[1]
for line in open(sys.argv[2]):
    nid, path = line.split()
    data = open(path.replace('/c/', 'C:/', 1)).read().strip()
    n += 1
    r = post({"jsonrpc":"2.0","id":n,"method":"tools/call","params":{"name":"set_image_fill","arguments":{"document_id":doc,"id":nid,"image_data":data,"scale_mode":"FILL"}}})
    s = json.dumps(r)
    print(nid, "FAIL " + s[:200] if ('"isError": true' in s or '"error"' in s) else "ok")
