# DOC-vf-apiproblem-code: file VF-NEW-P5.A.3.1 (ApiProblem has no code field)
status: concluida
branch: agent/estagiario/DOC-vf-apiproblem-code @ 31282e5
files: vukem-feedback.md (index row + entry VF-NEW-P5.A.3.1)
acceptance:
- `git show --stat 31282e5` → vukem-feedback.md, 9 insertions
- `grep -c "VF-NEW-P5.A.3.1" vukem-feedback.md` → 3 (index row, heading, workaround/request context)
- `git grep -n ApiProblem agent/senior/P5.A.3.1 -- packages/app-client apps/mobile/src/api` → packages/app-client/src/api.ts:305-313: type has no `code`; apps/mobile/src/api/problems.ts:18 reads `code` through an `unknown` cast
- docs: <https://vukem.dev/docs/rest-apis/overview> "Responses and errors" (read 2026-10-31): RFC 9457 documented; example has no `code`; no codegen section → entry marked [NOT IN VUKEM DOCS]
validation: markdown is not linted (`.orchestration/**` ignored); vukem-feedback.md is the only repo file changed
changelog: none (report only)
roadmap: no box
risks / follow-ups: FIX-mobile-error-keys still waits for P5.A.3.1 to merge. Workaround comment not yet in code (lands with the mobile task).

## Review (especialista, 2026-10-31)
verdict: APROVADO
re-run: `bunx markdownlint-cli2@0.24.0 vukem-feedback.md` -> 0 errors
checked (sampled): `ApiProblem` in `packages/app-client/src/api.ts` on agent/senior/P5.A.3.1 has no `code` (lines 305-313); the app does send `code` as a problem extension (`app/api/quota_middleware.go:97`, `app/accounts/passkeys.go:642`); index row and `VF-NEW-<task-id>` naming follow the file rules
PRONTO_PARA_MERGE agent/estagiario/DOC-vf-apiproblem-code (merged at the sha in the board row)
