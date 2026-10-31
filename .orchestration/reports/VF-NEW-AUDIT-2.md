# VF-NEW-AUDIT-2 default app name from the project
status: concluida
branch: fw: fix/vf-app-name-default @ 0296b955 (not pushed, no PR; worktree wt/fw-VF-NEW-AUDIT-2)
files: app.go, options_runtime.go, http_test.go, internal/cli/starter/bootstrap/app.go.tmpl, internal/cli/new_test.go, site docs (configuration.mdx, reference-new.mdx), last-updated.json
acceptance:
- `go test -run TestWithName .` and `go test -run TestWriteStarter ./internal/cli` -> ok (WithName default, APP_NAME wins, starter writes vukem.WithName)
- `task ci` -> ci=0; `task api:check` -> api=0
validation: lint ok, tests ok
changelog: see commit message (CHANGELOG is generated per release)
roadmap: no box
risks / follow-ups: New exported API vukem.WithName (additive, api:check ok). Only the starter bootstrap passes it; existing apps keep "Vukem" until they add WithName or set APP_NAME. The admin title already defaults to "Admin" and does not read Config.Name (the "Vukem admin" text came from the app). Specialist: push, PR, merge on green CI. last-updated.json already refreshed in the commit (committer date preserved on amend).
vukem-feedback status line: VF-NEW-AUDIT-2 -> fixed in vukem branch fix/vf-app-name-default (vukem.WithName + starter default), pending PR/release.
PR: vukem #115 open (pushed, awaiting GitHub CI)

## Review (especialista, 2026-10-31)
verdict: APROVADO
re-run: `go test -count=1 -run 'TestWithNameIsTheDefaultOfAppName|TestWriteStarter|TestNew' . ./internal/cli/` ok (cli 557 s, end-to-end starters compile with `vukem.WithName`); `go run ./internal/tools/apicheck` rc 0 (no v1 tag yet, additive)
checked: APP_NAME still wins (test), option only fills the unset variable after config.Load, starter name charset cannot break the Go string literal, docs (new reference, configuration)
fixes: branch merged origin/main (app.go options conflict: kept both `name` and `fuzzAuth`), last-updated.json regenerated; no code change
PR: #115
