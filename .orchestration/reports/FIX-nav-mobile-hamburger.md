# FIX-nav-mobile-hamburger (impl-web, 2026-11-01)

Branch `agent/senior/FIX-nav-mobile-hamburger` @ 30b487f5, stacked on `agent/senior/FIX-nav-lg-width` @ 5afedecf (worktree `wt/impl-web-shell`). Commit-local only.

Owner instruction (chat, 2026-11-01): "no contexto web mobile, foque em deixar o titulo do site e o search input na navbar apenas, com o menu hamburguer na extrema direita com os outros botões". It supersedes round 2 item 7 (and the icon-only mobile search of item 16) on web mobile.

## What changed

- `resources/js/components/Nav.tsx`, below `lg` (the breakpoint the nav already used):
  - The bar shows only the title, a search field (`flex-1`, opens the Ctrl+K palette as before; `/search` link fallback kept) and a hamburger at the far right.
  - The language, theme, notifications, account and login cluster is wrapped in `hidden lg:flex`.
  - The second-row category nav is removed.
- Hamburger:
  - It is a 44 px button with `aria-haspopup="dialog"` and `aria-expanded`.
  - Its name is "Menu", or "Menu, N unread notification(s)" (plural-aware). A numeric badge on the button shows the unread count.
  - It opens a Base UI Dialog as a right-side sheet, the same primitive as the palette, which gives the focus trap, Esc to close and focus return through `finalFocus`.
  - The sheet slides in through `data-starting-style` / `data-ending-style`, with `motion-reduce:transition-none`.
  - Following a link closes the sheet. The theme toggle keeps it open so the change is visible.
- Sheet contents:
  - avatar and name when signed in;
  - the four categories in their accent colours;
  - Feed, when the flag is on and the user is signed in;
  - Notifications, with the count;
  - the account items (shared with the desktop dropdown through one fragment), or Log in plus the guest items (sign up, pricing, API docs);
  - the language list;
  - the theme toggle.
