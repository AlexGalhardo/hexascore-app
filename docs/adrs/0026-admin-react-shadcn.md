# ADR 0026: The admin panel moves to React with shadcn-style components over Base UI

Status: accepted (2026-11-03, owner decision in the admin redesign interview). Refines the "No shadcn, no other component kit" line of `.agents/rules/frontend.md`. Related: D-22 (Base UI), ADR 0020 (TanStack Query), SDD 03 section 6.3 (admin panel).

## Context

The `/admin` panel runs on the vukem `vadmin` package: Go templates with htmx, policies, audit and masking of sensitive columns. Its markup is unstyled HTML (owner screenshot 2026-11-03). The owner wants a modern dashboard in the Vercel style, with light and dark themes, built with shadcn/ui, Inertia and React Query.

shadcn/ui is a set of copied component sources, not a runtime kit. Its newest variant sits on Base UI, which the project already approved (D-22, `@base-ui/react`).

## Decision

- The admin UI is rewritten as Inertia pages under `resources/js/pages/Admin/**` with components in `resources/js/components/admin/**`.
- Components follow the shadcn/ui source style (CVA variants, `cn()` helper) over `@base-ui/react` primitives. No Radix.
- New runtime dependencies, exact versions, approved by the owner: `class-variance-authority`, `clsx`, `tailwind-merge`, `lucide-react` (admin only; tree-shaken into the admin chunk).
- Client data uses TanStack Query (ADR 0020) wrapped with `withQueryClient`.
- Security stays on the server: the admin gate, 2FA, CSRF, policies, audit and column masking are unchanged. React renders what the server allows and posts to the same endpoints.
- The public site keeps "no other component kit". The exception covers the admin only.

## Alternatives considered

- **Restyle the Go templates.** Lower risk and no new libraries, but it does not give the shadcn, Inertia and React Query stack the owner asked for.
- **Hybrid (React only for the dashboard and ingestion).** Two stacks to maintain; rejected by the owner in favour of the full React admin.

## Consequences

- Positive: one UI stack for the whole product, accessible Base UI behaviour, consistent theming from `packages/tokens`.
- Negative: the framework resource screens (users, score bursts, rating flags, score actions, formula versions, reports, contact messages, items, job runs, ingest failures, dq results, feature flags, audit log) cannot be read from the panel's own pages until `vadmin` exposes its data as JSON (VF-083). The app serves its own screens at `/admin/r/<slug>` over an allow-list of columns (`app/admin/resources_react.go`, so a password hash or 2FA secret is never selected). The framework panel is no longer mounted (only its guard is used), so no classic page remains: flags are set and reset by `app/admin/resources_write.go`, and contact messages are marked handled there; both write the same `admin_audit_logs` rows. Banning stays on `/admin/users/{id}/ban`. Deleting records and editing users beyond a ban are not offered (the old policies denied them or limited them to the ban fields). The old URLs redirect to the React screens.
- The admin chunk must stay lazy so public page budgets (Q-A20) are untouched.
- Rollback: revert the commit that unmounted the framework panel; the templates and resource definitions are in git history.
