#!/usr/bin/env sh
# Concatenate the v1 sources into one prelude per theme for op.py render.
cd "$(dirname "$0")"
for t in 0 1; do
	out=$([ $t = 0 ] && echo v1-all.js || echo v1-all-dark.js)
	{ echo "const DK = $t;"; cat v1-prelude.js v1-shell.js v1-item.js v1-pages.js v1-pages-r1.js v1-new.js; } > "$out"
done
# Syntax check: the renderer's parse errors carry no usable line numbers.
for f in "$@"; do
	python -I -c "
import sys
p=open('v1-all.js',encoding='utf-8').read(); b=open(sys.argv[1],encoding='utf-8').read().split('\n',1)[1]
open('.chk.jsx','w',encoding='utf-8').write('export const x = '+b.replace('/*H*/',p))
" "$f" && bun build --no-bundle .chk.jsx --outdir .chkout >/dev/null 2>.chk.err || { echo "SYNTAX $f"; grep -m3 -E "error|\^" .chk.err | cut -c1-300; }
done
rm -rf .chk.jsx .chkout .chk.err
