# DESIGN-v1-round-2

Worker: designer. Branch `agent/designer/v1-round-2` (stacked on `agent/designer/v1-editorial-web-catalog`). Status: review.

## Delivered

- Owner items 1 to 16 of `docs/openpencil/OWNER-FEEDBACK-ROUND-2.md` on v1 Editorial. Before/after list, dark contrast table and frame list: `docs/openpencil/ROUND-2.md`.
- `docs/openpencil/v1-editorial/hexascore-v1-editorial.fig`: new page `12 Web dark` (a dark copy of all 67 web frames), `99 Assets` with `Shell components (round 2)` and `Buttons (round 2)` in light and dark (components), 17 more real covers.
- 138 PNGs in `docs/openpencil/v1-editorial/frames/`: 67 light (`web-desktop__…`, `web-mobile__…`) and 67 dark (`web-desktop-dark__…`, `web-mobile-dark__…`), plus the shell and button sheets (`web__shell-components[-dark]__default.png`, `web__buttons[-dark]__default.png`). Round-1 web PNGs replaced; the `search/default` frame stays (results page), the item 5 dropdown frames were replaced by `search-palette/*` (item 16).
- Sources: every web frame now renders from `docs/openpencil/v1-editorial/src/` (both themes from one file per frame). `build.sh` builds the two preludes and syntax-checks with `bun build`; `gen_dark.py` writes the dark copies.
- Scripts: `scripts/openpencil/op.py` waits for the app after RPC timeouts and exports up to 8192 px; new `scripts/openpencil/check_container.py` (item 9 check).
- `docs/openpencil/v1-editorial/README.md`: round-2 palette (purple primary, danger, field, star, button ledge tokens), type rule (Playfair headings only, Inter tabular scores), Google mark exception.
- `.claude/skills/openpencil-design/SKILL.md`: eight new known gaps.

## Checks

- Container (item 9): `PAGES=0:7,663413479:19428 python -I scripts/openpencil/op.py py tab-2 scripts/openpencil/check_container.py` → 66 of 66 desktop frames OK (33 light, 33 dark). The one failure found on the way (`rating-widget/states`, right column 128 px past the edge) was fixed.
- Frame set: every source frame exists once on its page (script compare of names, no duplicates, none missing).
- Contrast: computed with the WCAG 2.2 formula for every token pair used (table in `ROUND-2.md`); all text pairs ≥ 4.5:1 and UI pairs ≥ 3:1 except the light books accent on surface (4.28), which is used on the page background only.
- Visual: full-size exports viewed for `item/logged-in`, `search-palette/results` and `empty`, `rating-widget/states` (desktop and mobile, light and dark), plus home, shell and button sheets.
- `bunx markdownlint-cli2@0.24.0` on the changed Markdown files: clean.

## OpenPencil tools used

`list_documents`, `get_page_tree`, `get_jsx`, `get_node`, `node_children`, `render` (via `op.py`), `node_to_component`, `rename_node`, `create_page`, `delete_node`, `node_move`, `set_image_fill` (via `set_image_fills.py`), `activate_document`, `export_image`, `save_file`.

## Not done / open

- Android frames (`20 Android 390x844`) are not part of this round.
- Trending (last 7 days) needs a rollup; scores and counts are illustrative (production has no ratings).
- Games accent equals the primary purple; Google button uses Inter and 44 px with the v1 ledge (confirm against Google's guidelines).
- Superseded shell components (`old/…`, `old2/…`) remain on `99 Assets` in frames named `Shell components` and `… (superseded)`; no frame instances them any more, delete them in the pixel-perfect pass (frames first, then mains).
- `export_image` caps `maxEdge` at 4096: mobile item frames taller than 4096 px (light and dark) are exported at about 0.93 scale.
- Tabular figures cannot be drawn in OpenPencil; frames fake them with fixed-width boxes, implementation uses `tabular-nums`.
