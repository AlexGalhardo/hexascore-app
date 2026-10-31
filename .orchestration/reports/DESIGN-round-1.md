# DESIGN-round-1

Worker: designer. Branch `agent/designer/round-1`. Status: partial (v1 and v2 done, v3 handed off).

## Delivered

- `docs/openpencil/v1-editorial/`: `hexascore-v1-editorial.fig`, 19 PNGs in `frames/`, `README.md`.
- `docs/openpencil/v2-noir/`: `hexascore-v2-noir.fig`, 19 PNGs in `frames/`, `README.md`.
- `docs/openpencil/v3-paper/README.md` (palette and type only; frames pending).
- `docs/openpencil/ROUND-1.md`: side-by-side table, one paragraph per version, pt-BR owner checklist.
- `.claude/skills/openpencil-design/SKILL.md`, `.claude/rules/design.md`, `.grok/rules/design.md`, brief line in `.orchestration/briefs/designer.md` (owner request: adapt openpencil-skill, frontend-design, impeccable and figma-use to this project).

Frames per version: cover, foundations, components; home, catalog-index, item, search, library, login, pricing at 1440 and 390; Android home and item at 390x844.

## OpenPencil tools used

`list_documents`, `new_document`, `open_file`, `save_file`, `create_page`, `rename_node`, `list_pages`, `get_page_tree`, `describe`, `query_nodes`, `render` (design JSX with `Component`/`Instance`, inline SVG, `linearGradient`), `node_replace_with`, `delete_node`, `set_fill`, `set_text`, `set_image_fill` (via a subagent), `list_available_fonts`, `get_font_status`, `activate_document`, `export_image`.

## Gaps found

- Save and export paths must be inside the MCP root `C:\Users\aleex`; files are saved there and copied into the worktree.
- `export_image` fails with "No visible nodes to export" unless the target document is active, and again after the app loses focus.
- `set_image_fill` only takes base64 in the call: 6 of 23 covers applied; the rest are average-colour gradients. Scripted fills over the local MCP HTTP endpoint were blocked by the permission classifier.
- `linearGradient` with `transform` breaks `export_image`; 8-digit hex, `fillOpacity` and `rgba` are ignored in SVG paths; Unicode glyphs (◐ ▾ ★) render as missing; `<Frame grow={1} />` defaults to 100 px tall; `node_replace_with` appends instead of replacing in place; large renders hit a 20 s RPC timeout (node still created).
- Playfair Display renders old-style figures (uneven score digits).

## Checks

- `bunx markdownlint-cli2@0.24.0` on the new Markdown: clean (one pre-existing error in `vukem-feedback.md`, untouched).
- Every frame was rendered and spot-checked with `export_image` previews (home, catalog, item, library for v1; cover, components, home, catalog, item for v2).

## Not verified

- WCAG contrast of every pair was not measured with a tool (`analyze_colors` not run); palette pairs come from the concept doc and the current tokens.
- Production `/search` returns HTTP 500, so search counts in the frames are plausible values, not real ones.
- Production has no ratings: all score numbers are illustrative.
