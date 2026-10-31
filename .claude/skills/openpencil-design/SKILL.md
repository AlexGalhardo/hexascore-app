---
name: openpencil-design
description: Use when designing HexaScore screens in OpenPencil through its MCP tools (render design JSX, components and instances, variables, export) — the designer role's three-version flow, frames under docs/openpencil/, round exports and the later pixel-perfect pass. Adapted for this project from ZSeven-W/openpencil-skill (MIT) and the tool-agnostic parts of Figma's figma-use skill.
---

# OpenPencil design for HexaScore

Sources, adapted to the tools this project actually has (OpenPencil MCP 0.16.0, verified 2026-10-31):

- [ZSeven-W/openpencil-skill](https://github.com/ZSeven-W/openpencil-skill) (`openpencil-design`, MIT): layout and content rules. Its `op` CLI, batch DSL and tool names (`insert_node`, `batch_design`, `design_refine`, `export_frames`, …) do not exist in our MCP server; do not call them.
- [figma/mcp-server-guide `figma-use`](https://github.com/figma/mcp-server-guide/blob/main/skills/figma-use/SKILL.md): only the canvas-agnostic workflow rules (inspect first, skeleton then fill, components over copies, verify by image). The Plugin API parts (`use_figma`, `figma.*`, async page loading) do not apply.
- Design direction comes from `frontend-design` and `impeccable` (both vendored in `.claude/skills/`); audits from `web-design-guidelines`. Project constraints from `.claude/rules/design.md`.

## Before drawing

1. Read the brief `.orchestration/briefs/designer.md`, `docs/design-system.md` (tokens, page catalogue section 6, gaps), `docs/prompts-concept-site-pages.md` (page ids and the three art directions) and the legacy principles in `../legacy/docs/design-system.md`.
2. `list_documents`, then `describe` / `get_page_tree` (with `depth` or `root_id`) on what exists. Reuse existing components and match their naming before creating new ones.
3. Fonts: check with `list_available_fonts` (family filter) and, after the first render, `get_font_status`. Only OFL fonts the web app can self-host.
4. Content: real titles, years, counts, covers and prices from production, read-only (the Inertia props are in the `<script data-page="app">` JSON of each public page). Never log in to production.

## Document structure (one `.fig` per version)

Pages `00 Cover`, `01 Foundations`, `02 Components`, `10 Web desktop 1440`, `11 Web mobile 390`, `12 Web dark`, `20 Android 390x844`, `30 Flows`, plus `99 Assets` for real cover components (`cover/<slug>`) and shell components (`header/desktop`, `header/mobile`, `footer/*`). Top-level frame names are `<platform>/<page-id>/<state>`; PNG exports are `<platform>__<page-id>__<state>.png`.

## Building with `render` (design JSX)

- One screen per `render` call; build lists with `.map()` inside `{(() => { … return [...] })()}` so the string stays short. Large screens can hit the 20 s RPC timeout: the node is usually still created, so check with `get_page_tree` before re-rendering.
- Auto-layout everywhere (`flex="row"|"col"`, `gap`, `p*`). Never set `x`/`y` on children of a layout frame. Siblings in a row share one width strategy; never put `w="fill"` children in a `hug` parent. Text that must wrap or truncate needs `w="fill"`.
- Shared parts are `Component`s on `99 Assets` and are placed with `<Instance of="header/desktop" />` (by name, also resized with `w`/`h`). No detached copies.
- To swap a node in place, use `render` with `replace_id` (or `parent_id` + `insert_index`). `node_replace_with` appends the new node at the end of the parent.
- Icons: inline `<svg viewBox="0 0 24 24" size={20}><path d=… stroke=… fill="none" /></svg>`. Do not rely on Unicode glyphs (◐, ▾, ★, ⯪) — they render as missing glyphs.
- A spacer `<Frame grow={1} />` without `h` defaults to 100 px tall and stretches its row; write `<Frame grow={1} h={1} />`. A wrapper frame without `w`/`h` defaults to 100x100 and crops its child.
- `linearGradient(…, { transform })` renders but breaks `export_image` ("Cannot read properties of null"); use `set_fill` with `gradient="top-bottom"` instead.
- Colours: 6-digit hex only. `fillOpacity`, 8-digit hex and `rgba()` in SVG paths are ignored; pre-blend translucent fills against the background.
- 4 px grid for every size and gap; 44 px minimum touch targets; cover boxes 2:3.

## Images

`set_image_fill` takes base64 in the tool call. Keep images small (downscale covers with `ffmpeg` to 96x144 JPEG) and set them once on `cover/<slug>` components, then instance them. Never retype base64 by hand; a mistyped payload still reports success and renders garbage. Use `scripts/openpencil/set_image_fills.py <doc-id> <map-file>` (lines `<nodeId> <path-to-.b64>`), which calls the local MCP HTTP server with the token from `%LOCALAPPDATA%\OpenPencil\mcp.json` without printing it (owner approved 2026-10-31). `scripts/openpencil/export_frames.py <doc-id> <out-dir> <pageId>…` exports every `<platform>/<page-id>/<state>` frame of those pages.

## Saving and exporting

- `save_file` and `export_image(path=…)` only write inside the MCP root (`C:\Users\aleex`). Save to `C:\Users\aleex\openpencil-hexascore\<version>\` and copy into `docs/openpencil/<version>/` of the worktree.
- `export_image` works only on the document that is active in the app ("No visible nodes to export" otherwise, also after the app loses focus): call `activate_document` with the frame's `page_id` before exporting from each page (export only sees the active page), then export with `document_id`, `page_id` and `ids`. Use `maxEdge` 900 to 1200 for a visual check and `scale` 1 (`maxEdge` 4096) for the delivered PNG.

## Verify

After each screen: one `export_image` at a small `maxEdge` and look for clipped or overflowing text, missing glyphs, wrong order after replacements, contrast, and placeholder copy. Then `describe` for naming issues and `get_font_status` for substitutions. Fix at the source and re-export only what changed. Record every tool gap in the round report.

## Known gaps (OpenPencil 0.16.0, 2026-10-31)

- Deletion order: delete frames that hold instances before their main components. Deleting `Shell components` first left orphans, and `save_file`, `get_page_tree` and `find_nodes` then failed with "Missing materialized component". Recovery: close without saving and redo the deletes in order.
- `open_file` on a path that is already open reloads into the same tab and merges stale nodes. To open a second copy, copy the file to a new name first.
- Font loading: text rendered while a Google font is still loading is stored with wrong glyphs (invisible in the first export, scrambled after save and reopen). Render a test text and export it before the real screens; re-render scrambled frames, or reshape a text with `batch_update` `{"auto_resize":"WIDTH_AND_HEIGHT"}` (keeps fill width).
- `set_font` with `style:"ExtraBold"` silently falls back to Regular; the style name is `Extra Bold` (with a space).
- `get_jsx` output does not round-trip through `render` ("Unterminated JSX contents"); rebuild from your own JSX instead.
- The local MCP HTTP endpoint answers 503 to new sessions after a few script runs in the same app session, so `scripts/openpencil/*.py` may fail until OpenPencil restarts; use the MCP tools directly.
- `render` takes one JSX expression: a `const` before the root never runs ("x is not defined"). Put helpers inside the root, `<Frame …>{(() => { const T = …; return [ … ]; })()}</Frame>`, and avoid spread arguments and destructured parameters (parse error).
- Clipping is `overflow="hidden"` on a frame; `clip` is ignored with a warning. Without it, a horizontal scroller widens the exported frame.
- `render` drops children nested more than two array levels deep; flatten lists (`kids.flat(4)`).
- A wrapper `<Frame w={…}>` without `h` defaults to 100 px tall; frames whose children are hidden keep a fixed height (set `sizing_vertical: HUG`).
- Text that shows stale (one line, clipped) after render or reopen: `batch_update` `{"sizing_horizontal":"FILL","auto_resize":"HEIGHT"}`.
- `fontFeatures` / `fontVariantNumeric` on `<Text>` are ignored, so tabular figures cannot be drawn: put numbers in fixed-width, right-aligned text boxes and specify `tabular-nums` in the handoff.
- `shadow="0 2px 0 #RRGGBB"` (CSS syntax) becomes a drop shadow effect; `shadows={[…]}` is ignored. `justify="between"` gives space-between.
- Render errors report a line and column that do not map to the source. Syntax-check the expanded JSX first: `docs/openpencil/v1-editorial/src/build.sh <frame.jsx…>` wraps it as a module and runs `bun build --no-bundle`.
- Long frames often return "RPC timeout (20s)" although the node is created (and a `replace_id` swap completes); check names with `get_page_tree` instead of re-rendering.
- Replacing a frame that holds component main nodes deletes them; delete the frames that use those components first, then re-render the sheet and call `node_to_component` again. To swap components without breaking old instances, rename the old mains (`old/…`) and render new ones under the original names.
