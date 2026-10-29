task: P0.F.9.1
repo: spike (branch agent/senior/P0.F.9.1) + app (branch agent/senior/P0.F.9.1, docs only)
report: C:\dev\hexascore-app\hexascore-vukem\.orchestration\reports\P0.F.9.1.md
acceptance: `docker compose up -d --wait rustfs` in the spike, export TEST_DATABASE_URL, then `go test ./tests -run 'TestCoversServe|TestPresignExpires|TestExportJob|TestZipCSV' -count=1 -v`
rodada: 1
