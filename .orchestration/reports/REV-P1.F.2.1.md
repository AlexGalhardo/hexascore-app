# REV-P1.F.2.1 (rodada 1)
VERDICT: CORRIGIDO_E_APROVADO
Branch: agent/especialista/P1.F.2.1-fix (from origin/main, 1 commit 30665f2). Files: app/library/profile.go, app/controllers/profile_public_controller.go, tests/library_test.go.
Fixes:
- SDD 14 §8 "Banned owner: noindex": /u/{nickname} of a currently banned owner (banned_at set, not expired) now sends X-Robots-Tag noindex + seo.indexable=false; doc comment claimed banned = 404 (not true) -> corrected. Asserted in TestPublicProfilePrivacy.
- New TestMergeMovesUserEditions: user edition follows work on catalog merge; target choice wins (merge.go path had no test).
Checked OK: SQL bound params, sort via whitelist map, filter 422 validation, composite FK enforces edition-of-work, merge delete-before-update ordering, deletion purge, 4 queries max (no N+1), private profile leaks no counts (web + API), rate limit on edition writes.
Re-run: go test -count=1 -tags integration ./tests -run 'TestLibraryGroups|TestEditionNotOfWork|TestPublicProfilePrivacy|TestMergeMovesUserEditions|TestMergeMovesListItems' -> 5x PASS, ok 4.565s. golangci-lint run --build-tags integration ./... -> 0 issues. go build ./... ok.
Notes (no action): /u/{nickname} has no rate limit (NFR-14-4 lists none for reads); Initial panics only if username is empty (prevented by signup validation [NÃO VERIFICADO]).
PRONTO_PARA_MERGE agent/especialista/P1.F.2.1-fix
