# FIX-search-prod-slow search over the 800 ms budget at production size
status: concluida
branch: agent/senior/FIX-search-prod-slow (cut from main 2600b4a)
files: internal/search/rank.go, internal/search/searcher.go, internal/search/scale_test.go
reproduction: scratch DB search_scale (migrated, 160k items, 288k translations, 53k alternative titles, hash-made words with a Zipf skew; no scale seed existed, SQL in the test). `EXPLAIN ANALYZE` of the grouped site query: 480 ms for a rare word, 1.5 to 2.5 s for common ones (statement timeout 800 ms), so the 503 "search timeout".
root causes (3):
1. Planner prices `<%`/`%` as cheap, so it seq-scanned item_alt_titles (word_similarity on every row, 160 ms per 53k rows) instead of its GIN index.
2. The trigram branches at theta 0.2 return tens of thousands of weak candidates for a short or common word (54k of 288k rows for one 4-letter word); every one went through sort, aggregation, items join and a display-title lateral lookup before the top 8 per type was cut.
3. The display title lookup (LATERAL with sort) ran per candidate instead of per returned row.
fix:
- searcher.go: `enable_seqscan = off` in the search transaction (set_config, local), documented.
- rank.go: each trigram branch keeps its 500 most similar rows (`trigramCap`); display title joined after the page is cut; sources check kept as EXISTS so result sets stay identical. Totals of a query that hits the cap are lower bounds (still far above the 200 reachable results).
acceptance:
- before/after, 40 words on search_scale (psql, ms): max 2499 -> 352, p50 about 100.
- `SEARCH_SCALE_ITEMS=160000 go test -tags integration ./internal/search -run TestSearchScaleP95 -v` -> PASS, 160000 items: p50 122 ms, p95 347 ms, max 356 ms (the first run with only the seqscan setting failed with ErrTimeout; the unfixed code takes 0.5 to 2.5 s per query).
- `go test -tags integration ./internal/search` -> ok (golden ranking tests unchanged); `go test -tags integration ./tests -run Search` -> ok.
validation: go build ok, gofmt ok; golangci-lint not run (8.2).
changelog: Fixed: /search no longer times out on a production-size catalogue (bounded trigram candidates, GIN index forced for alt titles, display title looked up after paging).
roadmap: no box
risks / follow-ups: the synthetic seed is skewed, real titles differ; worst case still about 350 ms locally for a word that trigram-matches 54k titles (heap recheck of the GIN candidates), so p95 on the real catalogue should be re-measured on staging with the owner's data. Cold cache on Railway could add to it. Totals shown for very broad queries are now capped. If production still times out, next step is a GiST word-distance index or raising theta for 3-4 letter queries (an SDD 07 decision, not made here).

## Review (especialista, 2026-10-31)
verdict: CORRIGIDO_E_APROVADO
re-run: `go test -tags integration ./internal/search` ok; `./tests -run Search` ok; `SEARCH_SCALE_ITEMS=160000 go test -tags integration ./internal/search -run TestSearchScaleP95 -v` PASS: p50 94 ms, p95 345 ms, max 519 ms (reproduces the author's numbers)
checked: SDD 07 §6/§8 and NFR-07-1; SQL still fully parameterised (the cap is a code constant); ranking goldens unchanged; display title now joined per returned row with the same locale order
decisions:
1. `enable_seqscan = off` is kept: it is set with `set_config(..., true)`, so it lives only in the search transaction (the same one that carries the threshold and statement_timeout), and it is documented in code and now in SDD 07. A planner-level fix (statistics, index shape, GiST word distance) is better long term but needs measurements on the real catalogue; recorded as the replacement path in SDD 07.
2. Trigram cap 500 per branch: `total` and the per-type counts of grouped mode (`(N)` next to each section heading, `total` in the JSON) become lower bounds only for broad queries. Pagination still reaches the 200-result ceiling. No copy promises an exact count today; the SDD now says totals are lower bounds. Follow-up if the owner wants it: show "N+" when a branch hits the cap.
3. Residual risk: p95 345 ms (max 519 ms) on the synthetic 160k seed is above NFR-07-1 (300 ms at 20 rps); cold cache on Railway may add more. Re-measure with k6 on the real catalogue after the release.
fixes: 1. SDD 07 records the cap, lower-bound totals, the transaction-local seqscan setting and the measured p95 (97d54e3) 2. golines on the scale test comment (1659419); branch agent/especialista/FIX-search-prod-slow-fix
gate on main after merge (1c95323): go build ok, golangci-lint (integration) 0 issues, bun lint/typecheck ok, vitest 27 files / 96 tests, go test -tags integration ./... exit 0
PRONTO_PARA_MERGE agent/especialista/FIX-search-prod-slow-fix (merged 1c95323)
