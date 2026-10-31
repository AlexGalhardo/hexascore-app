# Handoff: designer (DESIGN-round-1, v3 in progress)

Date 2026-10-31. Worktree `C:\dev\hexascore-app\wt\designer-round-1`, branch `agent/designer/round-1`. Only this designer edits the worktree.

## Done

- v1 Editorial Press and v2 Neon Noir: documents, 19 PNG frames each, READMEs, real covers (`260db51`).
- v3 Soft Paper (`docs/openpencil/v3-paper/`), built on the refreshed base with the 23 real covers:
  - `99 Assets`: `Covers` frame (`cover/<slug>`) and `Shell components` with `header/desktop`, `header/desktop-auth`, `header/mobile`, `header/mobile-auth`, `footer/desktop`, `footer/mobile` (v3 palette, Nunito, 44 px pills).
  - `00 Cover`, `01 Foundations`, `02 Components` rendered, checked and exported: `frames/web__cover__default.png`, `web__foundations__default.png`, `web__components__default.png`.
  - `hexascore-v3-paper.fig` saved in `C:\Users\aleex\openpencil-hexascore\v3-paper\` and copied into the worktree.
  - README palette corrected for AA (below).

## Next (v3, 16 frames left)

Pages `10 Web desktop 1440` (`0:7`) and `11 Web mobile 390` (`0:8`) are empty; `20 Android 390x844` (`0:9`) is empty. Build, with the same names as v1/v2:

1. `web-desktop/<page>/default` at 1440 and `web-mobile/<page>/default` at 390 for `home`, `catalog-index` (`/games`), `item` (`/games/diablo-v`), `search`, `library` (use `header/*-auth`), `login`, `pricing`.
2. `android/home/default` and `android/item/default` at 390x844.
3. Content: copy the v2 frames' content (open `C:\Users\aleex\openpencil-hexascore\v2-noir\hexascore-v2-noir.fig`, read the PNGs in `docs/openpencil/v2-noir/frames/`): hero counts 24 873 games · 34 968 movies · 14 911 series; Diablo V 2029 Blizzard Entertainment, Action RPG / Windows / PlayStation 5 / Xbox Series X|S, 8.4 high · 213 ratings, categories Gameplay 8.8, Visuals 9.1, Sound 8.2, Storytelling 7.4, Technical 6.9, Fun 8.6; pricing Free $0 lifetime (100 requests in total, up to 10 items per request, 1 request per second), Pro $6/month (50 000 per month, 100 items, 10 rps), Business $29/month (500 000 per month, 100 items, 50 rps), USD/BRL/EUR switch, "Data coverage" note.
4. Then: export the 16 PNGs, save, copy, fill the v3 column and paragraph in `docs/openpencil/ROUND-1.md`, add the contrast table (below; also run `analyze_colors` on each version's `01 Foundations` page as asked and note its output), update `.orchestration/reports/DESIGN-round-1.md` to done, commit.
5. After v3: claim `DESIGN-v1-editorial-web-catalog` and continue the `DESIGN-<version>-<group>` rows in board order (orchestrator, Q-D1): one branch per row stacked on `agent/designer/round-1`, report, row to `review`, REV row.

## v3 JSX conventions (reuse in every render call)

```jsx
const T = (t, o = {}) => o.fill ? <Text font="Nunito" w="fill" size={o.s || 17} weight={o.w || 400} color={o.c || "#2B2A3A"}>{t}</Text> : <Text font="Nunito" size={o.s || 17} weight={o.w || 400} color={o.c || "#2B2A3A"}>{t}</Text>;
const S = (v, z = 1) => { const b = v == null ? ["#ECEAF2", "#5C5970"] : v >= 7.5 ? ["#CDEFE0", "#1D6B45"] : v >= 5 ? ["#FFE7A8", "#7A5200"] : ["#FFD6D6", "#A3262A"]; return <Frame flex="row" h={[28, 36, 64][z]} px={[10, 12, 20][z]} items="center" rounded={[14, 18, 32][z]} bg={b[0]} stroke="#FFFFFF" strokeWidth={2}>{T(v == null ? "–" : v.toFixed(1), { s: [14, 17, 32][z], w: 800, c: b[1] })}</Frame>; };
```

Page bg `#FFF9F2`, cards `#FFFFFF` + stroke `#EADFD0` + `rounded={20..24}`, primary button `#6B4CE6` pill 48 px with white 800 text, section tints lilac/peach/sky/mint with accent text `#5B3FC4` / `#9A3B12` / `#1F5F9E` / `#1D6B45`, muted `#6E6B80`, gutters 144 (1440) and 16 (390). Shell via `<Instance of="header/desktop" />`, covers via `<Instance of="cover/diablo-v" w={..} h={..} />`.

## Contrast (computed, WCAG 2.2, 2026-10-31)

v3 light: fg/bg 13.4, muted/bg 4.93, muted/surface 5.15, white on primary `#6B4CE6` 5.52 (old `#7A5CFA` 4.43, failed), primary text on bg 5.28, danger `#C2282D`/bg 5.52, accent on tint: game 5.02, movie `#9A3B12` 5.30 (old `#B5481A` 4.08, failed), series 5.22, book 5.25; chips high 5.25, mid 5.68, low 5.52, none `#5C5970` 5.65 (old `#6E6B80` 4.32, failed); `line` `#EADFD0` on bg 1.26 (decorative only; fields use `#D9CDBC` plus label, toggles off `#8C8499`).
v3 dark: fg 15.4, muted 8.2 (7.2 on surface), primary button text 6.8, accents 6.3 to 6.7, chips 5.9 to 6.8, danger 7.6.
v1 light: fg/bg 17.3, muted 6.9, white on primary 5.34, primary on bg 4.99, ink on score fills high 5.42 / mid 9.12 / low 5.43 / none 6.96; v1 dark fg 16.2, muted 8.3, primary 6.4. v2 not computed yet: run it from the v2 README palette.

## OpenPencil gaps found this session (add to the skill and the round report)

- Delete instance-bearing frames before their main components: deleting `Shell components` first left orphans and `save_file`, `get_page_tree` and `find_nodes` then failed with "Missing materialized component". Recovery: close without saving and redo the deletes in order.
- `open_file` on a path that is already open reloads into the same tab and merged stale nodes; copy the file to a new name to open a second copy.
- Text rendered while a Google font is still loading is stored with wrong glyphs: invisible in the first export, scrambled after save and reopen. Fix: render a test text and export it first; re-render scrambled frames, or reshape a text with `batch_update` `{"auto_resize":"WIDTH_AND_HEIGHT"}` (keeps fill width). `set_font` with `style:"ExtraBold"` falls back to Regular; the style name is `Extra Bold`.
- `get_jsx` output does not round-trip through `render` ("Unterminated JSX contents") for these frames.
- The local MCP HTTP endpoint answers 503 to new sessions after a few script runs in the same app session, so `scripts/openpencil/*.py` may be unavailable until OpenPencil restarts; use the MCP tools directly.

## Open points

- Production `/search` returns HTTP 500 (`q=diablo`, `q=witcher`, 2026-10-31): needs a `FIX-` row.
- Score scale: frames follow production (out of 10), not the 0 to 100 of the brief.
