# Review P0.F.4.1 (rodada: 0)
Report: C:\dev\hexascore-app\hexascore-vukem\.orchestration\reports\P0.F.4.1.md
Branches: spike `agent/senior/P0.F.4.1` (C:\dev\hexascore-app\wt\spike-senior); app `agent/senior/P0.F.4.1` (pushed).
Re-run in wt\spike-senior: export TEST_DATABASE_URL (postgres admin URL; password via docker inspect hexascore-spike-postgres-1); `go test ./app/http/middleware/ ./tests/ -run "APIHost"`; `set -a; . ./.env; set +a; API_URL=http://api.localhost:3001/v1 MSYS_NO_PATHCONV=1 vukem openapi:export -prefix /api/v1 | jq -r '.servers[0].url'`. Check docs/spike-report.md and VF-034 on the app branch.
