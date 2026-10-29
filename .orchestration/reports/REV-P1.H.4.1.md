REV P1.H.4.1: APROVADO (branch already merged in origin/main). Branch: agent/senior/P1.H.4.1.
Re-run on the branch: bash scripts/openapi-check.sh -> exit 0; tampered openapi/app-v1.json -> exit 1 (restored); sdks/ts bun install --frozen-lockfile + typecheck ok, bun test 3 pass 0 fail; `bun run sdk:generate` leaves no git diff (gen in sync with golden); tsc --noEmit -p packages/app-client -> 0 errors; golangci-lint -> 0 issues; bun lint/typecheck clean, test 57 passed; markdownlint "**/*.md" 0 issues.
Checked: CI lint job has setup-go for the goldens step; SDK install uses --ignore-scripts; dev deps pinned exact (hey-api 0.99.0, typescript 6.0.3); no keys committed; sdks/ts/node_modules ignored; apiKey required, Bearer header only.
Not re-run: Python pytest (uv not installed here; same as author). Python SDK has no CI coverage -> risk, needs a pinned setup-uv decision (orchestrator).
Nits (not fixed): openapi-check.sh leaks its mktemp file if `go run` fails; double blank line.
