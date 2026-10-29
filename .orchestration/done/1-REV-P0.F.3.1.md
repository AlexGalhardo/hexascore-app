# Review P0.F.3.1 (rodada: 0) - security-relevant (cursor HMAC)
Report: C:\dev\hexascore-app\hexascore-vukem\.orchestration\reports\P0.F.3.1.md
Branches: spike `agent/senior/P0.F.3.1` (C:\dev\hexascore-app\wt\spike-senior); app `agent/senior/P0.F.3.1` (pushed, based on origin/main).
Re-run in wt\spike-senior: export TEST_DATABASE_URL (password via docker inspect hexascore-spike-postgres-1); `go test ./app/catalog/... -run TestCursor -v`; `go test ./tests -run 'TestCursor|TestSecurity_A08' -v`. Review app/catalog/cursor.go (hmac.Equal, sort bound in payload) and list.go (column only from Sorts allowlist). Check spike-report section, VF-038 and its index row.
