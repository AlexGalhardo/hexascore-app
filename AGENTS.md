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
- Follow the Karpathy guidelines (think first, simplicity, surgical changes, verifiable goals), `ponytail` (YAGNI, minimal solution) and `graphify` (map the code before broad searches). Vendored skills live in `.claude/skills/`.
- Validate every schema and piece of data with zod wherever possible (env, actions, forms, API params/responses, webhooks, external responses, jsonb reads); infer types with `z.infer`.
- Env vars are validated with zod in `apps/web/lib/env.ts` before build and server start; failures print every bad variable explicitly.
- Handle errors and fallbacks everywhere: timeouts on external calls, `notFound()`/404, `error.tsx`/500, RFC 9457 problems with 402/404/429/500 in the API, graceful degradation of optional services.
- Security rules in `docs/architecture.md` are mandatory.
- All reference documentation lives in `docs/` (architecture, data model, API, frontend, design system, payments, ingestion, security, testing, process). Update the matching doc in the same change.
- Pre-commit runs Biome and `tsc --noEmit`; never bypass hooks.
- When a workflow repeats or is easy to get wrong, capture it as a project skill in `.claude/skills/<name>/SKILL.md` and list it here.
- Project skills:
  - `hexascore-billing` — Stripe/AbacatePay webhooks locally, idempotency rules.
  - `hexascore-catalog` — catalog repository/cache layers, Next 16 caching without `cacheComponents`, PGlite tests, dev seed.
- Record notable changes in `CHANGELOG.md` (Keep a Changelog).