- `lg` and wider: the desktop layout from FIX-nav-lg-width is unchanged. The language list, account links, guest links and avatar are now shared fragments.
- Avatar name at wider widths: the bar is capped at 1152 px by `shell`, so with the inline Feed there is no more room than in FIX-nav-lg-width. The name stays hidden in that case; the 1440 shot confirms no overflow.
- i18n: `nav.menuUnread` (one/other) and `nav.closeMenu` added in en, pt-BR and es. `nav.menu` already existed.
- `docs/openpencil/OWNER-FEEDBACK-ROUND-2.md`: appended a dated note that item 7 (and item 16's icon-only mobile trigger) is superseded on web mobile.
- No inline styles. v1 tokens and utilities only (`bg-backdrop/50`, `border-fg`, `bg-field-bg`, and so on).

## Evidence

- Red: `bunx vitest run resources/js/components/Nav.test.tsx` gave 3 failed and 5 passed. The three new tests failed: the mobile bar, the guest sheet and the signed-in sheet with unread.
- Green: the same command gives 8 passed. The old `justify-between` assertion for the removed category row was dropped.
- `bun run lint`: clean (190 files). `bun run typecheck`: clean. `bun run test`: 36 files, 142 tests passed.
- QA stack `hexascore-qa-implweb` (:3140), rebuilt from this branch with FLAG_SOCIAL=true:
  - Covered 16 combinations: logged out and logged in (admin, 1 unread) × light and dark × widths 390, 768, 1024 and 1440.
  - No page or header overflow in any of them: `scrollWidth <= clientWidth` and no header element past the viewport.
  - At 390 and 768:
    - The hamburger is 44×44, and `aria-expanded` is true when the sheet is open.
    - Focus starts inside the sheet. Every link and button in the sheet is at least 44 px tall.
    - Esc closes the sheet and focus returns to the hamburger.
- Focus trap:
  - Tabbing through the sheet three times at a 100 ms pace never left the dialog. Tab past the theme toggle wraps to the close button.
  - With synthetic Tabs pressed without any delay, focus can pass the Base UI guard, because the guard moves focus in a requestAnimationFrame. The palette uses the same primitive.
  - One trace run, the first one after the rebuild, also leaked into page content. I did not find the cause, and three later runs were clean.
- Shots: `.orchestration/reports/shots/impl-web/nav-mobile-{out,in}-{390,768,1024,1440}-{light,dark}-closed.png`, plus `-open.png` at 390 and 768.

## Not done / notes

- CHANGELOG: not edited. The project convention leaves it to the specialist at review, and FIX-nav-lg-width added no entry either.
- The sheet's Feed link at `/feed` uses the plain menu style. The current page is marked only with `aria-current`, which matches the desktop link minus the bold.
- One label flake: during one QA pass (in/dark/768), the hamburger label read "Menu" while the other passes read "1 unread". It looks like a timing effect of the shared prop on that load. It was not reproduced and not investigated.

## Follow-up: the two loose ends (impl-web, 2026-11-01)

New commit on the same branch: `ff264770 fix(ui): keep tab focus inside the mobile menu sheet` (on top of 30b487f5). Commit-local only.

### 1. Focus escaped the sheet: root cause and fix

- **Cause.** Base UI 1.9.0's `FloatingFocusManager` keeps the trap with two focus guards (`tabindex=0` spans around the popup). When a guard gets focus, it moves focus back into the popup through `enqueueFocus`, which waits for a `requestAnimationFrame`. Initial focus works the same way. The page behind the sheet is only `aria-hidden`, not `inert`. So when the browser handles a Tab before the next frame, focus moves from the after-guard to `<body>`, then to "Skip to content", the logo and the search field. A busy main thread does that: input events are handled before frames, which fits "the first run after a rebuild" (a cold host). Keys pressed right after opening do it too. It has nothing to do with lazy content or with mounting order.
- **Repro before the fix.** `repro-trap.mjs` on :3140 at 390, Tab with no pause between key presses: `12:GUARD, 13:OUT:BODY, 14:OUT:A:Skip to cont, 15:OUT:A:HexaScore`, in both runs.
- **Fix** (`Nav.tsx`). While the sheet is open, a document `keydown` handler wraps Tab and Shift+Tab synchronously:
  - Tab on the last link or button goes to the first one, and Shift+Tab on the first goes to the last.
  - If focus is not in the sheet yet, the next Tab goes to the first item and Shift+Tab to the last.
  - So no Tab ever reaches a guard, and the trap no longer depends on frame timing. Base UI still handles initial focus (the close button), Esc and `finalFocus`.
  - The popup gets `ref={sheet}`.
  - `initialFocus` was not needed: when Base UI's deferred initial focus runs, focus is already on the same element or has already moved inside the sheet, and in both cases it leaves focus where it is.
- **Test.** `Nav.test.tsx`, "Tab and Shift+Tab right after opening never leave the sheet": user-event with `delay: null`, so no frame runs between keys, and 40 presses each way straight after the click.

### 2. The hamburger read "Menu": root cause

- **Cause: the QA harness, not the app.** The old `navshot.mjs` logged in again in every pass. The login limiter (`app/http/limits/limits.go`: `Login = 5 per minute`, keyed by identifier and IP) answers the 6th attempt with 429. In the old loop, in/dark/768 was the 6th logged-in pass. The script then went to `/games` as a guest, and its only check ("URL is not /login") let that through.
- **Repro.** `repro-login.mjs`, 7 logins in a row: attempts 1 to 5 got POST 303, `user:true` and the label "Menu, 1 unread notification". Attempts 6 and 7 got POST 429, `user:false` and the label "Menu".
- **The prop is eager in practice.** `unreadNotifications` is a lazy shared prop (`app/http/middleware/unread_notifications.go`), so it is resolved on every page render and is not deferred. The label is derived from it on every render.
- **Test.** "the hamburger name follows the unread count when it changes" rerenders with 0, then 2, and checks the name goes from `nav.menu` to `nav.menuUnread` and the badge shows 2. It passed before the fix too: it pins the behaviour and does not fix a bug.
- **Harness fix.** The new `qa-hamburger.mjs` logs in once, fails on any status other than 303, reuses `storageState`, and records `auth.user` from the page props in every pass.

### Evidence

- **Red** (before the `Nav.tsx` change), `bunx vitest run resources/js/components/Nav.test.tsx`: 1 failed, 9 passed. Outside the sheet: `1:SPAN` (guard), `2:BODY`, `3:A:HexaScore`, `17:SPAN`, `18:BODY`, ...
- **Green:** 10 passed.
- `bun run lint`: clean (190 files). `bun run typecheck`: clean. `bun run test`: 36 files, 144 tests passed.
- **QA stack.** `hexascore-qa-implweb` (:3140) was rebuilt from ff264770 with FLAG_SOCIAL=true. `qa-hamburger.mjs` ran logged in as admin, dark, three runs each at 390 and 768. The first run was cold, right after the rebuild. Every pass: `user:true`, `social:true`, `unread:1`, label "Menu, 1 unread notification", Feed link in the sheet. 40 Tab presses and 40 Shift+Tab presses with no pause gave 0 escapes. Esc returns focus to the hamburger. The whole set ran twice, with the same result.
- **Old repro after the fix.** `repro-trap.mjs` (no pause, 2 runs): no escapes.
- **Scripts** are in the session scratchpad: `repro-trap.mjs`, `repro-login.mjs`, `qa-hamburger.mjs`.

## Review (especialista, 2026-11-01)

verdict: CORRIGIDO_E_APROVADO
re-run (stack top `agent/especialista/IMPL-web-nav-stack`, main merged in): `bunx vitest run resources/js/components/Nav.test.tsx resources/js/components/social` -> 3 files, 16 passed; `bun run typecheck` -> exit 0; `bunx biome check` on Nav files -> clean; `go test ./tests/i18n` -> ok; `go test -tags integration ./tests -run 'TestSocialFlagShared|TestFeed'` -> ok
checked: SDD 14 §12.3 (flag shared through vukem `flags.Share`, one bool, client acts only for signed-in users), SDD 09 nav, authz (no server change beyond the share), secrets none, sql none, i18n parity (plurals one/other in en/pt-BR/es; mobile catalogues do not ship `social`), scope, tests
fixes: 1. merge of main into the stack: `routes/web.go` import conflict (`social` + `wellknown` both kept); `Nav.tsx` keeps the hamburger sheet and main's lazy `SearchPalette` (mounted on first open, PERF-web-lcp), the removed phone category row stays removed; `Nav.test.tsx` keeps both the stack's tests and main's lazy-palette test.
notes: hamburger focus trap selector (`a[href],button:not([disabled])`) covers every control in the sheet: Inertia `Link` renders `a[href]` (logout renders a button), `BaseDialog.Close` and the theme toggle are buttons; there are no inputs or selects in the sheet.
PRONTO_PARA_MERGE agent/especialista/IMPL-web-nav-stack
