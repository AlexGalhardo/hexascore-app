# Handoff: designer (DESIGN-round-1)

Date 2026-10-31. Worktree `C:\dev\hexascore-app\wt\designer-round-1`, branch `agent/designer/round-1`.

## Done

- v1 Editorial Press and v2 Neon Noir: documents, 19 PNG frames each, READMEs (`docs/openpencil/v1-editorial/`, `v2-noir/`).
- `docs/openpencil/ROUND-1.md` with v1 and v2 filled, v3 marked pending, owner checklist in pt-BR.
- Skill `.claude/skills/openpencil-design/SKILL.md` (adapted from openpencil-skill and figma-use, with every OpenPencil gap found) and rule `.claude/rules/design.md` (mirrored in `.grok/rules/`), owner request 2026-10-31.

## Next (v3 Soft Paper)

1. `open_file` `C:\Users\aleex\openpencil-hexascore\v3-paper\hexascore-v3-paper.fig` (MCP root is `C:\Users\aleex`; it is a copy of the v1 base). Delete the v1 frames on pages `00 Cover` … `11 Web mobile 390` and the `Shell components` frame on `99 Assets`; keep the `Covers` frame (`cover/<slug>` components; 6 real images, the rest average-colour gradients).
2. Build shell components (`header/desktop`, `header/desktop-auth`, `header/mobile`, `header/mobile-auth`, `footer/desktop`, `footer/mobile`) in the v3 palette, then the same 19 frames as v1/v2 with the same names. Palette and type: `docs/openpencil/v3-paper/README.md` (Nunito only, 16 px radius, pill chips, sticker score chips).
3. `activate_document`, export each frame with `path` into `C:\Users\aleex\openpencil-hexascore\v3-paper\frames\`, `save_file`, copy the `.fig` and PNGs into `docs/openpencil/v3-paper/`.
4. Fill the v3 column and paragraph in `ROUND-1.md`, commit, set `DESIGN-round-1` to `review` (already `review` for v1+v2; add a note) and ask the owner the checklist.

Read `.claude/skills/openpencil-design/SKILL.md` first: it lists the render pitfalls (spacer height, gradient transform, glyphs, replace order, export activation).

## Owner directive (2026-10-31, after the round-1 handoff)

"Create ALL screens, one by one, pixel perfect, web and mobile, nothing missing for review." After v3, the scope is every page of `docs/design-system.md` section 6 and every screen in `apps/mobile/app/` (tabs `index`, `[type]`, `search`, `lists`, `profile`; `[type]/[slug]`, `entities/[id]`, `lists/[id]`, `(auth)/login`, `update-required`), at 1440 and 390 plus Android 390x844, with their empty, loading, error and logged-in states, and `12 Web dark`. Plan it as one board row per page group and per version (or for the chosen version only, once the owner picks one) so that no single session runs out of context. Follow the pixel-perfect rules in the brief: 4 px grid, auto-layout, instances only, tokens for every value.

## Open points

- Cover images: `set_image_fill` needs base64 in the tool call; a script calling the local MCP HTTP server was blocked by the permission classifier, and hand-typed base64 corrupts. The owner can allow a scripted fill (token read from `mcp.json`, never printed) or accept average-colour placeholders.
- Production `/search` returns HTTP 500 (`q=diablo`, `q=witcher`, 2026-10-31): needs a `FIX-` row.
- Score scale: brief and legacy notes say 0 to 100; production and `docs/design-system.md` use out of 10. Frames follow production.
