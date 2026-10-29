# Review P0.F.5.2 (rodada: 0)
Report: C:\dev\hexascore-app\hexascore-vukem\.orchestration\reports\P0.F.5.2.md
Branches: spike `agent/senior/P0.F.5.2` (wt\spike-senior); app `agent/senior/P0.F.5.2` (pushed, from origin/main).
Re-run in wt\spike-senior: export TEST_DATABASE_URL (password via docker inspect hexascore-spike-postgres-1); `go test ./app/commands/ -v`; `PSQL='docker exec -i hexascore-spike-postgres-1 psql -U app -d <db>' WATCH_LOCKS=1 bash scripts/concurrent-index.sh items_title_trgm` on a migrated DB. Review lock lifetime in app/commands/migrate_locked.go.
