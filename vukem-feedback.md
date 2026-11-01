# vukem feedback from HexaScore

Open queue only. Fixed, withdrawn and wontfix entries are removed; git history keeps them. Do not reuse removed ids.

Removed on 2026-10-31, all resolved on vukem `main` (23cde35a) but not in a release yet: VF-062 (#103), VF-NEW-AUDIT-1 (#107), VF-NEW-AUDIT-2 (#115), VF-NEW-AUDIT-3 (#109), VF-NEW-AUDIT-4 (#116), VF-NEW-AUDIT-5 (#110), VF-NEW-AUDIT-6 (#105), VF-NEW-P2.H.3.1 (#117), VF-NEW-P5.A.3.1 (#118), VF-NEW-FIX-gate-timeouts (#106), VF-NEW-P4.A.3.1b (#119), VF-NEW-P2.H.3.5 (withdrawn, #111), VF-NEW-FIX-covers-route (#120; adoption needs a `drivers/s3` release, Q-A14) and VF-NEW-OWNER-schemathesis (#113). HexaScore adopts them in an `ADOPT-vukem-<version>` row once the owner asks for a vukem release.

Adopted in vukem v0.35.0 (ADOPT-vukem-0.35.0), entries removed: all of the above plus VF-NEW-P4.B.2.5 (#121), VF-NEW-PGTEST-orphans (#122), VF-NEW-migrator-lock-snapshot (#123) and VF-NEW-OPENAPI-DOC-INFO (#124).

One entry per finding. Update `status` when it changes. Cite <https://vukem.dev/docs> (page URL + heading). The local `vukem/docs` tree is cited only when the site has no page, tagged `[SITE PAGE MISSING]`.

Severity: `critical` (blocks a roadmap task), `high` (workaround costs more than a day), `medium` (workaround under a day), `low` (DX or docs).
Status: `open`, `reported` (issue or PR link), `fixed in vX.Y.Z`, `wontfix`, `withdrawn`.

| id | date | severity | status |
| --- | --- | --- | --- |
| VF-NEW-EXPORT-WRITER | 2026-11-01 | medium | open |

## VF-NEW-EXPORT-WRITER: export streams only to an HTTP response, so background jobs reimplement CSV/NDJSON

- **Date**: 2026-11-01. **Severity**: medium. **Status**: open.
- **vukem version**: v0.35.0. **Docs consulted**: <https://vukem.dev/docs/digging-deeper/exports> (table "`export.NDJSON(w, r, name, rows, opts...)`"). Every streaming helper (`CSV`, `WriteCSV`, `JSON`, `NDJSON`) takes `http.ResponseWriter` and `*http.Request`; none accepts an `io.Writer`. `[NOT IN VUKEM DOCS]`: how to produce the same output from a queue job.
- **Context**: HexaScore builds two exports outside a request: the account data export (`library.export`, zipped CSV/JSON written to the bucket) and the daily bulk ID exports for the paid API (`app/jobs/export_ids.go`, NDJSON written to S3, served later through a presigned URL). Both re-implement CSV formula guarding and NDJSON line encoding that `export` already has (`workaround(vukem): VF-029` in `app/jobs/export_ids.go`).
- **Repro**: in a job with no request, try to stream an ORM iterator to a `*os.File` or an S3 upload writer with `export.NDJSON`: there is no request and no `ResponseWriter`.
- **Workaround**: hand-written encoders in the app, duplicating the formula guard (CWE-1236) logic.
- **Request**: add writer-level helpers, for example `export.EncodeCSV(io.Writer, header, rows, record, opts...)` and `export.EncodeNDJSON(io.Writer, rows, opts...)`, sharing the same formula guard and flush options; keep the HTTP helpers as thin wrappers over them, and document the job use case on the exports page.
