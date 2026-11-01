# DESIGN-round-1

Worker: designer. Branch `agent/designer/round-1`. Status: done (v1, v2 and v3 delivered, 19 frames each).

## Delivered

- `docs/openpencil/v1-editorial/`: `hexascore-v1-editorial.fig`, 19 PNGs in `frames/`, `README.md`.
- `docs/openpencil/v2-noir/`: `hexascore-v2-noir.fig`, 19 PNGs in `frames/`, `README.md`.
- `docs/openpencil/v3-paper/`: `hexascore-v3-paper.fig`, 19 PNGs in `frames/`, `README.md` (palette corrected for AA).
- `docs/openpencil/ROUND-1.md`: side-by-side table, one paragraph per version, WCAG contrast table for v1, v2 and v3, OpenPencil gaps, pt-BR owner checklist.
- `.claude/skills/openpencil-design/SKILL.md`, `.claude/rules/design.md`, `.grok/rules/design.md`, brief line in `.orchestration/briefs/designer.md` (owner request: adapt openpencil-skill, frontend-design, impeccable and figma-use to this project).

Frames per version: cover, foundations, components; home, catalog-index, item, search, library, login, pricing at 1440 and 390; Android home and item at 390x844.

## OpenPencil tools used

`list_documents`, `new_document`, `open_file`, `save_file`, `create_page`, `rename_node`, `list_pages`, `get_page_tree`, `describe`, `query_nodes`, `render` (design JSX with `Component`/`Instance`, inline SVG, `linearGradient`), `node_replace_with`, `delete_node`, `set_fill`, `set_text`, `set_image_fill` (via a subagent), `list_available_fonts`, `get_font_status`, `activate_document`, `export_image`.

## Gaps found

- Save and export paths must be inside the MCP root `C:\Users\aleex`; files are saved there and copied into the worktree.
- `export_image` fails with "No visible nodes to export" unless the target document is active, and again after the app loses focus.
- `set_image_fill` only takes base64 in the call. After the owner's yes, `scripts/openpencil/set_image_fills.py` applied all 23 real covers over the local MCP HTTP endpoint and `scripts/openpencil/export_frames.py` re-exported every frame (export only sees the active page).
- `linearGradient` with `transform` breaks `export_image`; 8-digit hex, `fillOpacity` and `rgba` are ignored in SVG paths; Unicode glyphs (◐ ▾ ★) render as missing; `<Frame grow={1} />` defaults to 100 px tall; `node_replace_with` appends instead of replacing in place; large renders hit a 20 s RPC timeout (node still created).
- Playfair Display renders old-style figures (uneven score digits).
- v3 session: deleting main components before their instances corrupts the document; `open_file` on an open path merges stale nodes; text rendered while a Google font loads is stored with wrong glyphs; `set_font` needs `Extra Bold`; `get_jsx` does not round-trip; the MCP HTTP endpoint answers 503 after a few script sessions; `render` runs one expression only (helpers inside an IIFE) and clips with `overflow="hidden"` (`clip` ignored). All are in the skill's "Known gaps".

## Checks

- `bunx markdownlint-cli2@0.24.0` on the new Markdown: clean (one pre-existing error in `vukem-feedback.md`, untouched).
- Every frame was rendered and spot-checked with `export_image` previews (home, catalog, item, library for v1; cover, components, home, catalog, item for v2).

## Not verified

- WCAG contrast was computed with the relative-luminance formula from each README palette (table in `ROUND-1.md`), not measured on rendered pixels; `analyze_colors` was not run. v2 fails on the "none" score colour (4.40 light, 4.09 dark): fix before tokens are cut.
- v3 frames were checked with `export_image` previews; the pricing subtitle on desktop is left-aligned in a centred block (cosmetic, round 2).
- Production `/search` returns HTTP 500, so search counts in the frames are plausible values, not real ones.
- Production has no ratings: all score numbers are illustrative.

## Review (especialista, 2026-10-31)

APROVADO at 0fb8e59. Scope: `docs/openpencil/` (three `.fig` files, 57 PNG frames, READMEs, `ROUND-1.md`), `.claude/skills/openpencil-design/`, `.claude/rules/design.md`, `.grok/rules/design.md`, designer brief and handoff, and two local helper scripts in `scripts/openpencil/` (read the MCP token from the user's OpenPencil config, never print it). No app code, migrations or lockfiles changed. `bunx markdownlint-cli2@0.24.0` on the new Markdown: clean (only the pre-existing `vukem-feedback.md` error on the branch base, already fixed on main). The production `/search` 500 is covered by FIX-search-timeout-500 and FIX-search-prod-slow, both on local main and waiting for a deploy.

PRONTO_PARA_MERGE agent/designer/round-1
