# FIX-ci-release-checks Schemathesis and Lighthouse red on release PR #65

status: concluida
branch: agent/senior/FIX-ci-release-checks @ 400438ce (86236e0d fix + docs commit; from main 0caee8d1; local, not pushed)
files: qa/schemathesis/{run.sh,ci.sh,schemathesis.toml}, lighthouserc.cjs, scripts/ci/{local.sh,schemathesis-run.test.ts,lighthouse-config.test.ts}, .github/workflows/ci.yml, vitest.config.ts, AGENTS.md
root cause (one change, two symptoms; not bind address, not ScrapeGuard): `CanonicalHost` (70fe6ce2, production only) 301s every Host other than APP_URL. The app binds `0.0.0.0` (log `addr 0.0.0.0:3000`) and stays up.

- Schemathesis called `host.docker.internal:3000`, got a 301 to `http://127.0.0.1:3000/...`, and followed it to its own container loopback. That is the "Connection refused" error. Reproduced: no Host header gave 301, the original Host gave 401.
- A second break hid behind it. The api-v1 golden moved to the `/v1` server form (116e4df0), so the paths are `/books/...` and the scan hit HTML 404s. The toml `^/(api|app)/v1/` status list also no longer matched api-v1, so validation 422s failed `positive_data_acceptance`.
- Lighthouse loaded `127.0.0.1:3000` while compose APP_URL is `localhost:3000`, so every page started with a 301. The `redirects` audit shows 602 ms of simulated time. Earlier "clean container" runs used `localhost` and never saw it.
- LCP element on all 5 URLs: the cookie banner `section.fixed p.text-sm`. It is SSR-rendered: the initial state comes from Inertia props, so it paints with the first paint and is not a hydration delay. Not changed.

fix (test and CI config only; CanonicalHost and production binding untouched):

- `run.sh`: rewrites the URL to host.docker.internal, keeps the path, and sends `Host: <original host:port>`.
- `run.sh` default base for api-v1 is `/api/v1`. `ci.sh` passes it too.
- toml regex `^/`.
- `lighthouserc.cjs` and `local.sh` default to `localhost`.
- On Docker Desktop, k6 runs in the app's network namespace.
- 2 bun tests, wired into `ci.yml`, `local.sh` and the vitest exclude list.

acceptance:

- `bun test scripts/ci/schemathesis-run.test.ts scripts/ci/lighthouse-config.test.ts scripts/ci/local.test.ts`: 7 pass. Both new tests were red before the fix.
- Schemathesis 4.29.0 against the main image (APP_ENV=production, APP_URL 127.0.0.1:3901, reached via host.docker.internal):
  - api-v1: exit 0, 2056 passed;
  - app-v1: exit 0, 2979 passed.
  - Before the fix: 20 Network errors, then 20 failed.
- Lighthouse in `lhci-runner:local`, repo `lighthouserc.cjs`, median of 3, main image QA stack:

| URL | LCP via 127.0.0.1 (old) | LCP via localhost (new) | script |
| --- | --- | --- | --- |
| / | 1734 | 1131 | 138.0 KB |
| /games | 2263 | 1511 | 141.4 KB |
| /games/adipiscing-consectetur | 1813 | 1209 | 150.0 KB |
| /books/consectetur-2004 | 1815 | 1210 | 150.0 KB |
| /people/ana-oliveira | 1811 | 1132 | 138.0 KB |

- `lhci assert`: exit 1 with the old base, exit 0 with the new one. Budgets not changed.
- `scripts/ci/local.sh --only build,openapi` on 86236e0d: exit 0, ALL GREEN. build PASS in 51 s; openapi PASS in 782 s (export diff, Redocly, codegen, Schemathesis on both docs, k6 30/30 checks, p95 37.6 ms).

validation: bun lint ok, typecheck ok, `bun run test` 58 files / 215 tests ok, markdownlint AGENTS.md ok, `bash -n` ok. No Go code touched.
changelog: Fixed: CI Schemathesis and Lighthouse reach the production-mode QA app on its APP_URL host (no CanonicalHost redirect), and the api-v1 scan uses the `/api/v1` base.
roadmap: no box
risks / follow-ups:

- CI LCP minus 602 ms gives about 1.5 s on items and 1.8-1.9 s on `/` and `/games`. `/games` is the tightest; the runner CPU is slower than this host.
- Playwright still defaults to `127.0.0.1` (`qa/e2e/playwright.config.ts`, `global-setup.ts`). The e2e job is off; switch it to `localhost` before enabling `RUN_E2E`.
- Item-page script is 149976 B against 153600 B, so 3.6 KB of headroom.
- PR #65 needs these 2 commits on the release branch, then a CI re-run.

## Review (especialista, 2026-11-01)

verdict: APROVADO
re-run: `bun test scripts/ci/schemathesis-run.test.ts scripts/ci/lighthouse-config.test.ts scripts/ci/local.test.ts` → 7 pass; `bash -n` on run.sh, ci.sh, local.sh → ok
checked:

- `run.sh`: loopback base is rewritten to `host.docker.internal` with the path kept and `Host: <original host:port>`; non-loopback bases pass through with no Host override (test covers all three).
- api-v1 base `/api/v1` matches the golden's `/v1` server form (31 relative paths); app-v1 paths all start with `/app/v1`.
- toml `^/` only moves the `positive_data_acceptance` expected-status list onto the same operation set as before (every op of both docs); no check is enabled or disabled.
- k6: Linux CI keeps `--network host` and `127.0.0.1:<port>` = APP_URL; Docker Desktop joins `container:<app>` and reaches the same `127.0.0.1:<port>`.
- Lighthouse: CI `qa-up.sh` uses compose `APP_URL=http://localhost:3000`, so the new `localhost` default matches it; port published on 127.0.0.1, Chrome resolves `localhost` to loopback.
- vitest `scripts/**/*.test.ts` include plus the two new excludes; ci.yml and local.sh `test-ts` run them under bun.
- Scope: test and CI config plus AGENTS.md; CanonicalHost and production binding untouched. No secrets, no SQL, no i18n.

gate on main after merge: see especialista handoff (`bun run test`, scripts/ci bun tests)
PRONTO_PARA_MERGE agent/senior/FIX-ci-release-checks @ 400438ce (merged f9b76390)
