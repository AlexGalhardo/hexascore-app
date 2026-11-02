# ADR 0001: Rewrite HexaScore on the vukem Go framework

Status: accepted (2026-10-28). Decisions: owner requirement 1; brainstorming D-04, D-11.

## Context

The current application (`legacy/`, release 0.0.25) is Next.js 16 on Vercel with Neon Postgres, Upstash Redis, Vercel Blob and Python ingestion run by hand. Pain points found in the legacy audit: every page is dynamic because of the CSP nonce in `proxy.ts`, `unstable_cache` workarounds, Windows build crash, no CI in the cloud, manual ingestion against production, cost drivers tied to Vercel (Web Analytics, Pro plan), and a paid API that re-serves third-party data without a licence.

The owner also develops vukem (`github.com/AlexGalhardo/vukem`, v0.32.0 on 2026-10-25, BETA phase B.5, 352 test files, `TestSecurity_*` gates, OpenAPI 3.1 generation from typed handlers, Inertia v3 adapter, queue, scheduler, rate limiting, admin panel, Stripe contrib, Postgres driver). Using HexaScore as a real product on vukem is a product goal in itself: the framework gets a production consumer and HexaScore gets a stack the owner controls end to end.

## Decision

Rewrite the application in Go on vukem, in `hexascore-vukem/`, as a single Go service plus the Node SSR process required by ADR 0004. Use vukem's native subsystems first (router, `web.Ctx`, `z` schemas, ORM and migrations, auth, social, access tokens, admin, queue, scheduler, rate limiting, mail, cache, storage, i18n, OpenAPI, Inertia, Vite). A third-party Go dependency is added only when vukem has no equivalent and building it inside the server is unreasonable; each such case gets an ADR and, when the gap is a framework gap, an entry in `vukem-feedback.md`.

The Next.js tree is frozen on branch `legacy` (commit `97243cb`) and kept as the parity baseline and rollback target for 30 days after cutover (ADR 0013).

## Alternatives considered

- **Stay on Next.js and fix the pain points.** Cheaper short term; keeps the Vercel coupling, two runtimes (Node + Python) and does not serve the framework goal. Rejected by the owner.
- **Go with another framework (chi, echo, Fiber) or plain `net/http`.** Mature, but every subsystem (auth, admin, queue, OpenAPI, Inertia) would be assembled from third parties; vukem is `net/http`-native so the escape hatch exists anyway. Rejected.
- **Incremental strangler migration route by route.** Two auth systems and two databases during the transition; see ADR 0013. Rejected.

## Consequences

- Positive: one runtime for application logic; typed OpenAPI generated from handlers; security defaults with tests; cheaper hosting (ADR 0002); ingestion inside the app (ADR 0007).
- Negative: a full rewrite of about 26k lines of TypeScript and 10k of Python; React components are reused only partially (ADR 0004); vukem `auth`, `social` and Inertia are ALPHA/BETA, so breaking changes are expected before vukem v1 and must be absorbed by pinning versions and reading upgrade guides (`vukem/docs/upgrade/`).
- Obligations: `hexascore-vukem/go.mod` pins the vukem version; local development may use a `go.work` or `replace` to `../vukem`, never committed; every framework bug or DX gap is logged in `vukem-feedback.md` with the format required by the owner (id, date, severity, context, reproduction, workaround, status).
