# REV-P4.A.2.2

Branch: `agent/implementer/P4.A.2.2` in worktree `C:\dev\hexascore-app\wt\hexascore-p3c51`.

Report: `.orchestration/reports/P4.A.2.2.md`.

Acceptance, from that worktree, with `TEST_DATABASE_URL` pointing at the compose Postgres (`127.0.0.1:5433`, database `postgres`, sslmode disable) and `APP_ENV=test`:

- `go test -tags=integration ./app/community/... -count=1 -run "TestDuplicate|TestFieldLock|TestStaleBeforeAtSubmit"`
- Own open field returns 409 `suggestion-duplicate` with `ExistingID` set to that suggestion.
- An identical `after` from another author returns 409 `suggestion-duplicate` with `ExistingID` empty.
- A different `after` from another author returns 303 and sets `meta.competing` on both rows.
- An expired lock is ignored. A live lock returns 423 `field-locked` with `LockReason` and `Field`, and the suggestion stays `draft`.
- A scalar `before` that is not the current catalogue value returns 409 `stale-before` with `Field` and `Current`.

Reject if the implementer wrote `PRONTO_PARA_MERGE`. Reject a refused submit that changes status or consumes the daily quota. Reject an expired lock that is deleted. Reject a new HTTP route.

On approval, append `PRONTO_PARA_MERGE agent/implementer/P4.A.2.2` to the report and commit that mark on the branch with message `docs: mark P4.A.2.2 ready to merge`. Do not merge. Do not tick ROADMAP. Do not push.
