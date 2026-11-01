# FIX-nav-lg-width navbar overflow with the Feed link

status: concluida
branch: agent/senior/FIX-nav-lg-width @ 5afedecf (stacked on agent/senior/IMPL-web-social-v1 @ cf307e2, because the Feed link only exists there; worktree `wt/impl-web-shell`)
files: resources/js/components/Nav.tsx, resources/js/components/Nav.test.tsx
acceptance:

- Before, on QA :3140 (social-v1, signed in, FLAG_SOCIAL on), at 1024 px: the page scrolled sideways (scrollWidth 1038 in en/pt-BR and 1064 in es), and the bell and the account button were cut off. Shots: `reports/shots/impl-web/nav-lg-before-*.png`. At 1280 and 1440 px, the right cluster ran 53 px past the 1152 px bar.
- Fix: below `xl`, the inline Feed link is hidden (`hidden … xl:inline-flex`, with explicit classes because a bare `inline-flex` from `item` would win over `hidden`). The account-menu Feed item is now `xl:hidden`, where it used to be `lg:hidden`. While Feed is inline, the name next to the avatar is dropped; the menu still shows it. The categories stay visible at every width.
- Red first: `bunx vitest run resources/js/components/Nav.test.tsx` → 1 failed | 4 passed, both for the breakpoint and for the name assertion. After the fix → 5 passed.
- After, at 1024, 1280 and 1440 × en/pt-BR/es × light/dark: no page overflow and 0 px overrun past the bar. Feed is inline from 1280 and in the account menu at 1024. Shots: `nav-lg-after-{en,pt-br,es}-{1024,1280,1440}-{light,dark}.png` and `nav-lg-after-menu-es-1024-{light,dark}.png`.

validation: `bun run lint` ok, `bun run typecheck` ok, `bun run test` 36 files and 139 tests pass.
changelog: Fixed: the desktop navbar no longer overflows at 1024–1440 px when the Feed link is on.
roadmap: no box
risks / follow-ups: at `xl`, signed-in users with the social flag see the avatar without their name. Without the flag, the name still shows. If design wants the name back, the cost is the 280 px search width or a smaller nav gap. That is the owner's or the designer's call.

## Review (especialista, 2026-11-01)

verdict: CORRIGIDO_E_APROVADO
re-run (stack top `agent/especialista/IMPL-web-nav-stack`, main merged in): `bunx vitest run resources/js/components/Nav.test.tsx resources/js/components/social` -> 3 files, 16 passed; `bun run typecheck` -> exit 0; `bunx biome check` on Nav files -> clean; `go test ./tests/i18n` -> ok; `go test -tags integration ./tests -run 'TestSocialFlagShared|TestFeed'` -> ok
checked: SDD 14 §12.3 (flag shared through vukem `flags.Share`, one bool, client acts only for signed-in users), SDD 09 nav, authz (no server change beyond the share), secrets none, sql none, i18n parity (plurals one/other in en/pt-BR/es; mobile catalogues do not ship `social`), scope, tests
fixes: 1. merge of main into the stack: `routes/web.go` import conflict (`social` + `wellknown` both kept); `Nav.tsx` keeps the hamburger sheet and main's lazy `SearchPalette` (mounted on first open, PERF-web-lcp), the removed phone category row stays removed; `Nav.test.tsx` keeps both the stack's tests and main's lazy-palette test.
notes: hamburger focus trap selector (`a[href],button:not([disabled])`) covers every control in the sheet: Inertia `Link` renders `a[href]` (logout renders a button), `BaseDialog.Close` and the theme toggle are buttons; there are no inputs or selects in the sheet.
PRONTO_PARA_MERGE agent/especialista/IMPL-web-nav-stack
