# FIX-mobile-blocked-screens: banned and deletion-scheduled screens

Branch `agent/senior/FIX-mobile-blocked-screens` @ 73ac1a6. It is stacked on `agent/senior/P5.A.4.3` @ b790018 because it edits the same `app/_layout.tsx`.

Merge order: FIX-mobile-root-wiring, then P5.A.4.1, P5.A.4.2, P5.A.4.3, then this branch.

## Why

P5.A.3.2, which is ticked, built the session states `banned` and `deletionScheduled`, but no screen used them. The two gaps were:

- A banned session fell through to the login screen.
- A deletion-scheduled session got the full app.

SDD 10 §6.2 requires two other things:

- On 403 `account-banned`, wipe the session and show the ban reason and end date.
- When `deletionScheduledAt` is set, show a cancel screen that sends the user to web `/login` (SDD 03 §8).

P5.B.4.3 also names the deletion-scheduled screen. When P5.B.4.3 is done, it can tick that part as delivered here.

## What changed

- `src/providers/root.tsx`: `routeGuards` now returns `{app, auth, update, banned, deletion}`, and each session state opens exactly one group:
  - A pending update still wins over every other state.
  - `loading` opens none of them.
  - A deletion-scheduled session no longer gets `app`.
- `app/_layout.tsx` adds two `Stack.Protected` groups: `banned` and `deletion-scheduled`.
- `src/account/Blocked.tsx` holds two pure components:
  - `BannedBody` shows a title, the reason (only when there is one), and "Until {date}" or "No end date", with a "Back to login" button that calls `session.signOut()`.
  - `DeletionScheduledBody` shows the existing `account.banner.text` with the date, plus two buttons:
    - "Cancel deletion" opens web `/login`.
    - "Back to login" runs `logout` with the token revoke, like `update-required`.
  - Both format dates with `Intl.DateTimeFormat(locale, {dateStyle: "long", timeZone: "UTC"})`.
- `app/banned.tsx` and `app/deletion-scheduled.tsx` connect these components to `session` and `appClient`.
- i18n: new keys `account.blocked.{title,reason,until,permanent,back}` in en, pt-BR and es, then `catalogues` was re-run. The cancel label reuses the existing `account.banner.*` keys.

No new dependencies.

## Acceptance

Tests were written first. The red run, `bun run --cwd apps/mobile test -- src/account src/providers`, gave 4 failed and 4 passed: the guard shape was wrong and the catalogue keys were missing.

Green:

- `bun run --cwd apps/mobile test`: 22 suites, 163 tests passed. This includes:
  - the guard matrix for every state;
  - the banned screen in en (reason, "Until December 1, 2026", leave) and in pt-BR (no end date, no reason line);
  - the deletion-scheduled screen in es (date, `openWeb("/login")`, logout);
  - the i18n parity test.
- `typecheck` is clean, and `lint` (`biome ci .`, 93 files) is clean.

## Limits

- Nothing calls `session.observeMe(me)` yet, so the deletion-scheduled state is entered only once a `GET /me` read exists, which comes with P5.B. The screen and the routing are ready for it.
- I did not add `banned.tsx` or `deletion-scheduled.tsx` to the placeholder table in `navigation.test.tsx`. Both read the real runtime, and their bodies are covered in `src/account/blocked.test.tsx`.
- Not run on a device (owner-gated).

## Changelog line (for the specialist)

- Added: the Android app shows a suspension screen with the reason and end date when an account is banned, and a deletion-scheduled screen that links to the web login to cancel.

ROADMAP: no new box. This closes the screen part of P5.A.3.2, which is already ticked, and the deletion-scheduled screen named in P5.B.4.3.

## Review (especialista, 2026-10-31)

verdict: APROVADO
re-run: see gate below (merged as one ordered stack: root-wiring, P5.A.4.1, P5.A.4.2, P5.A.4.3, blocked-screens)
checked: SDD 10 §6.2, SDD 03 §8; one guard group per state, deletion-scheduled no longer gets the app; cancel opens web /login from the <origin>/app/v1 base like challengeUrl; UTC dates; i18n parity
gate on main after the stack merge (3e10c99): `bun run --cwd apps/mobile catalogues` no diff (437 keys per locale); mobile lint ok, tsc ok, jest 26 suites / 198 passed; packages/i18n parity 5 passed; web vitest 28 files / 111 passed; `go test ./app/i18n/... ./tests/i18n/...` ok
merge: 3e10c99
PRONTO_PARA_MERGE agent/senior/FIX-mobile-blocked-screens
