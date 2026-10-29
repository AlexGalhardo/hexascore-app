# senior handoff (twenty-fifth session)
Delivered (pushed, reports + 1-REV queued): FIX-highlight-score-scale, P2.D.5.1, P1.H.4.1.
Pitfalls: TEST_DATABASE_URL=postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable; Git Bash mangles /api/v1 args (set MSYS_NO_PATHCONV=1); python writes CRLF on Windows (use sed -i 's/\r$//'); openapi:export needs APP_KEY + DATABASE_URL env; heredocs with apostrophes break Bash; run `bunx biome check --write resources/js` before bun run lint.
Open: P1.H.4.1 Python tests not in CI; osv-scanner/CI jobs never run live.
