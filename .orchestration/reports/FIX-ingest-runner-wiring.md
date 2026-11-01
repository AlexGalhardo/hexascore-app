# FIX-ingest-runner-wiring
status: concluida
branch: agent/senior/FIX-ingest-runner-wiring @ 23aefce (stacked on P6.A.4.1), worktree wt/senior-P6.A.3.5a
files: `app/jobs/ingest.go`, `app/jobs/registry.go`, `internal/ingest/fetch/fetch.go`, `internal/ingest/wikidata/refresh.go`, `internal/ingest/run/continuation.go` (LazyRegistration), `tests/ingest/wiring_test.go`
red: with `handleIngest` commented out, all 3 new tests failed with `queue: no handler for job kind "ingest.run"`.
green: `go test -tags integration ./tests/ingest` ok (68s); `go test ./internal/ingest/...` ok; golangci-lint (integration tag, -j 2) 0 issues.
acceptance (second P6.A.4.1 one): `ingest:bootstrap --limit 20` then `queue:work -queue ingest,default -stop-when-empty` against an httptest fake of Wikidata/Wikipedia: fetch.wikidata, fetch.entities, normalize succeeded, 4 items (one per type), no pending records, `LIMIT 20` reached SPARQL. Also: Wikipedia chain (en/pt/es then normalize), weekly Wikidata refresh of stored entities, Open Library stage ends `skipped` and queues no normalize.
design:
- Registration is lazy (`run.LazyRegistration`): the Runner needs the queue, which is built from the jobs list.
- Steps keys: `fetch.wikidata`, `fetch.entities`, `fetch.wikipedia:{en,pt,es}`, `fetch.openlibrary`, `normalize`. `Runner.After = ingest.Advance`. Kill switches untouched (default off).
- `fetch.Step` is generic over `ingest.Connector`: 200 targets per call, cursor `{"after": id}`, parked targets skipped, failures to `ingest_failures`. Parked filtering runs after the connector iterator closes (a query on the tx while Wikipedia Targets held rows gave "conn busy").
- `fetch.entities` re-reads stored `wikidata:entity` records in full mode only; bootstrap/incremental get entities inline from `fetch.wikidata`.
- `fetch.openlibrary` fails permanently if the switch is turned on (no connector, A-05); off it is skipped by the Runner.
risks / follow-ups:
- `jobs.IngestOverride` is a package-level test seam (go rules ban package state); replace if the specialist objects.
- Wikidata `Targets` re-runs the SPARQL selection on every 200-target chunk (known design, "cheap"); for 35k movies that is ~175 big queries per type. Needs a selection cache or paging before the first real bootstrap.
- `wikidata.Connector.refs` memo is per process: if a chunk rolls back after Fetch, its entities are not re-emitted until restart.
changelog: Added (ingestion job registration, fetch steps, entity refresh).
