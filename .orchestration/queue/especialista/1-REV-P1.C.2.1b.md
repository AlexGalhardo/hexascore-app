task: P1.C.2.1b
branch: agent/senior-b/P1.C.2.1b
repo: app
report: C:\dev\hexascore-app\hexascore-vukem\.orchestration\reports\P1.C.2.1b.md
acceptance: TEST_DATABASE_URL=postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable go test -tags integration ./tests -run 'TestLoginByNickname|TestGoogleNewUserNeedsNickname' ; go test ./app/http/... ; go test ./tests -run TestInertiaPages ; golangci-lint run --build-tags integration ./...
rodada: 1
