REV P1.G.2.4: CORRIGIDO_E_APROVADO (base merged 79642c4). Final branch: agent/especialista/P1.G.2.4-fix (55bb6b0).
Bug fixed: Catalog/Show linked country/language entities by slug (/countries/brazil) while the new route matches entities.code -> 404. EntityLink gains `code` (e.code), Show.tsx uses code ?? slug; regression assertion in TestInertiaPages_Entity_Show (red before, green after). CHANGELOG [Unreleased] Fixed.
Reviewed OK: SQL bound params (col is a constant), public-only counts/items, entity_stats >=5 threshold, offset clamped (ParseOffset), item_entities_entity_idx covers lookups, i18n 3 locales, X-Robots-Tag + Seo robots prop.
Re-run: go test -tags integration ./tests -run TestInertiaPages -> ok; golangci-lint -> 0 issues; go test ./... and -tags integration ./... -> ok; bun lint/typecheck clean, bun run test 57 passed.
curl | grep -c 'noindex,follow' NOT RUN: creating a local .env for `vukem serve` is blocked by a deny rule on .env files, and the meta tag needs the SSR build; header covered by AssertHeader, meta by Show.test.tsx.
Open (not fixed, pre-existing): code case is stored as-is (BR) and the route is case-sensitive; Highlight score scale noted by author.
