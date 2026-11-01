# FIX-community-mail-optout-ui community email opt-out on the profile

status: concluida
branch: agent/senior/FIX-community-mail-optout-ui @ 2eab02c (stacked on agent/senior/FIX-community-create-kinds-ui @ b23b0a1)
files: `app/accounts/profile.go` (`Profile.CommunityMailOptOut`, `UpdateCommunityMail`), `app/requests/profile.go` (`ProfileNotificationsSchema`), `app/controllers/profile/profile.go` (`notifications` prop, `UpdateNotifications`), `routes/profile.go` (`PATCH /profile/notifications`, same `required, verified` stack as privacy), `tests/testdata/route_inventory.txt`, `app/library/export.go` (`profile.json` gains `communityMailOptOut`), `app/accounts/lang/*/auth.json` (`auth.profile.notifications_saved`), `resources/js/pages/Profile/Index.tsx` (section "Email notices"), `resources/lang/*.json` (`account.notifications.*`), `database/migrations/20261031203100_add_community_mail_opt_out.go`, `tests/profile_test.go`, `tests/export_test.go`.
sdd: SDD 04 FR-04-14 (in-app notice always, email optional); SDD 03 §6.4 export.

decisions:

- A separate endpoint and form, not a third privacy switch: `UpdatePrivacy` is also called by `PATCH /app/v1/me` with the full body, so a new privacy field would reset the opt-out every time the Android app saved its profile. The test saves privacy after the opt-out and checks it survives.
- Migration `20261031203100` (mine, P4.B.3.1, not on `main` yet) moves from the Postgres-only raw `ALTER` to the schema builder so the column exists on SQLite, where the profile tests run. Same ID, same Postgres column (`boolean NOT NULL DEFAULT false`); a database that already ran the old version is unchanged. The specialist may prefer to fold this into the P4.B.3.1 review.
- The switch is an opt-out ("Do not email me about my community suggestions"), matching the personalisation switch.

acceptance:

- `go test ./tests -run 'TestProfile|TestSecurity_A01_HexaRouteStack|TestApp' -count=1` → ok (`TestProfileCommunityMailOptOut` red before: prop missing; now default false, saved true, survives a privacy save).
- `go test -tags integration ./tests -run 'TestProfile|Export|Lgpd|LGPD|Notif|Migration' -count=1` → ok (export `profile.json` carries `communityMailOptOut`; community notify tests unchanged).
- `bun run lint`, `bun run typecheck` → clean; `bunx vitest run resources/js/lib resources/js/pages` → 35 passed (i18n parity included).

validation: `golangci-lint fmt` clean; `golangci-lint run --timeout 15m --build-tags integration` accounts, controllers/profile, requests, library, routes, tests, migrations → 0 issues (the first run hit the default timeout under host load).
changelog: Added: profile switch to stop community suggestion emails (in-app notices stay); the data export includes it
roadmap: none (follow-up of P4.B.3.1)
risks / follow-ups:

- The Android app has no switch for it (out of scope; `GET /app/v1/me` does not expose it).
- No page unit test for `Profile/Index.tsx` (none exists); checked in the browser, see below.

QA (isolated stack `hexascore-qa-seniora` on :3121, image built from 2eab02c, DemoCatalog, `FLAG_COMMUNITY_SUBMISSIONS=true`):

- `/profile` shows the "Email notices" section; ticking "Do not email me about my community suggestions" and saving → `select community_mail_opt_out from users where email='admin@example.test'` → `t`.
- The same stack covers FIX-community-create-kinds-ui (see that report). No screenshot: the browser tool may not write outside the repo roots; the DB rows are the evidence. Stack stopped and database dropped.

## Review (especialista, 2026-10-31)
verdict: CORRIGIDO_E_APROVADO
re-run: `go test ./tests -run 'TestProfile|TestSecurity_A01_HexaRouteStack|TestApp'` (SQLite) → ok; `bun test scripts/ci/migration-policy.test.ts` → 3 pass; `bun run typecheck`, `bun run lint` clean; `bunx vitest run resources/js/lib resources/js/pages` → 12 files, 36 passed; `golangci-lint run --build-tags integration` on database, accounts, controllers/profile, requests, library, routes, tests → 0 issues
checked: SDD 04 FR-04-14, SDD 03 §6.4 export, authz (same `required, verified` stack as privacy; route inventory updated), secrets none, sql (bound placeholders), i18n (en/pt-BR/es), scope, tests
fixes (agent/especialista/FIX-community-mail-optout-ui-fix, merge d394d06): 1. migration 20261031203100 (P4.B.3.1, already on main) restored unchanged; the SQLite column moved to a new migration `20261101140000_add_community_mail_opt_out_sqlite` (no-op on Postgres, schema builder elsewhere) registered after 20261101130000 2. lang conflict with main (mobile `account.*` keys) resolved keeping both sides
pending (Docker down): `go test -tags integration ./tests -run 'TestProfile|Export|Lgpd|LGPD|Notif|Migration'` and the migrations round trip
gate on main after merge (ac21bc3): go build ok; full gate waits Docker
PRONTO_PARA_MERGE agent/especialista/FIX-community-mail-optout-ui-fix
