# FIX-qa-port-isolation Parallel QA stacks without displacing each other
status: concluida
branch: agent/senior/FIX-qa-port-isolation @ debd454
files: docker-compose.yml (app-init/app only), lighthouserc.json -> lighthouserc.cjs
acceptance:
- default `docker compose --profile app config` -> port 3000, postgres DATABASE_URL, unchanged (CI uses qa-up.sh with COMPOSE_PROJECT_NAME, no new vars)
- two stacks up together, both `(healthy)`: `hexascore-qa-sb1-app-1` hexascore-app:FIX-e2e-nav-timeouts on 127.0.0.1:3101 and `hexascore-qa-sb2-app-1` hexascore-app:main on :3102; /up 200 on both; postgres/rustfs untouched
- `E2E_BASE_URL=http://127.0.0.1:3101 bunx playwright test -c qa/e2e/playwright.config.ts qa/e2e/specs/smoke.spec.ts --project=desktop-en` -> 1 passed
- `E2E_BASE_URL=http://127.0.0.1:3101/ node -e 'require("./lighthouserc.cjs")'` -> urls start with http://127.0.0.1:3101 (default unchanged; lhci 0.15.1 loads lighthouserc.cjs first, checked in @lhci/utils src)
validation: no Go/TS source touched; biome ci does not cover lighthouserc; lhci not run (bunx lhci breaks on this host)
design: Playwright already read E2E_BASE_URL; lighthouse now reads the same variable. Isolation = own compose project (-p), own host port (QA_PORT), own database (QA_DATABASE_URL, via host.docker.internal:5433 since another project cannot resolve "postgres"). Needs --no-deps so postgres/rustfs are not duplicated.
commands (for AGENTS.md "Local QA", parallel stacks), per worker with tag T and free port P (3100+):
  E=C:/dev/hexascore-app/hexascore-vukem/.env   # never copy .env into a worktree
  db=hexascore_qa_T   # T: letters/digits/_
  docker exec hexascore-postgres-1 createdb -U hexascore $db
  export E2E_IMAGE_TAG=T QA_PORT=P QA_DATABASE_URL="postgres://hexascore:hexascore@host.docker.internal:5433/$db?sslmode=disable"
  docker compose --env-file $E -p hexascore-qa-T --profile app build app-init
  docker compose --env-file $E -p hexascore-qa-T --profile app run --rm --no-deps app-init
  docker compose --env-file $E -p hexascore-qa-T --profile app up -d --wait --no-deps --no-build app
  E2E_BASE_URL=http://127.0.0.1:P bun run e2e      # lighthouse: same variable
  stop: docker compose --env-file $E -p hexascore-qa-T --profile app stop app   (never `down` on project hexascore)
  drop db: docker exec hexascore-postgres-1 dropdb -U hexascore $db
flag: app env now carries FLAG_COMMUNITY_SUBMISSIONS=${FLAG_COMMUNITY_SUBMISSIONS:-false} (orchestrator request, for P4.A.3.1b); open with FLAG_COMMUNITY_SUBMISSIONS=true before up.
changelog: Changed: QA compose app stack takes QA_PORT and QA_DATABASE_URL so parallel stacks (own -p project) coexist; Lighthouse base URL follows E2E_BASE_URL (lighthouserc.cjs).
roadmap: no box
risks / follow-ups: --env-file supplies APP_KEY from the main checkout .env (no copy into the worktree; I only exported the key in my shell). lighthouserc.json became lighthouserc.cjs; on this host @lhci/cli@0.15.1 does not install via bunx, use bunx lighthouse@12.6.1 per URL (E2E_BASE_URL still applies to the cjs). stale refs to lighthouserc.json in AGENTS.md:71, docs/sdd/09-web.md:238, docs/sdd/11-infra-deploy.md:166 (specialist: rename). passkeys.spec default origin is localhost:3000 when E2E_BASE_URL is unset (pre-existing). Stacks sb1 (3101) and sb2 (3102) left running for my r2 proof.

## Review (especialista, 2026-10-31)
verdict: CORRIGIDO_E_APROVADO
re-run: `docker compose --profile app config` (default) -> published 3000, DATABASE_URL `postgres:5432/hexascore` unchanged; `node -e 'require("./lighthouserc.cjs")'` -> `http://127.0.0.1:3000/`; used the documented commands for my own stack `hexascore-qa-esp` on :3110 (db `hexascore_qa_esp`, `--no-deps`), healthy next to sb1/sb2 and `hexascore-app-1` on :3000
checked: postgres/rustfs untouched, CI path (`qa-up.sh`) default unchanged, `extra_hosts host-gateway` works on Docker 20.10+ (CI ubuntu), budgets in lighthouserc.cjs identical to the JSON, no secret copied
fixes: 1. stale `lighthouserc.json` references renamed in AGENTS.md (commands table), SDD 11 §CI and `qa/lighthouse/run.sh` comment (SDD 09 has none) 2. AGENTS.md "Local QA" gains the parallel-stack recipe from this report
gate: no Go/TS source; markdownlint on the two docs clean (the run lists pre-existing SDD 03 MD009 only)
PRONTO_PARA_MERGE agent/especialista/FIX-qa-port-isolation-fix (merged d99e757)
r2 commits (454360e `E2E_DB_NAME` for seedNotification, debd454 `FLAG_COMMUNITY_SUBMISSIONS` default false) reviewed and merged at 4c28b1e; AGENTS.md recipe names `E2E_APP_CONTAINER` and `E2E_DB_NAME`.
