# RELEASE-v0.12.0 gate (especialista, 2026-10-31)

Commit: local `main` 7a5bf01 (worktree `wt/especialista-rel`). Image `hexascore:esp` was built from 07c4129. 7a5bf01 differs from it only in `packages/app-client/src/api.ts`, which is not in the image (binary, Vite assets and SSR bundle only).

Jobs ran in parallel. The QA stack is `hexascore-qa-esp` on :3110, with DB `hexascore_qa_esp` and `APP_URL=http://localhost:3110`. A throwaway Postgres 18.6 (`hx-esp-pg`, :5432) stood in for the CI service container.

| CI job | Result | Evidence |
| --- | --- | --- |
| lint | PASS except commitlint | `golangci-lint run --build-tags integration ./...`: 0 issues. `fmt --diff`: clean. `go vet`: ok. `scripts/openapi-check.sh`: no diff. `openapi:lint`: valid. Biome: ok. tsc: ok. markdownlint: 0 issues. `commitlint --from origin/main --to 7a5bf01`: 3 commits fail (67da7ed footer > 100, "E2E_DB_NAME" subject-case, "record seven reviews" footer > 100). Orchestrator decision: a single squashed commit on `release/v0.12.0-r2`. Message `chore(release): v0.12.0` + body tested with `git commit-tree 7a5bf01^{tree} -p origin/main`, then `commitlint --from origin/main --to <sha>`: rc 0. PR title via stdin: rc 0 |
| tests | PASS | `go test -count=1 -tags integration -parallel 4 -timeout 30m ./...`: all packages ok. vitest first run: 4 "Failed to start forks worker" errors (worker start timeout under the parallel load, 0 test failures). Rerun on a quiet host: 21 files / 74 tests. `migration-policy.test.ts`: 0 fail. `sdks/ts` typecheck + test: 0 fail |
| migrations | PASS | migrate, `migrate:rollback --step 1`, migrate on a fresh Postgres 18.6 |
| deps | PASS | `bun audit`: no vulnerabilities. osv-scanner v2.7.0 (container): first run reported GO-2026-5932 (x/crypto 0.58.0, no fix) because the image's Go 1.27.1 < go.mod 1.27.3 skipped call analysis. Rerun with `GOTOOLCHAIN=auto`: 0 vulnerabilities, like the CI job on PR #59 |
| vuln | PASS | `go tool govulncheck ./...`: affected by 0 vulnerabilities (1 in a required module, not called) |
| build | PASS | `docker build`: ok. `scripts/infra/check-image.sh`: ok. `/up` smoke without DB: `"ssr":"ok"` |
| openapi | PASS | `openapi-check.sh --image`: no diff. Redocly: valid. CI "App client matches codegen" (`codegen:client -prefix /app/v1` from the image vs `packages/app-client/src/api.ts`): no diff, app-client typecheck/test 2 pass. Schemathesis: api-v1 PASS (1500 generated, 1500 passed, 16 skipped; coverage, fuzzing and stateful phases green; warnings only). app-v1 and k6 were still running at the owner stop (13:58Z), so they are NOT RUN to completion |
| e2e | PASS (1 load flake), finished before the owner stop | full `bun run e2e` (6 projects): 695 passed, 1 failed in 27.7m. The failure: search.spec:41 [mobile-en] got 500 on `' OR 1=1 --`, and the app log says `search timeout` (1.86 s) while the Go integration suite and vitest loaded the host. Rerun of search.spec on a quiet host: 24/24. csp.spec and passkeys.spec pass in all projects, and so does auth.spec |
| lighthouse | NOT RUN | stopped by the owner order before it started; author run on FIX-csp-inline-style: best-practices 1.0 on all 5 URLs |
| mobile | PASS | `apps/mobile` lint, typecheck, test |
| setup-windows | not run | the Windows job runs `setups/setup-windows.ps1 -NoServices` + `go test ./...`; the Go tests are covered above |

Merged for this gate (prio 0): REV-FIX-e2e-release-0.12 (7500d45), REV-FIX-csp-inline-style (f3846ee), REV-FIX-covers-route CORRIGIDO (fd08b26), REV-FIX-passkeys-register-e2e (07c4129), REV-FIX-app-client-regen (7a5bf01).

Covers route against real RustFS: 200 with immutable cache, 404 for a missing key, traversal returns 404, and the demo SVGs win. The S3 driver sends no ETag, so there is no 304 (details in reports/FIX-covers-route.md).

Notes:

- `qa/schemathesis/ci.sh` does not run on Docker Desktop: the Windows-side `curl` cannot reach a `--network host` container. It was run with the same seed SQL; then the app was published on :3114 and `run.sh` and k6 ran against it (scratchpad `st.sh`).
- A search timeout surfaces as a 500. That is existing behaviour, but a hostile query under load should degrade without a server error. Worth a prio 2 row.

## Owner stop (13:58Z): release without e2e

| Item | Result |
| --- | --- |
| vitest | PASS (21 files / 74 tests on a quiet rerun; the first parallel run had 4 worker start timeouts and 0 test failures) |
| Go integration tests | PASS (`go test -tags integration -parallel 4 ./...`, all packages ok) |
| Schemathesis | PARTIAL: api-v1 PASS; app-v1 and k6 NOT RUN to completion |
| Lighthouse | NOT RUN |
| commitlint on the squash message | PASS (`chore(release): v0.12.0` single commit off origin/main, rc 0; PR title rc 0) |

The QA app was stopped with `stop app`. The full e2e run had already finished: 695/696; the one failure was a search timeout under load, and search.spec passed 24/24 on rerun.
