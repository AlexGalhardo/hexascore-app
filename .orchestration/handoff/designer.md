# Handoff: designer (DESIGN-v1-round-2, claimed, not started)

Date 2026-10-31. Worktree `C:\dev\hexascore-app\wt\designer-round-1`, branch `agent/designer/v1-round-2` (stacked on `agent/designer/v1-editorial-web-catalog` @ `473fa4e`). Only this designer edits the worktree.

## State

- DESIGN-v1-editorial-web-catalog: delivered `473fa4e`, row `review`, `REV-DESIGN-v1-editorial-web-catalog` added (specialist). Report `.orchestration/reports/DESIGN-v1-editorial-web-catalog.md`.
- Owner chose v1 Editorial; v2 and v3 are frozen (no more work, including the v2 "none" contrast fix; its computed values are recorded in the report above).
- DESIGN-v1-round-2 (prio 0) is `doing` on the board. Scope: `docs/openpencil/OWNER-FEEDBACK-ROUND-2.md` (8 points, on main at `7970771`).

## Document and pipeline

- `C:\Users\aleex\openpencil-hexascore\v1-editorial\hexascore-v1-editorial.fig`, open as `tab-2`. Pages: `99 Assets` 0:3 (covers `cover/<slug>`, shell components `header/desktop` 0:35, `header/mobile` 0:54, `footer/desktop` 0:76, `footer/mobile` 0:88, `header/desktop-auth` 0:101, `header/mobile-auth` 0:124), `00 Cover` 0:4, `01 Foundations` 0:5, `02 Components` 0:6, `10 Web desktop 1440` 0:7, `11 Web mobile 390` 0:8, `20 Android 390x844` 0:9.
- Renderer: `scripts/openpencil/op.py` (copy in `C:\Users\aleex\openpencil-hexascore\tools\op.py`).
  - `python -I op.py render tab-2 v1-all.js <frame>.jsx…`: each file starts with `//@ page=0:7 x=… y=…`; `/*H*/` is replaced by the prelude; an existing top-level frame with the same name is replaced in place.
  - `op.py export tab-2 <outdir> <page> <name-substrings…>` writes `<platform>__<page>__<state>.png` at full size.
  - `op.py py tab-2 <script.py>` runs a script with `call(tool, args)` (do not name a variable `n`).
  - Sources: `tools/v1-prelude.js` (tokens `C`, `T`/`D`/`K` text, `B` badge, `Bar`, `Btn`, `Field`, `Chip`, `ItemRow`, shell instances), `tools/v1-item.js` (item page `ItemD`/`ItemM`, data `WOW`, `WITCHER`, `ILIAD`, `DIABLO`), `tools/v1-pages.js` (`PD`/`PM` shells, all other pages). `cat v1-prelude.js v1-item.js v1-pages.js > v1-all.js` before rendering. Frame files in `v1-editorial/src/`. All mirrored in the worktree under `docs/openpencil/v1-editorial/src/`.
  - The HTTP endpoint can answer 503 after many sessions; wait a few minutes (op.py now closes its session).
- Round-1 frames (`home`, `catalog-index`, `item/default`, `search`, `library`, `login`, `pricing`, Android) were drawn without the prelude; they have no source files.

## Plan for round 2

1. Tokens in the prelude: add a theme switch (`C` light and `CD` dark from the v1 README palette, plus a purple `primary`; check AA for every pair with a script before use). Scores in Inter with tabular figures (`B`, `Bar`, ScoreHead); Playfair only for headings. `Bar`: label, track and value in three fixed columns.
2. Shell components on `99 Assets` (propagate to every instance): mobile category row `space-between` full width (point 7), theme toggle moon in light and add a dark header variant with sun (point 8), `footer/mobile` links stacked and centred (point 4), dark variants of all six shell components.
3. Rebuild the round-1 frames from source in the prelude (home, catalog-index, item/default, search, library, login, pricing; desktop and mobile) so every frame can be rendered in light and dark from one file; then re-render all the catalogue frames.
4. New: item score block from v3 ("High score · from N ratings · six categories, one formula"), genre/platform chips under the meta line (not stacked on the title), "More games like this" cover strip; signed-in rating as 10 yellow half-stars per category with the number beside it (radio group, 44 px rows); navbar search dropdown with 8 rows (2:3 cover, title + year, type, score or "–") plus empty, loading and "see all results" states, desktop and mobile.
5. Dark: a `12 Web dark` page (create it) with a dark copy of every v1 frame, AA checked.
6. Export, save, copy the `.fig`, write `docs/openpencil/ROUND-2.md` (before/after list), report `.orchestration/reports/DESIGN-v1-round-2.md`, row to `review`, add `REV-DESIGN-v1-round-2`, commit.

## Gaps to add to `.claude/skills/openpencil-design/SKILL.md` (not yet written there)

- `render` drops children nested more than two array levels deep; flatten (`kids.flat(4)`).
- A wrapper `<Frame w={…}>` without `h` defaults to 100 px tall.
- Frames whose children are hidden keep a fixed height; set `sizing_vertical: HUG`.
- Text that shows stale (one line, clipped) after render or reopen: `batch_update` `{"sizing_horizontal":"FILL","auto_resize":"HEIGHT"}`.

## Open points

- "Did you mean" has no catalogue key; Recent changes page is not built (SDD 09 §6.2).
- Score scale: frames follow production (out of 10).
