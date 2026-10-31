# DOC-commands-check Existence of every AGENTS.md "Commands" row
status: concluida (report only; no edits; nothing executed that deploys or spends)
branch: none (main @ 246ced8 at read time)
files: .orchestration/reports/DOC-commands-check.md
acceptance (checks run from `C:\dev\hexascore-app\hexascore-vukem`):
- `bun -e` read of `package.json` scripts → build:docs, lint, typecheck, test, build:ssr, openapi:lint, sdk:generate, sdk:generate:python, e2e: all yes. `apps/mobile/package.json` → lint, typecheck, test, catalogues: all yes.
- file checks (`[ -e ]`) → yes: go.mod, .golangci.yml, commitlint.config.ts, redocly.yaml, lighthouserc.json, docker-compose.yml, qa/schemathesis/ci.sh, scripts/openapi-check.sh, scripts/ci/qa-up.sh, scripts/legacy/twofactor-export.ts, scripts/release.sh, .config/cliff.toml, setups/setup-windows.ps1, setups/setup-unix.sh, openapi/api-v1.json, sdks/ts/src/gen, sdks/python/hexascore, resources/docs/docs.json, docs/runbooks/restore.md, docs/runbooks/migrations.md
- Go console commands (`rg --fixed-strings` in `*.go`) → hexascore:import-legacy (app/console/import_legacy.go), hexascore:cutover-verify (app/console/cutover_verify.go), backup:restore (app/console/backup_restore.go), backup:drill (app/console/backup_drill.go), hexascore:predeploy (app/console/predeploy.go), admin:grant / admin:revoke / user:reset-2fa (app/console/admin_roles.go): all found
- `vukem --help` (v0.32.0 on PATH) → serve and build listed; `migrate` runs in the app (help line 47, as the table says); `openapi:export` accepted (`vukem openapi:export --help` prints its usage)
- `setups/setup-windows.ps1` → `param([switch]$NoServices)` present
- `docker-compose.yml` → services under profile `app` incl. `app-init`
- decision table below

| command (AGENTS.md row) | exists | evidence |
| --- | --- | --- |
| `vukem serve`, `vukem build` | yes | vukem --help |
| `vukem build` + `bun run build:ssr` | yes | build:ssr script `vite build --ssr resources/js/ssr.tsx` |
| `go test ./...` | yes | go.mod |
| `golangci-lint run` | yes | .golangci.yml (not run) |
| `golangci-lint run --build-tags integration ./...` | yes (config) | .golangci.yml; flags not verified |
| `bunx commitlint --from --to` | yes | commitlint.config.ts, @commitlint/cli in package.json |
| `bun run build:docs` | yes | script + resources/docs/docs.json |
| `bun run lint` / `typecheck` / `test` | yes | root scripts |
| `bun run --cwd apps/mobile lint/typecheck/test`, `catalogues` | yes | apps/mobile/package.json |
| `vukem openapi:export -prefix` | yes | vukem openapi:export accepted; scripts/openapi-check.sh |
| `qa/schemathesis/ci.sh [image]` | yes | file |
| `bun run openapi:lint` | yes | script; redocly.yaml |
| `bun run sdk:generate` / `sdk:generate:python` | yes | scripts; sdks/ts/src/gen, sdks/python/hexascore |
| `bun run e2e` | yes | script `playwright test -c qa/e2e/playwright.config.ts` (not run, port 3000 busy) |
| `bunx --bun @lhci/cli@0.15.1 autorun` | file yes, command not verified | lighthouserc.json present; @lhci/cli not in package.json (as the row says); not run |
| `docker compose up -d --wait` | yes | docker-compose.yml (not run; stack on 3000 is someone else's) |
| `setups/setup-windows.ps1 [-NoServices]` | yes | param present |
| push to `main` (Railway) | not verified | policy row; no command run |
| `railway up --service app --detach` | not verified | not run (deploy forbidden without owner ask) |
| `./app hexascore:import-legacy ...`, `hexascore:cutover-verify`, `backup:restore`, `backup:drill`, `hexascore:predeploy`, `admin:grant`, `admin:revoke`, `user:reset-2fa` | yes (in Go source) | app/console/*.go; `./app` itself is not in the repo, see finding 1 |
| `scripts/release.sh vX.Y.Z` | yes | file; git-cliff not checked |
| `setups/setup-unix.sh` | yes | file |

findings:
1. `./app` is not a file in the repo. `app/` is the Go package tree. The binary is built by the Dockerfile (`go build -trimpath ... -o /out/app ./cmd/app`, Dockerfile:36), so locally `./app` only exists after `go build -o app ./cmd/app`, which no script runs. The table should say so or name `go run ./cmd/app <command>`.
2. The AGENTS.md note `go run . list` (from my brief) does not apply: there is no `main` package at the repo root (`ls *.go` empty). The entry point is `./cmd/app`.
3. `bunx --bun @lhci/cli@0.15.1` and `railway up` are rows I did not execute (network and deploy policy); they are marked "not verified" rather than yes.

validation: lint n/a, typecheck n/a, tests n/a, browser n/a
changelog: none
roadmap: no box
risks / follow-ups: `golangci-lint` flags and `--build-tags integration` were not run. The `vukem` binary on PATH is v0.32.0 and the framework checkout at `C:\dev\hexascore-app\vukem` was only searched, not built.

## Review (especialista, 2026-10-31)

verdict: APROVADO (report only)
re-run (sampled): `grep -c "o /out/app ./cmd/app" Dockerfile` → 1 (finding 1 holds: `./app` is the built binary, not a repo file); `package.json` scripts `build:ssr`, `e2e` present; `golangci-lint run --build-tags integration ./...` and `bun run e2e` ran this session (gate @ 57497a9, item/ratings/passkeys specs)
added fact: `bunx --bun @lhci/cli@0.15.1 autorun` fails to install on this Windows host (bunx git-clone error on the scoped package); `bunx lighthouse@12.6.1 <url> --form-factor=mobile` works. Worth a note in the AGENTS.md row (orchestrator or a DOC row)
checked: claims cite the command or file; nothing deploys; no edit
