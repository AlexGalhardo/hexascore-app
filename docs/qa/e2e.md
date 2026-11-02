# Playwright e2e

The suite lives in `qa/e2e` (config `qa/e2e/playwright.config.ts`, helpers in `qa/e2e/helpers`) and runs against the built container seeded with DemoCatalog. Projects are `desktop-<locale>` and `mobile-<locale>` for `en`, `pt-BR` and `es`.

## Run locally

Use your own QA stack (recipe in `AGENTS.md`, "Local QA"), then from the checkout:

```sh
E2E_BASE_URL=http://localhost:<port> \
E2E_APP_CONTAINER=hexascore-qa-<tag>-app-1 \
E2E_DB_NAME=hexascore_qa_<tag> \
bun run e2e --project=desktop-en --workers=2
```

Use `bun run e2e`, not `bunx playwright`: `bunx` fetches an unpinned Playwright that does not match the installed browsers. The base URL must use `localhost` (the image answers 301 to `APP_URL` for any other host). Signup, login and reset solve a proof of work in the browser, so a flow takes seconds and the per-test timeout is 120 s. Traces and screenshots of failures go to `qa/e2e/test-results/`.

## What gates what

| Trigger | e2e job | Projects |
| --- | --- | --- |
| push to `main`, ordinary pull request | skipped | none |
| push to `release/**`, pull request from a `release/*` branch | runs | `desktop-en` |
| `workflow_dispatch` | runs | `desktop-en` |
| repo variable `RUN_E2E=true` (any trigger) | runs | full matrix (3 locales, desktop and mobile) |

The job has a 25-minute timeout and uploads `qa/e2e/test-results/` and `playwright-report/` as the `playwright-traces` artifact when it fails (decision D-42). A product bug found by a spec is quarantined with `test.fixme` and a `FIX-` row on the board, never by deleting the spec.
