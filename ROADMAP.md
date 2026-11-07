# HexaScore ROADMAP: remaining work for this session

## Production catalogue worker

- [ ] Align the production web and worker on the same tested revision (the staging environment was removed on 2026-11-07, ADR 0028). Set `QUEUE_CONNECTION=database` on both, enable matching ingestion source flags, and confirm scheduled jobs reach the `ingest` queue.
- [ ] Complete the M0 proof, now run in production behind the ingest flags, in [the worker runbook](docs/runbooks/worker.md): seven healthy days; at least 1,000 refreshed items; at least 97% accuracy in the sample; healthy heartbeat, failed queue and digest; cover coverage and spend within the RFC-001 limits.
- [ ] After M0, provision and run the production Railway worker continuously, then record 24-hour throughput, freshness, failures, memory and cost. Follow the runbook's rollback steps if the checks fail.
- [ ] Implement and validate the discovery and per-source stale-record selection needed to move through the 250k, 500k and 1M catalogue gates; include trailer verification v2 and source-terms review before enabling each source in production.

## Home recommendations

- [ ] Build the RFC-002 external popularity writer so the released-item grids can substantiate the requested internet trends from the last 30 days; validate its data, ranking and release-date exclusion.
- [ ] Benchmark home queries and time to first byte against a 1M-item fixture before promoting the new home ranking to production.

## Local production catalogue preview

- [ ] Run [the local snapshot script](production-postgres-local.md) from the owner's terminal to create the ignored production catalogue seed and restore the isolated `hexascore_prod_preview` database. Railway CLI credentials and Docker are unavailable to the isolated coding terminal.
