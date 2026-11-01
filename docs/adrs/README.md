# Architecture decision records

Format: [MADR](https://adr.github.io/madr/) (title, status, context, decision, alternatives, consequences). One file per decision, numbered, never edited after acceptance except for the status line (`superseded by NNNN`). Decisions originate in [../brainstorming.md](../brainstorming.md); the ADR records the reasoning for agents that do not have that conversation.

| ADR | Title | Status |
| --- | --- | --- |
| [0001](0001-rewrite-on-vukem.md) | Rewrite HexaScore on the vukem Go framework | accepted |
| [0002](0002-railway-over-vercel.md) | Host on Railway instead of Vercel | accepted |
| [0003](0003-postgres-only.md) | PostgreSQL as the only data store; Redis and MongoDB require their own ADR | accepted |
| [0004](0004-inertia-react-with-ssr.md) | Inertia v3 + React with mandatory server-side rendering for public pages | accepted |
| [0005](0005-scoring-model.md) | Half-star input, published Bayesian aggregate, stars 0–10 display | accepted |
| [0006](0006-third-party-data-licensing.md) | Third-party data posture: Wikidata-only in the paid API, critic scores dropped | accepted |
| [0007](0007-ingestion-in-go.md) | Ingestion rewritten in Go on vukem queue and scheduler, single process | accepted |
| [0008](0008-auth-with-vukem.md) | Authentication with vukem `auth` + `social`; passkeys deferred | accepted |
| [0009](0009-web-mobile-sharing.md) | Web and mobile share an OpenAPI contract, generated clients and design tokens, not code | accepted |
| [0010](0010-community-edit-suggestions.md) | Community edits as moderated suggestions with audit trail | accepted |
| [0011](0011-search-postgres.md) | Search on PostgreSQL FTS + pg_trgm behind one interface | accepted |
| [0012](0012-books-work-and-editions.md) | Books modelled as work + editions | accepted |
| [0013](0013-versioning-releases-cutover.md) | Versioning from v0.1.0, vukem-style releases, big-bang cutover | accepted |
| [0014](0014-backups-rpo-rto.md) | Backups with RPO 24 h and RTO 1 h | accepted |
| [0015](0015-monetisation-v1.md) | Monetisation: paid read API via Stripe in v1, PIX in the second release | accepted |
| [0016](0016-bot-protection.md) | Bot and abuse protection without Vercel BotID | accepted |
| [0017](0017-passkeys.md) | Passkeys return through vukem `auth` (VF-001) | accepted |
| [0018](0018-nativewind.md) | NativeWind styles the Android app; tokens stay shared, components do not | accepted |
| [0019](0019-mobile-runtime-dependencies.md) | Mobile runtime dependencies for auth, cache and media | accepted |
| [0020](0020-web-tanstack-query.md) | TanStack Query for client-fetched data on the web | accepted |
| [0021](0021-social-feed.md) | Activity feed built from the write path, no fan-out until 10k followers | accepted |
