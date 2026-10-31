# FIX-moderator-2fa AdminGuard requires TOTP for moderators
status: concluida
branch: agent/senior/FIX-moderator-2fa @ 990c904 (stacked on agent/senior/P4.B.1.1, already in main)
files: app/http/middleware/admin_guard.go, app/http/middleware/admin_guard_test.go
acceptance:
- go test ./app/http/middleware -run 'Security_A0[17]_(Hexa|Moderator)' -> ok (TestSecurity_A07_ModeratorRequires2FA: 303 to /user/two-factor without TOTP, 200 with)
design: the guard's role check now admits admin and moderator, so both get the 30 min idle / 12 h session clock and the 2FA redirect on /admin. Role and ban checks stay with the panel gate.
changelog: Fixed: moderators now need confirmed 2FA to open /admin, like admins.
roadmap: no box
risks / follow-ups: moderators also fall under the admin session limits (SDD 13 section 7.5); intended. Test written alongside the one-line fix, red state not captured separately.
- also: tests/admin_users_test.go makeUser now confirms 2FA for moderators (else TestModerationGate "moderator flag on" would 303). `TEST_DATABASE_URL=postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable go test -tags=integration ./tests -run TestModerationGate` -> ok.

## Merge of local main (senior-a, 2026-10-31)

`git merge main` into `agent/senior/FIX-moderator-2fa` @ f88745c: picks up P4.B.1.1 as merged (311e39b, migration `20261031160000_allow_moderator_role.go`); the old `20261031150000` moderator migration is gone. `go test -tags integration ./tests -run 'Moderation|Migrat'` → ok.

## Review (especialista, 2026-10-31)

CORRIGIDO_E_APROVADO at f88745c: main is merged in, the only moderator migration is `20261031160000_allow_moderator_role.go`. Operator precedence in the guard (`err != nil || role != admin && role != moderator`) is correct. SDD 13 FR-13-5 still said admins only; it now names moderators and the new test (bd980cd).

Checks: `go test ./app/http/middleware` ok; `go test -tags integration ./tests -run 'Moderation|Admin|Security'` ok; golangci (integration tag) on the middleware and `./tests/` 0 issues.

PRONTO_PARA_MERGE agent/especialista/FIX-moderator-2fa-fix
