# vukem feedback from HexaScore

Open queue only. Fixed, withdrawn and wontfix entries are removed; git history keeps them. Do not reuse removed ids.

Removed on 2026-10-31, all resolved on vukem `main` (23cde35a) but not in a release yet: VF-062 (#103), VF-NEW-AUDIT-1 (#107), VF-NEW-AUDIT-2 (#115), VF-NEW-AUDIT-3 (#109), VF-NEW-AUDIT-4 (#116), VF-NEW-AUDIT-5 (#110), VF-NEW-AUDIT-6 (#105), VF-NEW-P2.H.3.1 (#117), VF-NEW-P5.A.3.1 (#118), VF-NEW-FIX-gate-timeouts (#106), VF-NEW-P4.A.3.1b (#119), VF-NEW-P2.H.3.5 (withdrawn, #111), VF-NEW-FIX-covers-route (#120; adoption needs a `drivers/s3` release, Q-A14) and VF-NEW-OWNER-schemathesis (#113). HexaScore adopts them in an `ADOPT-vukem-<version>` row once the owner asks for a vukem release.

Adopted in vukem v0.35.0 (ADOPT-vukem-0.35.0), entries removed: all of the above plus VF-NEW-P4.B.2.5 (#121), VF-NEW-PGTEST-orphans (#122), VF-NEW-migrator-lock-snapshot (#123) and VF-NEW-OPENAPI-DOC-INFO (#124).

One entry per finding. Update `status` when it changes. Cite <https://vukem.dev/docs> (page URL + heading). The local `vukem/docs` tree is cited only when the site has no page, tagged `[SITE PAGE MISSING]`.

Severity: `critical` (blocks a roadmap task), `high` (workaround costs more than a day), `medium` (workaround under a day), `low` (DX or docs).
Status: `open`, `reported` (issue or PR link), `fixed in vX.Y.Z`, `wontfix`, `withdrawn`.

| id | date | severity | status |
| --- | --- | --- | --- |
