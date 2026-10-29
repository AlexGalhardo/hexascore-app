REV P1.G.3.2c: APROVADO (merged cbb51f4). Branch: agent/senior/P1.G.3.2c.
Checked: GET /items/{id}/editions behind required+verified, uuid validated, public book only (404 otherwise), JSON matches Edition tags, PUT still validates edition-of-work (TestEditionNotOfWork), i18n 3 locales, no XSS.
Re-run: go test -tags integration ./tests -run 'TestEditionPickerList|TestEditionNotOfWork' -> ok (no SKIP); golangci-lint -> 0 issues; go test -tags integration ./... -> ok; bun lint/typecheck clean, test 57 passed.
Gap -> task 2-P1.G.3.2d (senior): picker lists only the first EditionPageSize (10) editions, so a user cannot pick an edition beyond them.
