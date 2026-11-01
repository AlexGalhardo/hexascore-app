# FIX-ingest-qids-pushdown
status: concluida
branch: agent/senior/FIX-ingest-qids-pushdown @ 86e4777a (worktree wt/senior-fw-qids, from local main 02fb1b90)
files: `internal/ingest/{tmdb,igdb,trailers,openlibrary}/*.go`, `internal/ingest/fetch/fetch.go` (comment), `tests/ingest/qids_pushdown_test.go`
red: test first; `TEST_DATABASE_URL=... go test -tags integration ./tests/ingest -run TestConnectorTargetsPushQIDsIntoSQL` failed for tmdb and igdb (with qids [Q2 Q9] = [Q1 Q2], want [Q2]; "qids never reached the SQL; args = [[ <nil> false 0]]"); trailers failed on a fixture shape, fixed, then red the same way.
green: same command ok, plus `-run QIDs` ok; unit tests of tmdb, igdb, trailers, openlibrary ok; golangci-lint -j 2 on ./internal/ingest/...: 0 issues; the new test file adds no findings with the integration tag.
design: each targetsSQL gets `AND (coalesce(cardinality($n::text[]), 0) = 0 OR <qid column> = ANY($n::text[]))` with p.QIDs as the last arg (nil or empty = no restriction), so keyset order, cursor and limit stay as before. The test seeds items and source_records, checks that a spy executor saw the []string argument, and compares the targets with and without qids (and resumed after a cursor).
note: the Go filter in fetch.collect stays: steam, gog and wikipedia still have no SQL pushdown (out of this row), so removing it would break their qids runs. A comment there says so.
changelog: Changed (qids runs no longer scan every target of tmdb, igdb, trailers and openlibrary).
roadmap: no box (follow-up of P6.B review)

## Review (especialista, 2026-11-01)
verdict: APROVADO
re-run: `go test -tags integration -p 1 ./tests/ingest -run 'TestConnectorTargetsPushQIDsIntoSQL|QIDs|Trailers|TMDB|IGDB|OpenLibrary'` → ok; unit tests of tmdb, igdb, trailers, openlibrary → ok; `golangci-lint run --timeout 15m -j 2 --build-tags integration ./internal/ingest/... ./tests/ingest/...` → 0 issues after merge (the two rowserrcheck findings on the branch base were already fixed on main); `golangci-lint fmt --diff` → clean
trailers red observed by the reviewer: with main's trailers.go the subtest fails (`with qids [Q1 Q3] = [Q1 Q2 Q3]`, `qids never reached the SQL`, `after Q1 = [Q2 Q3]`), so the inferred red holds.
checked: sql bound (`$n::text[]`, nil and empty mean no restriction), keyset/cursor/limit unchanged and LIMIT now applies after the qids filter (a limited qids run no longer loses targets to the Go filter), scope ok, the Go filter in fetch.collect stays for steam, gog and wikipedia
gate on main after merge: go build ok; lint ok; targeted ingest tests ok; full integration suite left to the batch gate (host short on memory)
PRONTO_PARA_MERGE agent/senior/FIX-ingest-qids-pushdown (merge f7f78a3c)
