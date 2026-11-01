# FIX-api-host-only-api API host serves only the API
status: concluida
branch: agent/senior/FIX-api-host-only-api @ HEAD (from main 9e6a0fc)
files: `app/http/middleware/api_edge.go`, `bootstrap/app.go`, `tests/api_edge_test.go`, `docs/sdd/05-public-api.md`, `CHANGELOG.md`
acceptance:
- `go test -tags integration ./tests -run 'TestAPIEdge|TestSecurity_A05_ApiHost'` -> ok (the new tests fail without the change: index, unknown path and all 5 A05 cases)
- `go test ./app/http/middleware` -> ok (CanonicalHost tests unchanged)
validation: go vet middleware + bootstrap ok
changelog: Fixed: API host serves only the API.
roadmap: none
decisions:
- `/` answers JSON `{name, version:"v1", docs:<APP_URL>/docs/reference, openapi}` instead of a 301 (SDD 05 section 6.1 records it). The real docs page is `/docs/reference`, not `/docs/api`.
- On `API_HOST`: `/v1/*` rewritten, `/api/v1/*` and `/up` pass, `GET|HEAD /` index, anything else 404 problem+json written inside APIEdge, before the web stack (no cookie, HTML or Inertia). Web hosts: `/v1/*` already fell through to a 404, now covered by a test.
risks / follow-ups:
- `APIEdge` gained an `appURL` parameter (only bootstrap calls it). Version in the index is the constant `v1` (no API_VERSION setting exists in code).
- Non-GET on `/` at the api host answers 404 not-found, not 405.
- Ships in the next release; the branch is from local `main`, mergeable alone.

## Review (especialista, 2026-10-31)
verdict: CORRIGIDO_E_APROVADO
re-run (main 4fea1a3 merged in, clean): `go test -tags integration ./tests -run 'TestAPIEdge|TestSecurity_A05_ApiHost|TestOpenAPI'` → ok; `go test ./app/http/middleware` → ok; golangci-lint middleware + bootstrap → 1 golines, fixed
checked: SDD 05 §6.1, A05 (api host never reaches the web stack: no cookie/HTML/Inertia; 404 body JSON-encoded, path only in `instance`), CORS unchanged, secrets none, scope (CHANGELOG line written by the author under Added), tests (5 A05 cases)
fixes: 1. golines wrap of the APIEdge call in bootstrap/app.go (agent/especialista/FIX-api-host-only-api-fix)
note: also in hotfix v0.12.4 by cherry-pick; expect a CHANGELOG/api_edge.go overlap when origin/main is merged back
gate on main after merge (cf25938): go build ok
PRONTO_PARA_MERGE agent/especialista/FIX-api-host-only-api-fix
