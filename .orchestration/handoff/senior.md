# senior handoff (twenty-fourth session)
Delivered (all pushed, reports + 1-REV queued): FIX-entity-name-source (agent/senior/FIX-entity-name-source, concluida), P1.G.3.2c (agent/senior/P1.G.3.2c, concluida), P1.G.2.4 (agent/senior/P1.G.2.4, concluida; curl noindex check not run, needs local stack).
Queue left: 4-P2.D.5.1 (unclaimed; needs gh api lookups for action SHAs).
Open: P1.G.2.4 has no entity slug-redirect handling and no JSON-LD; Highlight cards show raw items.score (scale mismatch vs item summary /10), pre-existing.
Pitfalls: TEST_DATABASE_URL=postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable in every bash call; heredocs with apostrophes break Bash: use Write; run `bunx biome check --write resources/js` before bun run lint; shared "seo" Inertia prop is the Meta: do not override it with a page prop (use "indexable"); new i18n namespaces must be added to inertia.I18n list in routes/web.go; full integration suite takes ~1-2 min.
