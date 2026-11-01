# IMPL-web-v1-catalog v1 catalogue listing (Games, Movies, TV Series, Books)
status: parcial (numbered pagination and the desktop category columns wait for FIX-catalog-numbered-pagination)
branch: agent/senior/IMPL-web-v1-catalog @ bec4e3e (stacked on IMPL-web-v1-search-palette)
files: resources/js/pages/Catalog/Index.tsx, resources/js/components/catalog/{Filters,ItemCard}.tsx, resources/js/components/ui/{ScoreBadge,ScoreBadge.test}.tsx, resources/css/app.css (h1 focus target), resources/lang/*.json (catalog.filters is now plural, catalog.itemsFoundLabel)
acceptance:
- `bun run lint` → exit 0; `bun run typecheck` → exit 0; `bun run test` → 31 files, 123 tests passed
- ScoreBadge.test.tsx: bands 5.0/7.5, one decimal per locale ("8,3" in pt-BR), dash without score, activeCount
validation: browser ok on :3140 /games desktop 1440 + mobile 390, light + dark, mobile filters open; shots reports/shots/impl-web/catalog-*.png vs web-desktop__catalog-index__default.png and web-mobile__catalog-index__filters-open.png
changelog: Changed: catalogue listings in the v1 look (title alone with the result count, ruled filter row, Reset as a button, score badges with the band colour)
roadmap: no box
risks / follow-ups:
- Still InfiniteScroll: the controller pages with an opaque cursor and sends no page number or last page, so "previous / 1 2 3 … N / next" and the rank column cannot be built (FIX-catalog-numbered-pagination, senior).
- Desktop frame shows six category averages and the studio per row; listing cards carry neither (same FIX row). Rows use the mobile layout on every width for now.
- `catalog.filters` changed from a string to a plural object; no other user found (grep over .go/.html/.ts/.tsx).
- ScoreBadge is the shared badge for home, item and search rows.

## Review (especialista, 2026-10-31)

verdict: APROVADO
re-run: batch branch agent/especialista/IMPL-web-batch (shell, search-palette, catalog, item, home merged on main 54105f9; lang conflicts were sibling additions, union-resolved and JSON-validated): `bun run lint` ok, `typecheck` ok, `test` 33 files / 129 ok; mobile catalogues recopied (catalog.filters is now plural, unused on mobile), mobile test 47 / 292 ok; `go build ./...` ok; `go test ./tests/i18n` ok\nchecked: i18n parity (3 locales, no hardcoded copy), no dangerouslySetInnerHTML or inline scripts, suggest fetch encodes params and drops stale answers, localStorage guarded, trailers on youtube-nocookie behind consent, scope, tests\nnot re-checked in a browser by the reviewer (authors' frames in the reports)\ngate on main after merge: bun lint/typecheck/test ok (full local.sh at the end of the batch)\nPRONTO_PARA_MERGE agent/senior/IMPL-web-v1-catalog
