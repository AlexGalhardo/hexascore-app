# vukem feedback from HexaScore

Open queue only. Fixed, withdrawn and wontfix entries are removed; git history keeps them. Do not reuse removed ids.

Removed on 2026-10-31, all resolved on vukem `main` (23cde35a) but not in a release yet: VF-062 (#103), VF-NEW-AUDIT-1 (#107), VF-NEW-AUDIT-2 (#115), VF-NEW-AUDIT-3 (#109), VF-NEW-AUDIT-4 (#116), VF-NEW-AUDIT-5 (#110), VF-NEW-AUDIT-6 (#105), VF-NEW-P2.H.3.1 (#117), VF-NEW-P5.A.3.1 (#118), VF-NEW-FIX-gate-timeouts (#106), VF-NEW-P4.A.3.1b (#119), VF-NEW-P2.H.3.5 (withdrawn, #111), VF-NEW-FIX-covers-route (#120; adoption needs a `drivers/s3` release, Q-A14) and VF-NEW-OWNER-schemathesis (#113). HexaScore adopts them in an `ADOPT-vukem-<version>` row once the owner asks for a vukem release.

One entry per finding. Update `status` when it changes. Cite <https://vukem.dev/docs> (page URL + heading). The local `vukem/docs` tree is cited only when the site has no page, tagged `[SITE PAGE MISSING]`.

Severity: `critical` (blocks a roadmap task), `high` (workaround costs more than a day), `medium` (workaround under a day), `low` (DX or docs).
Status: `open`, `reported` (issue or PR link), `fixed in vX.Y.Z`, `wontfix`, `withdrawn`.

| id | date | severity | status |
| --- | --- | --- | --- |
| VF-NEW-P4.B.2.5 | 2026-10-31 | medium | open |
| VF-NEW-PGTEST-orphans | 2026-10-31 | low | open |

## VF-NEW-P4.B.2.5: Admin panel updates write every column from a record loaded outside the transaction

- **Date**: 2026-10-31. **Severity**: medium. **Status**: open.
- **vukem version**: v0.34.0; same code on `main` 23cde35a. **Docs consulted**: <https://vukem.dev/docs> admin resources, `BeforeSave` and `AfterSave`. Neither the docs nor the godoc of `Resource.BeforeSave` mention concurrent writers. `[NOT IN VUKEM DOCS]`.
- **Context**: HexaScore applies community edits to catalogue rows in a transaction under an advisory lock and a row lock, and records per-field provenance. P4.B.2.5 planned admin field edits through the panel with `BeforeSave`. `admin/handlers.go` `save` loads the previous record before `DB.Tx` and then calls `orm.Update`. `orm.Update` (`orm/write.go`) sets every non-key, non-readonly column from the in-memory struct. A panel save that runs while another transaction changes the same row writes back the stale values of fields the admin never touched, including provenance columns. This is a lost update. The row is not locked and there is no version check, so the panel cannot detect it.
- **Repro**: open the panel edit form for a row. In another session, update a different column of that row and commit. Submit the form unchanged except one field: the other session's column is reverted.
- **Workaround**: HexaScore keeps the catalogue resources read-only in the panel and edits fields through its own route (`POST /admin/catalog/{type}/{id}/fields`), which locks the row and writes only the changed columns with `Query.Update(ctx, orm.Set{...})`.
- **Request**: in `Resource` saves, either (a) re-read the record `FOR UPDATE` inside the transaction and write only the columns the form changed, or (b) support an optional version column (`orm:"version"`) that `orm.Update` checks and increments, and map a mismatch to the existing 409 form error. Document the behaviour on the admin resources page.

## VF-NEW-PGTEST-orphans: pgtest databases leak when the test process is killed

- **Date**: 2026-10-31. **Severity**: low. **Status**: open.
- **vukem version**: v0.34.0; same code on `main` 23cde35a (`drivers/postgres/pgtest/pgtest.go`). **Docs consulted**: <https://vukem.dev/docs> testing with Postgres; nothing on cleanup after an interrupted run. `[NOT IN VUKEM DOCS]`.
- **Context**: `pgtest` creates one database per test and drops it in `tb.Cleanup`. Cleanup does not run when the process is killed, for example on a timeout, a Ctrl+C or a stopped benchmark. HexaScore stopped long `BenchmarkComputeSimilar` runs and was left with orphan `pgtest_*` databases on the shared Postgres, each holding a 50k-row seed.
- **Repro**: start `go test -tags integration -run TestX ./...` with a test that uses `pgtest`, and kill the process while the test runs. The database stays.
- **Workaround**: drop `pgtest_*` databases by hand when no test run is active.
- **Request**: at startup, sweep `pgtest_*` databases older than a threshold (for example 6 h, configurable with `PGTEST_SWEEP_AFTER`) that have no active connections. Alternatively, name databases with the creating process id and drop those whose process is gone. Document the leak and the sweep.
