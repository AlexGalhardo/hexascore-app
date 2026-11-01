# Generate frames-dark/*.jsx (page "12 Web dark") from the light frames/web-*.jsx:
# platform "web-desktop" -> "web-desktop-dark", page bg -> dark bg, grid positions.
# Render the outputs with the dark prelude (v1-all-dark.js).
import glob
import os
import re
import sys

page = sys.argv[1]  # id of "12 Web dark"
here = os.path.dirname(os.path.abspath(__file__))
os.makedirs(os.path.join(here, "frames-dark"), exist_ok=True)
files = sorted(glob.glob(os.path.join(here, "frames", "web-*.jsx")))
desk = [f for f in files if "web-desktop_" in f]
mob = [f for f in files if "web-mobile_" in f]
for group, (cols, dx, dy, y0) in ((desk, (8, 1600, 4200, 0)), (mob, (16, 500, 6000, 30000))):
    for i, f in enumerate(group):
        body = open(f, encoding="utf-8").read().split("\n", 1)[1]
        body = re.sub(r'name="web-(desktop|mobile)/', r'name="web-\1-dark/', body, count=1)
        body = body.replace('bg="#FAF7F0"', 'bg="#14120F"', 1)
        x, y = (i % cols) * dx, y0 + (i // cols) * dy
        out = os.path.join(here, "frames-dark", os.path.basename(f).replace("web-desktop_", "web-desktop-dark_").replace("web-mobile_", "web-mobile-dark_"))
        open(out, "w", encoding="utf-8").write("//@ page=%s x=%d y=%d\n%s" % (page, x, y, body))
print(len(desk) + len(mob))
