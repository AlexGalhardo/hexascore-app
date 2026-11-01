# PERF-compute-similar
status: parcial (code done; NFR-08-3 NOT demonstrated)
branch: agent/senior/P6.D.1.2 @ 7673421 (commits b88ab0d + 7673421 on top of 6febd4f; worktree wt/senior-P6.D.1.1)
files: app/recommend/similar/job.go, app/recommend/similar/bench_test.go, docs/sdd/08-recommendation.md
changes:
- Incremental by feature_hash moved before candidate generation: only items whose hash changed pay for holders, candidate reload, Filter and cooccur (before, every item did all of it and only the write was skipped).
- Cooccur: one unnest-pairs query per chunk instead of one query per item (6.1 s of 18.6 s at 5k with an empty lists table, measured with temporary timers).
- Write: one multi-row INSERT .. ON CONFLICT per chunk instead of one statement per row.
- holders: no join to items (candidates are validated by the candidate load, which already filters type and visibility); LATERAL index probe per feature. Found by EXPLAIN: on a freshly bulk-loaded table without statistics the planner estimated items at 1 row and chose a nested loop over the unnest (quadratic).
- Candidate feature extraction cached per chunk.
- Bench: ANALYZE after seeding (autovacuum does this in production; the seed's own join on an expression was quadratic without it, the 50k seed alone took about 30 min).
- SDD 08 section 5: user_recommendations.user_id is bigint (matches migration 20261031190000); SDD 03 section 5 still says text and should be corrected by its owner.
measurements (BenchmarkComputeSimilar, -benchtime 1x, shared Windows host at 85-90 % CPU from other workers, so noisy):
| items | before | after |
| --- | --- | --- |
| 5,000 | 14.0 s | 4.9 s |
| 20,000 | not run (killed after 188 s) | 36.9 s |
| 50,000 | 642 s | 212 s (other runs of intermediate code under heavier load: 258 s, 382 s) |
extrapolation to 500k: exponent from 5k to 50k is 1.63, which gives about 2.5 h, so the 60 min target is NOT shown. From the 20k point (1.85 ms/item) a linear run would take about 15 min. The residual superlinearity is not explained: DB time sampled in pg_stat_activity is small and the Go process is mostly waiting, so the contended host is a prime suspect, but I did not get a quiet-machine run or a per-phase profile at 50k.
acceptance:
- `go test -tags integration ./app/recommend/similar` -> ok (TestSimilarJob, TestSimilarSkipsUnchanged, TestApiVariantUsesAllowlistedSources unchanged and green)
- `golangci-lint run --build-tags integration ./app/recommend/similar/...` -> 0 issues
risks / follow-ups:
- Rerun the benchmark at 50k and 500k on a quiet machine with per-phase timers (holders / candidate load / Filter / write); if still superlinear, move candidate generation into one set-based SQL per chunk (self-join of item_entities on rare features) or cap candidates per item.
- Candidate set per chunk saturates at about 65k ids once n > 5k; Filter/Sim over those (26 % of CPU at 20k) and Scan (19 %) are the CPU hot spots.
- Weights drift for unchanged items is unchanged behaviour (hash covers F_i only), now cheaper to hit.
changelog: Changed: ComputeSimilar batches co-occurrence and writes per chunk and skips unchanged items before candidate generation.

## Review (especialista, 2026-10-31)

verdict: APROVADO
re-run: on 860b48b: `go test -tags integration ./app/... ./database/... ./routes/... ./bootstrap/...` -> ok; `./tests -run TestHome|ForYou|Inventory|Recommend|Trending|Similar|Security_A01` -> ok after the inventory fix; tsc ok
checked: PERF report; SQL placeholders; incremental skip by feature_hash
decision: no new follow-up row, PERF-compute-similar-500k (prio 2, todo) already covers the quiet-machine 500k run and set-based candidates; NFR-08-3 stays unproven until then
gate on main after merge: see batch gate in handoff/especialista.md (scripts/ci/local.sh on 127eeeb)

PRONTO_PARA_MERGE agent/senior/P6.D.1.2 @ 7673421
