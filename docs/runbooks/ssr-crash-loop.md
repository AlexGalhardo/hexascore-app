# R1: SSR crash loop

Alerts A5 (`alert=A5` on a public route) and A2 (`/up?checks=1` returns 503 with `ssr: fail`). Design: SDD 12 §7.8, SDD 09 §6.1. Private routes (`/admin`, `/profile`, `/me`, `/billing`, `/suggest`, `/suggestions`) skip SSR; a fallback there logs a plain WARN with `public=false` and is not an incident.

1. Railway log search: `alert=A5`, then `"inertia ssr fallback"` for the `type` and `component` of the failures.
2. Run `/app/app ssr:check` in the service (`railway ssh`; the image has no `vukem` CLI).
3. If the failures began with a deploy, the bundle is bad: roll back ([rollback](rollback.md)).
4. If the Node process runs out of memory, disable SSR on the heaviest logged-in pages with `inertia.DisableSSR`, then redeploy.
5. Pages still render on the client during the incident; the loss is SEO and first paint. Confirm the A5 lines stay flat for 15 minutes.
