# QA-ci-local-script: report

Branch `agent/senior/QA-ci-local-script` @ `b6b671b` (3 commits on `main` 713cad2). Files: `scripts/ci/local.sh` (new), `qa/schemathesis/ci.sh`.

## What it does

`scripts/ci/local.sh [--only a,b] [--skip a,b] [--base REV] [--tree] [--keep]` runs every job of `ci.yml` except e2e, in parallel, on a clean detached checkout of HEAD (warns when the tree is dirty; `--tree` tests the working tree).

- Jobs: lint, test-go, test-go-integration, test-ts (the three halves of CI `test`), migrations, vuln, deps, setup-windows (Windows hosts only, else SKIP), build, openapi, lighthouse, mobile. e2e is printed as SKIP (owner).
- Pins are read from the workflow (`GOLANGCI_LINT_VERSION`, `MARKDOWNLINT_VERSION`, `OSV_SCANNER_*`, Postgres image). osv-scanner is the checksum-verified pinned release, cached in `~/.cache/hexascore-ci` and run in the pinned Postgres image. No dependency added.
- One log per job in `.ci-local/logs/<job>.log` (ignored by `*.log`), summary table, last 25 lines of each failed log, exit 1 if any job fails.
- golangci-lint retries every 10 s (90 times) while another run holds its lock (exit 3, "parallel golangci-lint").
- Isolation per AGENTS.md: databases `hx_ci_<id>_*` on the shared Postgres 5433, free ports in 3100-3899, image `hexascore:ci-local-<id>`, compose project `hexascore-qa-<id>` (never touches project `hexascore`), everything removed on exit (`--keep` keeps it).
- Two waves: test-ts and mobile first (vitest forks and jest 5 s timeouts failed under the load of go -race and the Docker build), then the heavy jobs.
- `qa/schemathesis/ci.sh`: psql now uses `$DATABASE_URL` (was hardcoded 5432/`hexascore`); on Docker Desktop it publishes the app port instead of `--network host` (the host cannot reach a host-network container there). CI (Linux) path unchanged.

## Evidence

- Green (script on top of `main` 713cad2, full run `--base HEAD~3`, 16 min wall): lint, test-go-integration, test-ts, migrations, vuln, setup-windows, build, openapi (incl. Schemathesis and k6), lighthouse, mobile all PASS.
- Red: throwaway commit with `const broken: number = "not a number"` in `apps/mobile/src/broken.ts`, `--only mobile,vuln`: `mobile FAIL`, `tsc` TS2322 in the log, `FAILED on bbf4937`, exit 1; vuln PASS.
- Two jobs were not green on `main` in that run, both real findings, not script bugs:
  1. `deps`: osv-scanner reports GO-2026-5932 in `golang.org/x/crypto 0.58.0` (no fixed version). CI `deps` would be red too; needs an entry in `osv-scanner.toml` (owner/orchestrator decision) or a bump.
  2. `test-go`: `app/ssr` `TestSuperviseRestartsUntilCancelled` flaked under load ("restarted 2 times, want >= 3"; the helper could not open `runs` in the temp dir). Passed in the earlier full run (run 2).
- Not exercised: the golangci-lint lock retry path (no concurrent run happened); bun 1.4.4 is installed vs `packageManager` 1.4.3 (the script warns).
- Per the coordinator, no further full runs.

## AGENTS.md line (add to the Commands table; the Release gate paragraph can say "run `scripts/ci/local.sh`")

```text
| `scripts/ci/local.sh [--only a,b] [--skip a,b] [--base REV] [--tree] [--keep]` | release gate in one command: every `ci.yml` job except e2e, in parallel, on a clean checkout of HEAD, with its own databases, ports and compose project; one log per job in `.ci-local/logs/`, summary table, exit 1 if any job fails (needs Docker, the compose Postgres on 5433, go, bun, golangci-lint; `setup-windows` runs only on Windows) |
```

## Changelog line

`Added: scripts/ci/local.sh runs every CI job locally in parallel (release gate); qa/schemathesis/ci.sh honours DATABASE_URL and works on Docker Desktop.`

## Review (especialista, 2026-10-31)
verdict: CORRIGIDO_E_APROVADO
re-run (targeted, owner rule 8.2): `bash -n` on both scripts ok; `scripts/ci/local.sh --help` ok; no full run (author's green/red evidence above stands; shellcheck not installed)
checked: cleanup trap removes only its own containers, `hexascore-qa-<id>` projects, `hx_ci_<id>_*` databases, image and detached worktree; the shared project `hexascore` only gets `up`, never `down`; pins read from the workflow; `qa/schemathesis/ci.sh` keeps the CI (Linux) behaviour (`$db` defaults to the same 5432/hexascore URL the hardcoded psql used)
fixes: 1. the branch committed `qa/schemathesis/__pycache__/ip_hook.cpython-314.pyc`; removed and `__pycache__/` added to `.gitignore` (4f3fedf) 2. AGENTS.md: Commands row and Release gate pointer added at merge
follow-ups (from the author's run, for the orchestrator): osv-scanner GO-2026-5932 in golang.org/x/crypto 0.58.0 makes CI `deps` red (needs an `osv-scanner.toml` decision); `app/ssr` TestSuperviseRestartsUntilCancelled flakes under load
PRONTO_PARA_MERGE agent/especialista/QA-ci-local-script-fix (merged 43dd3b5)
