# VF-NEW-P4.A.3.1b flag name rules and admin routes docs
status: concluida
branch: fw: fix/vf-flags-docs @ b8643f57 (not pushed, no PR; worktree wt/fw-VF-NEW-P4.A.3.1b)
files: flags/flags_test.go, site/content/docs/digging-deeper/feature-flags.mdx, admin.mdx, last-updated.json
acceptance:
- `go test ./flags` -> ok (new boundary rows: dotted name, leading digit, 101 characters)
- `task ci` -> ci=0; `task api:check` -> api=0
validation: lint ok, tests ok
changelog: see commit message (CHANGELOG is generated per release)
roadmap: no box
risks / follow-ups: Docs only plus tests. The /admin/_flags mount is driven by install:admin passing app.Flags() to admin.Config.Flags (documented); no opt-in API was added because the panel is already opt-in via Config.Flags. Specialist: push, PR, merge on green CI. last-updated.json already refreshed in the commit (committer date preserved on amend).
vukem-feedback status line: VF-NEW-P4.A.3.1b -> fixed (docs) in vukem branch fix/vf-flags-docs, PR open.
PR: vukem #119 open (pushed, awaiting GitHub CI)

## Review (especialista, 2026-10-31)
verdict: APROVADO
re-run: `go test -count=1 -run TestNewRejectsMisconfiguration ./flags/` ok
checked: documented regex matches `flags.nameRe`; "/admin/_flags only with at least one flag" matches `Panel.hasFlags`; docs only plus three name cases
PR: #119
