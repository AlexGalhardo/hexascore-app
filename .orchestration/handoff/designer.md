# Handoff: designer (DESIGN-v1-editorial-web-catalog, just started)

Date 2026-10-31. Worktree `C:\dev\hexascore-app\wt\designer-round-1`, now on branch `agent/designer/v1-editorial-web-catalog` (stacked on `agent/designer/round-1`, head `0fb8e59`). Only this designer edits the worktree.

## Done

- DESIGN-round-1 complete: v1, v2 and v3 have 19 PNGs each; `ROUND-1.md` has the v3 column, the v3 paragraph, the WCAG contrast table for all three versions and the OpenPencil gaps; report `.orchestration/reports/DESIGN-round-1.md` is `done`; commit `0fb8e59` on `agent/designer/round-1`. Board row DESIGN-round-1 notes updated (status stays `review`, REV-DESIGN-round-1 is the specialist's).
- Board row `DESIGN-v1-editorial-web-catalog` set to `doing` (designer) in the main checkout's `.orchestration/queue/TASKS.md`.
- `.claude/skills/openpencil-design/SKILL.md` has a "Known gaps" section; read it before drawing.

## Next: DESIGN-v1-editorial-web-catalog

Document: `C:\Users\aleex\openpencil-hexascore\v1-editorial\hexascore-v1-editorial.fig` (open as `tab-2` in the current app session; pages `99 Assets` 0:3, `10 Web desktop 1440` 0:7, `11 Web mobile 390` 0:8). Round 1 already has `home`, `catalog-index`, `item`, `search`, `pricing`, `login`, `library` default frames at 1440 and 390.

Frames still to draw for this row (design-system.md 6.1 + public pages), desktop 1440 and mobile 390, `default` state unless named:

1. `catalog-show` states: `no-score` ("Not enough ratings" on badge and panel), `logged-out` ("Log in to rate"), `no-cover`, `no-synopsis`; plus the full item page blocks round 1 skipped (RatingPanel with six ScoreInputs, Synopsis, Details, Editions, Entities, Trailers).
2. `catalog-edition` (book edition: h1, subtitle, definition list, back link).
3. `entity-show` (kind label, name, facts, description, role chips, one section per type with "Showing x of y"), plus `empty`.
4. `search` states: `empty-query`, `too-short`, `no-results` with "did you mean".
5. `catalog-index` states: `empty` ("No items match..."), filters open on mobile.
6. `home` state: `recent` block present.
7. Recent changes, docs (API docs page), legal pages (terms, privacy), error pages (404, 500, 419/403 if the catalogue lists them). Check section 6.1 and the "public" pages in `docs/design-system.md` for exact ids and states.

Then export to `docs/openpencil/v1-editorial/frames/`, save and copy the `.fig`, write `.orchestration/reports/DESIGN-v1-editorial-web-catalog.md`, row to `review`, add `REV-DESIGN-v1-editorial-web-catalog`, commit on this branch.

v1 conventions: not recorded yet. Before drawing, export one v1 round-1 frame per page at `maxEdge` 900 and read its tokens from `docs/openpencil/v1-editorial/README.md` (Playfair Display + Inter, paper bg, square band fills with ink numbers, 2 px radius). Write the v1 JSX helpers here once settled.

## Render notes (verified this session)

- `render` takes one JSX expression: put helpers inside the root, `<Frame name=… w={1440} flex="col" bg=…>{(() => { const T = …; return [<Instance of="header/desktop" />, …, <Instance of="footer/desktop" />]; })()}</Frame>`. No spread args or destructured params.
- Clip horizontal scrollers with `overflow="hidden"` (on the root of mobile frames too).
- Activate the page before `export_image`; export with `path` under `C:\Users\aleex\openpencil-hexascore\<version>\frames\` at `maxEdge` 4096, then copy.
- Run a font test render plus export at the start of each app session (font loading bug).

## Open points

- Production `/search` returns HTTP 500 (`q=diablo`, `q=witcher`, 2026-10-31): needs a `FIX-` row.
- v2 "none" score colour `#6B7385` fails as text (4.40 light, 4.09 dark): fix before tokens are cut (ROUND-1.md contrast table).
- v3 desktop pricing subtitle is left-aligned inside a centred block; v3 library side panel nests tinted rows in a card (round 2 if v3 is chosen).
- Score scale: frames follow production (out of 10), not the 0 to 100 of the brief.
