# Hexascore — Guide for AI agents

Media database (games, movies, TV series, books) with 6-category 0–100 user scores and a paid public REST API.
Codename `hexascore`; final brand name is pending.

**Status: planning.** No application code yet. The single source of truth is [`PLAN.md`](PLAN.md): decisions, architecture, data model, API, milestones. Read it before any work and keep its checkboxes up to date.

## Rules

- Communicate with the owner in pt-BR; code, identifiers and commit messages in English.
- Conventional Commits (commitlint). Changes go through PR to `dev`, then `dev` → `main`.
- Only stable, pinned dependency versions (no `^`, no beta/rc/canary). Ask before adding a dependency.
- Bun is the package manager. Python code lives in `webscrappings/` and uses uv.
- Local setup: `setups/` scripts (Unix/Windows × Docker/PGlite, see `setups/README.md`); Docker stack in `infra/`. Before a PR, `bun run lint:ci`, `bun run typecheck`, `bun run test` and `bun run build` must pass (required CI checks).
- Follow the Karpathy guidelines (think first, simplicity, surgical changes, verifiable goals), `ponytail` (YAGNI, minimal solution) and `graphify` (map the code before broad searches). Vendored skills live in `.claude/skills/`; always-on rules in `.claude/rules/` (karpathy, ponytail, graphify).
- Validate every schema and piece of data with zod wherever possible (env, actions, forms, API params/responses, webhooks, external responses, jsonb reads); infer types with `z.infer`.
- Data fetching: Server Components first; client-side fetches go through TanStack Query (`useQuery`/`useInfiniteQuery`), never `useEffect` + `fetch` (see `docs/frontend.md`).
- Listing pattern (every item listing, now and future): split by category (one `CategoryLabel` sub-header per type, empty groups hidden), `CoverThumb` left of every item title (queries return `coverUrl`/`coverThumbUrl`), and every category name rendered with `CategoryLabel` (icon in the section accent). See `docs/frontend.md` → "Padrão de listagens".
- Env vars are validated with zod in `apps/web/lib/env.ts` before build and server start; failures print every bad variable explicitly.
- Every modal is centred on screen: Base UI `Dialog`/`AlertDialog` use `DIALOG_BACKDROP` and `dialogPopup()` from `apps/web/components/ui/dialog.ts` (enforced by `components/ui/dialog.test.ts`); native `<dialog>` uses `showModal()` with `m-auto`.
- Handle errors and fallbacks everywhere: timeouts on external calls, `notFound()`/404, `error.tsx`/500, RFC 9457 problems with 402/404/429/500 in the API, graceful degradation of optional services.
- Security rules in `docs/architecture.md` are mandatory.
- All reference documentation lives in `docs/` (architecture, data model, API, frontend, design system, payments, ingestion, security, testing, process). Update the matching doc in the same change.
- Pre-commit runs Biome and `tsc --noEmit`; pre-push runs `bun run ci:local` (`scripts/ci-local.sh`: lint, typecheck, test, openapi check, build, and ruff/pytest when `webscrappings/` changed) as the stand-in for GitHub Actions, whose workflows are kept in `.github/workflows/` but disabled on GitHub (`gh workflow enable <name>` re-enables). Never bypass hooks.
- When a workflow repeats or is easy to get wrong, capture it as a project skill in `.claude/skills/<name>/SKILL.md` and list it here.
- Project skills:
  - `hexascore-billing` — Stripe/AbacatePay webhooks locally, idempotency rules.
  - `hexascore-secure-endpoint` — checklist for actions, routes, webhooks and crons.
  - `hexascore-local-verification` — run the CI gates locally before done.
  - `hexascore-release` — cut or backfill releases (tags, GitHub releases, CHANGELOG, release-please manifest).
  - `hexascore-catalog` — catalog repository/cache layers, Next 16 caching without `cacheComponents`, PGlite tests, dev seed.
  - `hexascore-api` — public API v1: add an endpoint (zod DTO → OpenAPI registry → route via `apiRoute` → route test → regenerate `openapi.json` + SDKs).
  - `hexascore-qa` — QA layers: Playwright + axe E2E (`qa/e2e`) against `next start` or the preview, CSP/security headers, k6 / Schemathesis / OWASP ZAP API gates (`qa/*/run.sh`), scale seed, `qa.yml` gate.
- Record notable changes in `CHANGELOG.md` (Keep a Changelog).

## Project skills

- `.claude/skills/hexascore-ratings/SKILL.md` — rating write path (lock item row, recompute stats from `ratings` in the same transaction, threshold applied on read), merge, local E2E gotchas. Admin role: `bun run admin:grant <email>`.
- `.claude/skills/hexascore-auth/SKILL.md` — auth/account work: Better Auth server-only endpoints, Server Action result contract, BotID local setup, testing flows locally (verification link in the dev log, TOTP).
