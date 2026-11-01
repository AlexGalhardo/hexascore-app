# DESIGN-v1-editorial-web-catalog

Worker: designer. Branch `agent/designer/v1-editorial-web-catalog` (stacked on `agent/designer/round-1`). Status: review.

## Delivered

- `docs/openpencil/v1-editorial/hexascore-v1-editorial.fig` (saved from `C:\Users\aleex\openpencil-hexascore\v1-editorial\`).
- 37 new PNGs in `docs/openpencil/v1-editorial/frames/` plus re-exported `web-desktop__home__default.png` and `web-mobile__home__default.png`:
  - `item`: `logged-out`, `logged-in`, `no-score`, `no-cover`, `no-synopsis` (desktop and mobile).
  - `home/recent`; `home/default` now hides the "Recently rated" block, as production does.
  - `edition/default`, `entity/default`, `entity/empty`, `search/empty-query`, `search/too-short`, `search/no-results`, `catalog-index/empty`, `recent-changes/default`, `docs/default`, `legal/default`, `error/404`, `error/500` (desktop and mobile); `catalog-index/filters-open` (mobile only, desktop filters are always visible).
- `docs/openpencil/v1-editorial/src/`: the design JSX that renders every new frame (`v1-prelude.js`, `v1-item.js`, `v1-pages.js`, `frames/*.jsx`).
- `scripts/openpencil/op.py`: renders JSX files over the local MCP HTTP endpoint (shared helper prelude, replaces a frame in place by name) and exports frames by name; closes its MCP session on exit.
- `docs/openpencil/v1-editorial/README.md`: frame list, sources, content notes.

## Fixes to round-1 frames

- Home desktop: the "Six categories, one score" panel text was clipped to one line; long titles ("Intergalactic: The Heretic Prophet", "Spider-Man: Beyond the Spider-Verse") overlapped the score chip. Titles now fill and truncate.
- The new item frames replace the round-1 defects (category bar values overlapping the bar, trailer copy clipped); `item/default` itself is unchanged.

## OpenPencil tools used

`list_documents`, `get_page_tree`, `get_jsx` (reading conventions only), `render` (via `op.py`), `clone_node`, `node_move`, `batch_update` (rename, visibility, sizing, text reshape), `set_text`, `set_text_properties`, `node_children`, `find_nodes`, `query_nodes`, `get_node`, `activate_document`, `export_image`, `save_file`, `delete_node`.

## Gaps found (added to the handoff; skill update in the next row)

- Nested arrays more than two levels deep inside a frame's children are silently dropped by `render` (legal sections and docs bullets vanished on mobile). Flatten child arrays (`kids.flat(4)`).
- A wrapper `<Frame w={100}>` without `h` still defaults to 100 px tall and inflates table rows.
- The MCP HTTP endpoint answered 503 again after about ten script sessions; it recovered after a few minutes. `op.py` now sends `DELETE` on its session at exit.
- Frames whose content was hidden keep their old fixed height; set `sizing_vertical: HUG`.

## Open points

- v2 "none" colour: computed AA values light `#5F6779` (5.24 on bg, 5.67 on surface, 4.88 on raised) and dark `#8A93A6` (6.30, 5.69, 5.15); 32 nodes on desktop, 14 on mobile and 8 on Android use the old `#6B7385`. Not applied: the owner chose v1 and froze v2 and v3 (orchestrator, 2026-10-31).
- "Did you mean" has no catalogue key; Recent changes page is not built (SDD 09 §6.2).
- Code blocks use Inter (two-family budget); a monospace would be a third family.

## Checks

- Every frame exported and visually checked at reduced size; no clipped text, overflow or missing glyphs left in the checked frames.
- `bunx markdownlint-cli2@0.24.0` on the touched Markdown: clean (only the pre-existing `vukem-feedback.md` MD033).
