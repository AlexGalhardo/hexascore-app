STATUS: concluida
Branch: agent/implementer/FIX-e2e-signup-load-r2 (4747815, from origin/main c1edff7). Not pushed. Not merged.
Files: app/http/middleware/challenge.go, app/http/middleware/challenge_limit_test.go, routes/challenge.go, tests/pow_test.go, docker-compose.yml, CHANGELOG.md, docs/sdd/13-security.md
Knob already on dev (80f291d). This branch is only that change on the app/challenge global trip. Gzip and Lighthouse are not in it. Unset stays 30. Invalid values refuse boot.
Commands:
- go build ./... → exit 0
- golangci-lint run --allow-parallel-runners --build-tags integration ./... → 0 issues
- TEST_DATABASE_URL=postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable go test -tags integration -count=1 -timeout 30m ./... → exit 0, no FAIL (tests 187s)
- go test -tags integration -v ./app/http/middleware ./tests -run "TestGlobalLimitOverride|TestSecurity_A07_PowGlobalLimit" → PASS TestGlobalLimitOverride, TestSecurity_A07_PowGlobalLimitOverride, TestSecurity_A07_PowGlobalLimitRefusedAtBoot
- playwright twice, E2E_BASE_URL=http://127.0.0.1:3200, E2E_APP_CONTAINER=hx-implementer-app-1, image hexascore-app:e2e-implementer → 636 passed (5.8m), 636 passed (5.7m), exit 0 both, 0 signup failures
Risk: compose sets CHALLENGE_GLOBAL_LIMIT only on the local e2e app. P1.I.5.1 stays open (Lighthouse not run).
PRONTO_PARA_MERGE. Third playwright on the same image: 636 passed (5.9m), 0 signup failures. Review: .orchestration/reports/REV-FIX-e2e-signup-load-r2.md
Queue closed 2026-10-30: knob already on local main (challenge.go, pow_test.go, docker-compose.yml). Branch 4747815 was not merged again.
