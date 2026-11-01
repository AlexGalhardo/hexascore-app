# IMPL-web-v1-shell v1 editorial shell: fonts, Btn, fixed navbar, footer, container
status: concluida
branch: agent/senior/IMPL-web-v1-shell @ e9f3dce (stacked on agent/senior/IMPL-mobile-v1-tokens @ 33ece5f, merged in)
files: resources/css/app.css, resources/fonts/{inter-latin-wght,playfair-display-latin-700}.woff2, resources/js/components/{Nav,Footer,Nav.test}.tsx, resources/js/components/ui/{Btn,Dropdown}.tsx, resources/js/layouts/AppLayout.tsx, resources/lang/{en,pt-BR,es}.json, qa/e2e/specs/i18n.spec.ts
acceptance:
- `bun run lint` → biome ci, 173 files, no errors (exit 0)
- `bun run typecheck` → tsc --noEmit exit 0
- `bun run test` → vitest exit 0 (new Nav.test.tsx: Btn states, language dropdown Esc + focus return, trigger/toggle/sections/footer)
validation: lint ok, typecheck ok, tests ok, browser ok on QA stack hexascore-qa-implweb :3140 (/games, desktop 1440 + mobile 390, light + dark, language menu open, scrolled); shots in reports/shots/impl-web/ (desktop-light-_games-lang.png matches web__shell-components__default.png language-open; mobile-dark-_games-scroll.png)
changelog: Changed: v1 editorial shell (Inter + Playfair Display self-hosted, Btn with ledge/pressed states, sticky navbar with language dropdown, moon/sun toggle and Ctrl K trigger, footer without language links, one 1152 px container)
roadmap: no box
risks / follow-ups:
- Fonts: two OFL WOFF2 files (latin subset, 48 KB + 23 KB) downloaded from Google Fonts, served through Vite (hashed). No metric-override fallback yet: check CLS on Lighthouse.
- Derived tokens in app.css (no new hues): `hover` (surface mixed 4 % toward fg), `backdrop`, ledge shadows, nav shadow. Proposal to move `hover`/`backdrop` into packages/tokens.
- Navbar is `sticky` (same visible result as fixed, no layout offset). Language/account/more menus use a small Dropdown (disclosure, Esc + outside click), not Base UI Menu: Base UI positioner writes inline styles.
- `footer.rights` key is now unused (kept). Deletion banner still uses amber utilities (outside this row).
- e2e i18n.spec.ts updated for the language button (not run: e2e out of gate).

## Review (especialista, 2026-10-31)

verdict: CORRIGIDO_E_APROVADO
fixes: 1. the self-hosted Inter and Playfair Display ship without their OFL licence texts; copied them into resources/fonts (ce8cf26)\nre-run: batch branch agent/especialista/IMPL-web-batch (shell, search-palette, catalog, item, home merged on main 54105f9; lang conflicts were sibling additions, union-resolved and JSON-validated): `bun run lint` ok, `typecheck` ok, `test` 33 files / 129 ok; mobile catalogues recopied (catalog.filters is now plural, unused on mobile), mobile test 47 / 292 ok; `go build ./...` ok; `go test ./tests/i18n` ok\nchecked: i18n parity (3 locales, no hardcoded copy), no dangerouslySetInnerHTML or inline scripts, suggest fetch encodes params and drops stale answers, localStorage guarded, trailers on youtube-nocookie behind consent, scope, tests\nnot re-checked in a browser by the reviewer (authors' frames in the reports)\ngate on main after merge: bun lint/typecheck/test ok (full local.sh at the end of the batch)\nPRONTO_PARA_MERGE agent/senior/IMPL-web-v1-shell
