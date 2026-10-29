# VF to ROADMAP microtask map (generated 2026-10-29 by estagiario, E-003)

Source: vukem-feedback.md (severity, title), ROADMAP.md (microtask = nearest preceding `- [ ] **P...**` line). Order inside a row = roadmap order.

| VF | severity | title | microtasks (earliest first) |
|---|---|---|---|
| VF-001 | high | No passkey / WebAuthn support in `auth` | none |
| VF-002 | medium | Inertia SSR falls back to client rendering silently by default | P0.F.2.1, P0.F.11.1, P1.A.4.2, P2.A.3.2, P2.C.3.3 |
| VF-003 | medium | No built-in bot/abuse challenge for auth and form routes | P1.C.5.2, P2.D.2.3 |
| VF-004 | low | `contrib/imaging` flagged as needing work in the framework's own review | P6.B.2.1 |
| VF-005 | medium | No keyset (cursor) pagination helper in the ORM | P0.F.3.1, P0.F.11.1, P1.B.2.2, P3.A.1.2, P4.B.1.2, P5.B.1.2 |
| VF-006 | low | Postgres mapping of `t.JSON` and `t.Enum` undocumented | P0.F.5.1, P0.F.11.1, P1.B.1.1, P4.A.1.1, P6.A.1.2 |
| VF-007 | low | No sitemap helper | P1.B.4.2 |
| VF-008 | medium | Admin panel has no custom row actions, bulk actions, assignment or diff view | P1.D.3.1, P1.F.3.2, P1.G.4.2, P4.B.1.1, P4.B.1.3, P6.B.3.3, P6.B.4.3 |
| VF-009 | low | Generated columns and the ORM on insert/update | P0.F.5.1, P0.F.11.1, P1.D.1.1 |
| VF-010 | low | Admin `BeforeSave` hook does not expose previous values | P1.C.3.1, P4.B.2.5 |
| VF-011 | low | No persistent in-app notifications | P1.F.3.1, P4.B.3.1 |
| VF-012 | low | No admin impersonation ("login as user") | P1.C.3.1 |
| VF-013 | low | Locale for mail sent from queued jobs | P1.C.5.1 |
| VF-014 | low | Envelope and problem key casing is fixed to snake_case | P0.F.7.1, P0.F.11.1, P1.H.2.2, P3.A.1.2 |
| VF-015 | medium | No host-based routing helper | P0.F.4.1, P1.H.2.2, P3.A.6.2 |
| VF-016 | low | Problem+json extension members undocumented | P0.F.7.1, P1.H.2.2, P3.A.6.2 |
| VF-017 | medium | No non-transactional migrations (`CREATE INDEX CONCURRENTLY`) | P0.F.5.2, P1.A.4.3, P1.E.1.1, P2.E.1.1 |
| VF-018 | medium | Queue priority and per-queue concurrency with the database driver inside `serve` | P0.F.6.1, P0.F.11.1, P1.B.1.2, P1.D.3.2, P1.E.1.2, P2.A.3.2, P6.A.1.3, P6.A.4.1 |
| VF-019 | medium | No documented API for app-level OpenTelemetry metrics | P2.C.2.2, P2.C.3.5, P2.C.5.1, P4.B.4.2, P6.B.4.5 |
| VF-020 | low | `httpclient` lacks per-host throttling and circuit breaking | P6.A.1.4 |
| VF-021 | low | No buffered counter primitive and no documented `HourlyAt` schedule helper | P0.F.6.1, P0.F.11.1, P1.D.3.2, P2.C.1.2, P2.C.2.2, P3.A.3.2, P6.C.1.2 |
| VF-022 | low | Request logger lacks privacy options (hashed user id, truncated IP, request id = trace id) | P2.C.1.1, P2.C.5.1 |
| VF-023 | low | Site documentation gaps found while writing the observability spec | P1.I.7.1, P2.C.5.1 |
| VF-024 | medium | `i18n.Middleware` has no URL-prefix locale source | P0.F.2.3, P0.F.11.1, P1.A.4.1 |
| VF-025 | medium | No cache hook for Inertia SSR output | P1.G.1.2, P1.I.3.1 |
| VF-026 | medium | Migration runner has no documented lock for concurrent starts | P0.F.5.2, P0.F.11.1, P1.A.4.3, P2.A.4.1, P2.E.1.1 |
| VF-027 | medium | `serve` and `build` gaps for the SSR sidecar and release metadata | P0.F.2.1, P0.F.2.2, P1.A.4.2, P2.A.3.1, P2.A.3.2, P2.B.2.3 |
| VF-028 | low | Root view data has no theme hook | P0.F.2.3, P1.A.4.1 |
| VF-029 | low | Export helpers only target `http.ResponseWriter` | P0.F.9.1, P0.F.11.1, P1.F.4.1, P3.A.4.1, P4.B.2.6, P6.D.3.1 |
| VF-030 | medium | No admin hardening options (mandatory 2FA, shorter sessions, IP allowlist) | none |
| VF-031 | medium | No re-encryption command after `APP_KEY` rotation | none |
