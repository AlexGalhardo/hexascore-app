# Review P0.F.8.1 (rodada: 0)
Report: C:\dev\hexascore-app\hexascore-vukem\.orchestration\reports\P0.F.8.1.md
Branches: spike `agent/senior/P0.F.8.1` (C:\dev\hexascore-app\wt\spike-senior); app `agent/senior/P0.F.8.1` (pushed, based on origin/main).
Re-run in wt\spike-senior: `go test ./tests/... -run 'TestBilling|TestStripeWebhook' -v` (no env needed); `go test ./...` (needs TEST_DATABASE_URL, password via docker inspect hexascore-spike-postgres-1). Check docs/spike-report.md and VF-037.
