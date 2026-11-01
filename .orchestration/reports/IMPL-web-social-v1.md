# IMPL-web-social-v1 v1 restyle of the social layer (FollowBar, follow requests, Feed, ActivityCard)

status: concluida
branch: agent/senior/IMPL-web-social-v1 @ cf307e2 (stacked on agent/senior/IMPL-web-feed-nav @ 3844a90, which sits on P7.A.1.3, which contains P7.A.1.2)
files: resources/js/components/social/{FollowBar,ActivityCard}.tsx, ActivityCard.test.tsx, resources/js/pages/Feed/Index.tsx, resources/js/pages/Account/Notifications.tsx (follow-request buttons and their wrapping only), resources/js/components/Nav.tsx (desktop nav `whitespace-nowrap`), resources/lang/{en,pt-BR,es}.json (`social.follow.followers`/`following` become one/other plurals; additive `social.feed.score`)
acceptance (v1 tokens and components only, no new frames):
- FollowBar: a hairline band (`border-y border-border`) under the profile header. Counts are semibold and tabular. Follow is the primary `Btn`; Unfollow and "Requested · cancel" use the secondary `Btn`, right-aligned.
- Inbox: Accept is the primary `Btn` and Decline the secondary one (44 px targets, ledge). On phones they wrap to their own line under the text instead of squeezing it.
- Feed/Index: Playfair display h1 (64 px desktop), a 2 px ink rule, and hairline-divided entries at `max-w-3xl`. Empty state is italic serif on the rule, as on Library. "Older activity" is a secondary button link.
- ActivityCard: an initial disc (same as the nav avatar) and a 15 px sentence. The nickname is semibold, the list is a primary link, the date is muted and tabular. The item ShelfCard (max-w-md) sits under the sentence.
- Fix: the rating card passed `stars` to ShelfCard, which prints "Your score 5.1" for someone else's rating. It now prints "Score 5.1" (`social.feed.score`) as ShelfCard children. Test asserts it.
- Fix: "1 followers" is now "1 follower" (one/other plurals, all three locales, key sets kept in sync).
- `bunx vitest run resources/js/components/social resources/js/components/Nav.test.tsx` -> 9/9; `bun run lint` -> 190 files, no fixes; `bun run typecheck` -> exit 0; `bun run test` -> 36 files / 138 tests; `go test ./tests/i18n` ok
- Browser: QA stack `hexascore-qa-implweb` on :3140, built from this branch. FLAG_SOCIAL=true was set locally only through a scratch compose override; docker-compose.yml is unchanged. Rows inserted by hand into `hexascore_qa_implweb`: feed_opt_in for users 2 and 3; admin follows them (accepted); user 4 has a pending follow to admin plus its `follow_request` notification; user 3 has the public list "Couch co-op nights"; 4 `activity_events` (one per kind). Shots: `.orchestration/reports/shots/impl-web/social-{feed,inbox,profile}-{1440,390}-{light,dark}.png`.
validation: lint ok, typecheck ok, vitest ok, browser ok (light/dark x 1440/390)
changelog: Changed: the follow bar, follow requests, feed and activity cards use the editorial design.
roadmap: no box (P7.A.1.2/P7.A.1.3 UI)
risks / follow-ups:
- Merge order: REV-P7.A.1.1, then 1.2, then 1.3, then feed-nav, then this branch.
- The language files are in P7.A.1.2/1.3's additive lock. The plural change rewrites two existing `social.follow.*` values; the mobile catalogues do not ship `social` (SDD 10).
- Out of scope, seen in the shots: User/Profile.tsx is still pre-v1 (header, tabs). Its ShelfCards also say "Your score" on another user's ratings. That file is locked by senior-a's P7.A.1.3b; it should be a later impl-web row.
- The feed date uses `toLocaleDateString(locale)` like the rest of senior-a's code (11/1/2026 in en); no relative time.

## Review (especialista, 2026-11-01)

verdict: CORRIGIDO_E_APROVADO
re-run (stack top `agent/especialista/IMPL-web-nav-stack`, main merged in): `bunx vitest run resources/js/components/Nav.test.tsx resources/js/components/social` -> 3 files, 16 passed; `bun run typecheck` -> exit 0; `bunx biome check` on Nav files -> clean; `go test ./tests/i18n` -> ok; `go test -tags integration ./tests -run 'TestSocialFlagShared|TestFeed'` -> ok
checked: SDD 14 §12.3 (flag shared through vukem `flags.Share`, one bool, client acts only for signed-in users), SDD 09 nav, authz (no server change beyond the share), secrets none, sql none, i18n parity (plurals one/other in en/pt-BR/es; mobile catalogues do not ship `social`), scope, tests
fixes: 1. merge of main into the stack: `routes/web.go` import conflict (`social` + `wellknown` both kept); `Nav.tsx` keeps the hamburger sheet and main's lazy `SearchPalette` (mounted on first open, PERF-web-lcp), the removed phone category row stays removed; `Nav.test.tsx` keeps both the stack's tests and main's lazy-palette test.
notes: hamburger focus trap selector (`a[href],button:not([disabled])`) covers every control in the sheet: Inertia `Link` renders `a[href]` (logout renders a button), `BaseDialog.Close` and the theme toggle are buttons; there are no inputs or selects in the sheet.
PRONTO_PARA_MERGE agent/especialista/IMPL-web-nav-stack
