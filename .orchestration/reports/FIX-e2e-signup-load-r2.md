STATUS: parcial
Code already on dev (CHALLENGE_GLOBAL_LIMIT in app/http/middleware/challenge.go, routes/challenge.go, docker-compose.yml, tests/pow_test.go).
Acceptance so far: `go test ./tests/ -run TestSecurity_A07_PowGlobalLimit` → ok (override to 2 trips attack bits; value "nope" refuses boot). Compose sets CHALLENGE_GLOBAL_LIMIT=100000. Production stays unset.
Not done: full Playwright suite twice on a rebuilt image. The container on :3000 predates this knob, so a run against it would not exercise the shipped binary.
