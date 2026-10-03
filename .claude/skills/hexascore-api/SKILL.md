---
name: hexascore-api
description: Add or change a Hexascore public API v1 endpoint (apps/web/app/api/v1, lib/api) and keep OpenAPI, docs and SDKs in sync. Use when touching apiRoute, lib/api/schemas.ts, lib/api/openapi.ts, an /api/v1 route, openapi.json or sdks/.
---

# Hexascore public API (v1)

Reference: `docs/api.md`. Code: `apps/web/lib/api/` (`handler.ts`, `schemas.ts`, `openapi.ts`, `quota.ts`, `plans.ts`, `problem.ts`), routes in `apps/web/app/api/v1/`, tests in `apps/web/lib/api/v1.test.ts`.

## Rules (never break)

1. Every route is `export const GET = apiRoute({ params?, query, response }, handler)`. Never hand-roll auth, quota, errors or headers.
2. The response zod schema is the contract: the handler returns `z.input<typeof schema>` and `apiRoute` parses it. snake_case fields, ISO dates, `null` instead of missing for nullable data.
3. Query schemas are `z.strictObject` (unknown params → 400) with `z.coerce` for numbers; `limit` must be bounded by `limits.maxPageSize` (use the `(limits, params) => schema` form).
4. Never expose `external_scores` (third-party scores are frontend-only by license), user ids, emails or individual ratings.
5. Expected misses throw `NotFoundError`; handler-only input errors throw `BadRequestError`. Anything else is a 500 that is refunded to the caller's quota.
6. Read the catalog through `lib/catalog/repository.ts` functions (`fn(db, …)`), passing `db` from `@/lib/db/client`.
7. Additive changes only inside `/v1` (see `/docs/changelog`); breaking changes need `/v2`.

## Add an endpoint

1. **DTO**: add/extend schemas in `lib/api/schemas.ts` (with `.meta({ description, examples })` written for developers) and a mapper from the repository DTO if needed.
2. **Registry**: in `lib/api/openapi.ts` register new component schemas (`components.add(schema, { id })`) and add the path with `operation({ id, tag, summary, description, query, schema, errors })`. `operationId` becomes the SDK function name — pick it carefully (`listX`, `getX`).
3. **Route**: `apps/web/app/api/v1/<path>/route.ts` via `apiRoute`. Static segments win over `[type]`.
4. **Test**: extend `lib/api/v1.test.ts` (it mocks `@/lib/db/client` with PGlite and imports the route dynamically): happy path parsed with the response schema, 400/404 cases, headers.
5. **Regenerate** (from the repo root) and commit the outputs:
   ```bash
   bun run openapi              # apps/web/public/openapi.json
   bun run sdk:generate         # sdks/ts/src/gen
   bun run sdk:generate:python  # sdks/python/hexascore (needs uv)
   ```
6. **Docs**: update `docs/api.md`, the `/docs` page if behaviour changed (e.g. errors, limits), `/docs/changelog` entries and `CHANGELOG.md`.
7. **Verify**: `bun run lint:ci && bun run typecheck && (cd apps/web && bun test lib/api) && bun run openapi:check`, then `bunx @redocly/cli@2 lint apps/web/public/openapi.json`.

## Gotchas

- `bun run openapi:check` is in the CI `lint` job: forgetting step 5 fails the PR.
- `public/openapi.json` is excluded from Biome on purpose (generated, 2-space JSON).
- Locally, real Redis cannot run `@upstash/ratelimit` scripts → the API returns 503 (fail closed). Use `NODE_ENV=test bun --bun next start` for the in-memory store (see `docs/testing.md`).
- Scalar (`components/api/api-reference.tsx`) scrolls the window to its hash routes when it mounts; on mixed pages mount it on demand.
- Exports other than components from a `"use client"` module become client references in server components — keep shared constants in plain modules.
