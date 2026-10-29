# ADOPT-VF-046 report
STATUS: concluida
Branch: agent/senior/ADOPT-VF-046 (pushed). Files: app/accounts/ban.go, app/admin/{users_resource,registry,report_resources}.go, routes/admin.go, tests/admin_users_test.go, vukem-feedback.md, CHANGELOG.md.
- accounts.Ban/Unban now wrap new BanTx/UnbanTx(ctx, *database.Tx, ...) (need *Tx, not Executor, for tx.AfterCommit); same guards, event, session revoke, user.banned dispatch after commit.
- UsersResource(acc): schema key `banned` (z.Bool) declared Virtual (renders from banned_at), AfterSave applies ban/unban inside the panel Tx; absent key = no change (programmatic edits keep the ban); reason/expiry without a ban rejected in AfterSave (rolls back what the panel wrote; moved from BeforeSave since it cannot see virtuals); self/admin ban rolls the whole save back; record reloaded so the audit diff shows banned_at. mapBanErr extended to 409/422 cases.
- The POST /admin/users/{id}/ban|unban routes and their tests are kept (moderation pages; task says optional). Existing 'banned' guard tests pass unchanged.
- New test TestAdminFormBansAndUnbans (ban via form, events, unban, omitted-checkbox edit, self-ban rollback).
- go build ok; golangci-lint 0 issues; `go test -tags integration ./...` 0 FAIL (19 ok).
- Not verified in a browser: the checkbox UI rendering of the vukem form.

REV (especialista): APROVADO, see reports/REV-ADOPT-VF-046.md (CHANGELOG conflict only)
PRONTO_PARA_MERGE agent/senior/ADOPT-VF-046
