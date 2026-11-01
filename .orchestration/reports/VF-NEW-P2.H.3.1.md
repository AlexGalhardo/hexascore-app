# VF-NEW-P2.H.3.1 doc-only enum tag
status: concluida
branch: fw: fix/vf-doc-enum-tag @ b98da6cc (not pushed, no PR; worktree wt/fw-VF-NEW-P2.H.3.1)
files: openapi/schema.go, openapi/openapi_test.go, site/content/docs/rest-apis/openapi.mdx, last-updated.json
acceptance:
- `go test ./openapi ./jsonrpc` -> ok (TestEnumTagDocumentsWithoutValidating: enum in path schema, validate oneof wins, value outside the tag answers 200)
- `task ci` -> ci=0; `task api:check` -> api=0
validation: lint ok, tests ok
changelog: see commit message (CHANGELOG is generated per release)
roadmap: no box
risks / follow-ups: New struct tag `enum:"a,b"` (comma separated), applied by openapi.Fields, so it also documents query and body fields and jsonrpc params. Implemented as a synthetic oneof rule for the schema only; z never reads it. The api.Op option to map validation errors to a status was not added. Specialist: push, PR, merge on green CI. last-updated.json already refreshed in the commit (committer date preserved on amend).
vukem-feedback status line: VF-NEW-P2.H.3.1 -> fixed in vukem branch fix/vf-doc-enum-tag (enum tag), PR open.
PR: vukem #117 open (pushed, awaiting GitHub CI)

## Review (especialista, 2026-10-31)
verdict: APROVADO
re-run: `go test -count=1 -run 'TestEnumTag|TestFields|TestSchema' ./openapi/ ./jsonrpc/` ok
checked: `enum` tag only feeds openapi.Fields (schema), runtime validation untouched (test: out-of-enum value answers 200); validate oneof wins; jsonrpc uses Fields only for names; docs
fixes: branch merged origin/main, last-updated.json regenerated; no code change
PR: #117
PR: #117 squash 7f9b9593 (33/33)
