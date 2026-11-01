# IMPL-web-feed-nav Feed link in the navbar behind the social flag

status: concluida
branch: agent/senior/IMPL-web-feed-nav @ 3844a90 (stacked on agent/senior/P7.A.1.3 @ 962e3459)
files: routes/web.go (`r.Use(ff.Share(social.Flag))` after ShareProps + import), resources/js/types.ts (`flags?: { social?: boolean }`), resources/js/components/Nav.tsx, resources/js/components/Nav.test.tsx, tests/feed_test.go (`TestSocialFlagShared`)
acceptance:
- The flag reaches the client through vukem's own allowlisted share (`flags.Share`, <https://vukem.dev/docs/digging-deeper/feature-flags>): every Inertia page gets `flags: { social: bool }`.
- Nav: a "Feed" link (existing key `social.feed.title`, no new strings) in the desktop sections nav after a hairline, and as the first item of the account menu below `lg` (the phone section bar keeps its four sections). It shows only when `auth.user` and `flags.social` are both set. `aria-current="page"` on /feed.
- `bunx vitest run resources/js/components/Nav.test.tsx` -> 4/4 (new: anonymous + flag on, signed in + flag off, and signed in + flag on give 0, 0, and 1 link to `/feed`)
- `go test -tags integration ./tests -run 'TestSocialFlagShared|TestFeedDisabledByDefault'` -> ok (flags.social false/true follows FLAG_SOCIAL)
- `bun run lint` -> 190 files, no fixes; `bun run typecheck` -> exit 0; `bun run test` -> 36 files / 138 tests passed; `go build ./...` and `go vet ./routes` ok
- Browser: see IMPL-web-social-v1 (same QA build). Shots `.orchestration/reports/shots/impl-web/social-*-1440-*.png` show the link. At 1440 the link pushed "TV Series" onto two lines, so the desktop sections nav is now `whitespace-nowrap` (in the social-v1 commit).
validation: lint ok, typecheck ok, vitest ok, go integration ok, browser ok
changelog: Added: a "Feed" link in the navbar for signed-in users while the social layer is on.
roadmap: no box (P7.A.1.3 UI glue)
risks / follow-ups:
- `routes/web.go` is in the PERF-web-lcp / PERF-i18n locks. The overlap is 6 added lines after the `NotBanned` middleware plus one import, so a trivial merge.
- `flags` is shared on every page, anonymous ones included (one bool). This is vukem's documented pattern; the client only acts on it for signed-in users.
- Not checked at 1024 px (lg). The search box shrinks to w-56 there, but five links plus the Feed link are tight.

## Review (especialista, 2026-11-01)

verdict: CORRIGIDO_E_APROVADO
re-run (stack top `agent/especialista/IMPL-web-nav-stack`, main merged in): `bunx vitest run resources/js/components/Nav.test.tsx resources/js/components/social` -> 3 files, 16 passed; `bun run typecheck` -> exit 0; `bunx biome check` on Nav files -> clean; `go test ./tests/i18n` -> ok; `go test -tags integration ./tests -run 'TestSocialFlagShared|TestFeed'` -> ok
checked: SDD 14 §12.3 (flag shared through vukem `flags.Share`, one bool, client acts only for signed-in users), SDD 09 nav, authz (no server change beyond the share), secrets none, sql none, i18n parity (plurals one/other in en/pt-BR/es; mobile catalogues do not ship `social`), scope, tests
fixes: 1. merge of main into the stack: `routes/web.go` import conflict (`social` + `wellknown` both kept); `Nav.tsx` keeps the hamburger sheet and main's lazy `SearchPalette` (mounted on first open, PERF-web-lcp), the removed phone category row stays removed; `Nav.test.tsx` keeps both the stack's tests and main's lazy-palette test.
notes: hamburger focus trap selector (`a[href],button:not([disabled])`) covers every control in the sheet: Inertia `Link` renders `a[href]` (logout renders a button), `BaseDialog.Close` and the theme toggle are buttons; there are no inputs or selects in the sheet.
PRONTO_PARA_MERGE agent/especialista/IMPL-web-nav-stack
