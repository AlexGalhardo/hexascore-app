# FIX-legacy-import-no-login
status: concluida
branch: agent/senior/FIX-legacy-import-no-login @ f1bb3f5 (e6d3700 opt-in flag, b8a5cb2 runtime check, f1bb3f5 lint style)
files: app/console/import_legacy.go, database/legacyimport/{preflight.go,catalog.go,allow_test.go}
acceptance:
- cherry-pick 661114b and 6f99082 onto main; CHANGELOG conflicts resolved by keeping main's file (lines below)
- `go test -tags integration ./database/legacyimport/...` → ok
- golangci-lint on app/console and database/legacyimport: see validation
validation: build ok; tests ok; lint after fixing golines and errors.AsType (modernize) in preflight.go
decision on 263a3a7 (cmd/activity, cmd/dbcounts wip tools): DROPPED, not cherry-picked. Throwaway diagnostics from the import worktree with no tests, outside app/ layout (repo has only cmd/app); the commit stays reachable on agent/implementer/import-allow-no-login until the orchestrator deletes that branch.
changelog: Added: `hexascore:import-legacy --allow-no-login` opts in to users with no password, Google link or passkey (duplicate emails and a two-factor flag without a row still abort). Fixed: legacy catalog import stores a runtime or page count only when it fits the items check; out-of-range values become empty instead of aborting the batch.
roadmap: no box
risks: the opt-in imports accounts nobody can log into; owner decides when to use it (production import is an owner gate).

## Review (especialista, 2026-10-31)
verdict: CORRIGIDO_E_APROVADO
re-run: `go test -tags integration ./database/legacyimport/...` → ok (6.6 s); golangci-lint (integration tag; database/legacyimport, app/console) 0 issues
checked: sdd 11 §7.6 pre-flight (the opt-in drops only `no_login_method`; duplicate emails and two-factor without a row still abort; a non-preflight error passes through), the flag is off by default and the production import stays an owner gate, sql (bounds are constants in the statement, no user input), secrets none, i18n n/a, scope ok (263a3a7 dropped, as decided)
fixes: 1. `TestCatalogStageDropsOutOfRangeNumbers`: the `boundedNumber` change (runtime 0, 99,999 pages → NULL instead of a batch abort) had no test (see the fix branch head)
gate on main after merge: fe00949 → go build ok, golangci-lint (integration tag) ./... 0 issues, go test -tags integration ./database/legacyimport/... ok (rest of main unchanged since the green eda3a96 gate)
PRONTO_PARA_MERGE agent/especialista/FIX-legacy-import-no-login-fix
