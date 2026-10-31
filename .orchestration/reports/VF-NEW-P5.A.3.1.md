# VF-NEW-P5.A.3.1 ApiProblem code
status: concluida
branch: fw: fix/vf-apiproblem-code @ e9ef40b0 (not pushed, no PR; worktree wt/fw-VF-NEW-P5.A.3.1)
files: openapi/client.go, openapi/testdata/{shop,reserved}.client.ts, site/content/docs/rest-apis/client-generation.mdx, last-updated.json
acceptance:
- `go test ./openapi` -> ok (goldens regenerated with -update)
- `task ci` -> ci=0; `task api:check` -> api=0
validation: lint ok, tests ok
changelog: see commit message (CHANGELOG is generated per release)
roadmap: no box
risks / follow-ups: Generated ApiProblem now has `code?: string` and `[extension: string]: unknown`, so other extension members read as unknown. Apps must narrow extension values. Generated TS not typechecked here. Specialist: push, PR, merge on green CI. last-updated.json already refreshed in the commit (committer date preserved on amend).
vukem-feedback status line: VF-NEW-P5.A.3.1 -> fixed in vukem branch fix/vf-apiproblem-code (typed code member), PR open.
PR: vukem #118 open (pushed, awaiting GitHub CI)

## Review (especialista, 2026-10-31)
verdict: APROVADO
re-run: `go test -count=1 -run 'Client|TS' ./openapi/` ok (golden clients updated)
checked: `code?: string` plus `[extension: string]: unknown` index signature is compatible with every declared member; additive for consumers; docs table
PR: #118
