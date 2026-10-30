STATUS: APROVADO (rodada 1)
Branch: agent/senior/FIX-e2e-signup-load (diff: challenge.go override only when env set and rule has a global limit; routes reads CHALLENGE_GLOBAL_LIMIT; compose app env only; test).
Re-run: go test -tags integration ./tests -run TestSecurity_A07_Pow -v -> PowSignup/PowTokens/PowGlobalLimitOverride PASS.
Playwright on isolated hx-especialista (127.0.0.1:3200, 3 workers): 1st run 66 failed, all "no mail ... in hexascore-app-1 logs" (my env: E2E_APP_CONTAINER unset, not the fix); rerun with E2E_APP_CONTAINER=hx-especialista-app-1 on the SAME db (already >440 signups, far above 30/h): 444 passed (3.5m), 0 failed.
Security: prod default unchanged (unset -> rule limit); malformed value -> Atoi 0 -> ignored (fail closed to default). Per-IP rule untouched.
Nit (no action): var not in .env.example; compose comment documents it.
PRONTO_PARA_MERGE agent/senior/FIX-e2e-signup-load
