# FIX-account-purge-lists Account purge leaves lists, notifications, reports
status: concluida
branch: agent/senior/FIX-account-purge-lists @ 4f484ff
files: app/accounts/deletion.go, tests/deletion_test.go (replay.go unchanged: ReplayDeletions calls the same anonymiseTx, so the restore drill is covered)
sdd: SDD 14 section 8 (user deletion purge, architect default of 14-7), SDD 03 section 7.5, AUDIT-2026-10-31 finding 1
change: purgeUserLists no-op replaced by purgeUserData (inside anonymiseTx, postgres only like user_item_editions): deletes private lists (list_items cascade), clears list_items.note and lists.description of public lists (title kept), deletes user_notifications, closes open reports about the user resolved/no_action, nulls reports.reporter_id.
acceptance:
- red first: new test failed on all 7 checks; after: `go test -tags integration -run 'TestSecurity_A01_PurgeDeletesPrivateLists|TestDeletion' ./tests/` → ok 8.9 s
- other user's lists, notifications and open reports asserted untouched
validation: golangci-lint (app/accounts, tests, integration tag) 0 issues; go build ok
changelog: Fixed: account purge now deletes private lists and notifications, clears public list free text and closes reports about the user (SDD 14 section 8)
roadmap: no box
risks / follow-ups: no dedicated replay test (needs pg_dump/pg_restore, skipped when absent); replay shares anonymiseTx. Reports on the user's public lists are left open (a moderator may still need them).

## Review (especialista, 2026-10-31)
verdict: APROVADO
re-run: `go test -tags integration ./tests -run 'TestSecurity_A01_Purge|TestDeletion|TestPurge|TestReplay|TestExport'` → ok (203 s under load); golangci-lint (app/accounts, integration tag) 0 issues; `go build ./...` ok
checked: sdd 14 §8 and sdd 03 §7.5 (private lists deleted, public list description and item notes cleared, notifications deleted, reports about the user resolved `no_action`, reporter nulled; `reports.reporter_id` is nullable and `ON DELETE SET NULL`, so the same end state as a row delete), target id format matches `library.Reports` (`target_type = 'user'`, decimal user id), sql (all values bound, placeholders via the dialect), authz n/a (purge job), runs in the anonymise transaction, so replay shares it; other users' rows asserted untouched; scope ok
note: reports on the deleted user's public lists stay open by design (moderator decision), as the report says
gate on main after merge: eda3a96 → go build ok, golangci-lint (integration tag) 0 issues, bun lint/typecheck ok, vitest 19 files / 67 tests ok, go test -tags integration -parallel 4 -timeout 30m ./... 29 packages ok
PRONTO_PARA_MERGE agent/senior/FIX-account-purge-lists
