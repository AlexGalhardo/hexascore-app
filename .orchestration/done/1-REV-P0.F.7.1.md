# Review P0.F.7.1
Task: P0.F.7.1. Repo: spike + app docs. Branch: `agent/senior/P0.F.7.1` (spike commit 2270240 in wt/spike-senior, branched from spike main; app branch pushed). rodada: 1
Report: C:\dev\hexascore-app\hexascore-vukem\.orchestration\reports\P0.F.7.1.md
Re-run: `go test ./tests -run TestAPIContract` (needs TEST_DATABASE_URL), `go test ./app/http/problems/`; `app openapi:export -prefix /api/v1` and `/app/v1`; curl with Host api.localhost.
Note: spike branch is based on main, not on the P0.F.2.3 branch (independent files except routes/registry.go).
