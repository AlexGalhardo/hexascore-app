task: P1.A.4.3
repo: app, branch agent/senior/P1.A.4.3
report: C:\dev\hexascore-app\hexascore-vukem\.orchestration\reports\P1.A.4.3.md
acceptance: `docker compose up -d --wait postgres; TEST_DATABASE_URL=postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable go test ./app/console -run TestMigrateLockedSerialises -v; golangci-lint run`
rodada: 1
