# REV-ADOPT-VF-046 (rodada 0)
STATUS: APROVADO
Branch: agent/senior/ADOPT-VF-046 (no changes). Merge into current origin/main conflicts only in CHANGELOG.md [Unreleased].
Auth/security semantics of the AfterSave ban path: OK. BanTx/UnbanTx run in the panel's own transaction with LockForUpdate, so the panel write, ban columns, session/token revocation, account event and the post-commit user.banned dispatch commit or roll back together. Actor comes from the web guard in ctx; a missing user errors and rolls back (fails closed). Admin/self ban -> ErrBanAdmin -> whole save rolled back (tested). Absent "banned" key changes nothing; reason/expiry without a ban rejected and rolled back. Panel access stays admin-only (unchanged A01 coverage).
Re-run: `go build ./...` ok; `golangci-lint run --build-tags integration ./...` 0 issues; `go test -tags integration ./tests/ -run "Admin|A07"` ok (TestAdminFormBansAndUnbans PASS).
Minor (not blocking): the new test does not assert session revocation after a form ban; covered by the shared BanTx path and the existing ban-route tests. Checkbox rendering not checked in a browser.
