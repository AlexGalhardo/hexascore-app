STATUS: concluida
Branch: agent/senior/FIX-e2e-signup-load (from origin/main)
Root cause: signup rule GlobalLimit=30/h across all IPs; a full e2e run signs up >30 accounts so /challenge hands out AttackBits=21 (seconds to minutes under 3 workers) -> 60s POW_TIMEOUT hits on later signups.
Fix: optional env CHALLENGE_GLOBAL_LIMIT (Challenger.GlobalLimit) overrides the global threshold; set to 100000 only in docker-compose.yml app env. Production default unchanged (unset). Per-IP rule and 18-bit default untouched.
Files: app/http/middleware/challenge.go, routes/challenge.go, docker-compose.yml, tests/pow_test.go (TestSecurity_A07_PowGlobalLimitOverride), CHANGELOG.
Acceptance: go test -tags integration -run TestSecurity_A07_Pow: 3 PASS; full playwright on hx-senior (3 workers): 444 passed, 0 failed, 0 flaky (3.3m).
Risk: none for prod; the compose file is e2e-only.
