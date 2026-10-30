# ADR 0009: Web and mobile share an OpenAPI contract, generated clients and design tokens, not code

Status: accepted (2026-10-28). Mobile styling is amended by ADR 0018 (NativeWind); the sharing rules below are unchanged. Decisions: owner requirement 3; brainstorming D-04, D-12, A-01.

## Context

Requirement 3 asks for a React Native app (Android first) with the same visual style as the web and the same API and logic (login, CRUDs). The web is Inertia + React rendered by a Go server (ADR 0004); the mobile app is a separate React Native bundle talking HTTP. Business logic lives in Go. vukem generates OpenAPI 3.1 from typed handlers and can emit TypeScript types and clients (`https://vukem.dev/docs/rest-apis/openapi` "OpenAPI", "Types for clients", "Client generation"; `codegen:types`, `codegen:client`).

## Decision

Share three things, nothing else:

1. **The API contract**: a first-party API (`/app/v1/...`, access tokens with abilities) documented by OpenAPI 3.1 generated from the Go handlers. The mobile app uses a TypeScript client generated from that document in CI; a contract change that breaks the client fails the build. The web app (Inertia) does not use this API for page data; it receives props from the server. Shared validation rules are expressed once in Go `z` schemas and exported as JSON Schema for client-side pre-validation where useful.
2. **Design tokens**: colours, spacing, typography scale, radii, category colours and icon names in one TypeScript module (`packages/tokens`, published as a workspace package) consumed by the web (Tailwind config) and by React Native (StyleSheet). The star widget is implemented twice (DOM and RN) against the same token and behaviour spec in SDD 09/10.
3. **i18n message catalogues**: the same JSON files for the three locales, served to the web by the server and bundled in the app.

Not shared: React components (DOM vs RN), routing, state management, server logic.

## Alternatives considered

- **React Native Web / universal components (Expo + Tamagui/NativeWind).** One component tree for web and app; conflicts with Inertia SSR and vukem's Vite setup; heavy. Rejected.
- **Mobile app as a WebView wrapper.** Fastest, but poor UX and Play Store policy risk. Rejected.
- **Mobile app consuming the paid public API.** Mixes customer quotas with first-party traffic and exposes write operations to the paid surface. Rejected (A-01 separates them).

## Consequences

- Positive: least coupling; the server stays the single source of logic; contract drift is caught at build time.
- Negative: duplicated UI components; two widget implementations to keep in sync (tokens and a shared behaviour spec mitigate).
- Obligations: SDD 05 defines `/app/v1` and the generation pipeline; SDD 09 and 10 define tokens and the widget specs; CI publishes the generated client as an internal package.
