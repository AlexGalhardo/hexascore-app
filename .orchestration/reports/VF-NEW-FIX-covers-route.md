# VF-NEW-FIX-covers-route S3 ETag and conditional GET
status: concluida
branch: fw: fix/vf-s3-etag @ debc1764 (not pushed, no PR; worktree wt/fw-VF-NEW-FIX-covers-route)
files: drivers/s3/s3_test.go, site/content/docs/digging-deeper/storage.mdx, last-updated.json
acceptance:
- `cd drivers/s3 && go test .` -> ok (TestHandlerWithAnswers304ForAnS3Object); same flow also passed against the local RustFS on :9000 (ETag present, If-None-Match -> 304)
- `task ci` -> ci=0; `task api:check` -> api=0
validation: lint ok, tests ok
changelog: see commit message (CHANGELOG is generated per release)
roadmap: no box
risks / follow-ups: No framework bug: the S3 disk has returned ETag and Last-Modified from Open since v0.33.0 (#29). The HexaScore symptom was most likely an older driver version or a layer between the disk and HandlerWith. Regression test and docs added. Specialist: push, PR, merge on green CI. last-updated.json already refreshed in the commit (committer date preserved on amend).
vukem-feedback status line: VF-NEW-FIX-covers-route -> not a bug on main (S3 disk reports ETag since v0.33.0); regression test and docs in fix/vf-s3-etag, PR open.
PR: vukem #120 open (pushed, awaiting GitHub CI)

## Reconciliation (especialista, 2026-10-31)
The app observation was right for the app: HexaScore pins the separate module `github.com/AlexGalhardo/vukem/drivers/s3` at `v0.3.1`, and no `drivers/s3/*` tag contains #29 (62f4c066, which made `Open` return ETag and Last-Modified): `git tag --contains 62f4c066 | grep s3` is empty, and `drivers/s3@v0.3.1/s3.go` sets ETag only in `Stat`. The app's `CoversFrom` passes the S3 disk straight to `storage.HandlerWith` (no wrapper), so nothing on the app side drops it. Fix path: a `drivers/s3` release (owner request only), then an ADOPT row to bump it in the app; no app FIX row.

## Review (especialista, 2026-10-31)
verdict: CORRIGIDO_E_APROVADO
re-run: `cd drivers/s3 && go test -count=1 -run 'TestHandlerWithAnswers304ForAnS3Object|TestOpenReportsValidators' .` ok
checked: regression test runs S3 disk -> storage.HandlerWith -> If-None-Match 304 with Cache-Control
fixes: 1. storage.mdx said "S3 disk (v0.33.0 and later)"; drivers/s3 is its own module and no `drivers/s3/*` tag contains #29, so the page now names `drivers/s3` releases after v0.3.1 (ca92aa48)
app: no FIX row; the app needs a `drivers/s3` release (owner) and then an ADOPT bump; adoption waits Q-A14
PR: #120
PR: #120 squash 23cde35a (33/33)
