# REV-P4.A.2.3

Branch: `agent/implementer/P4.A.2.3` in worktree `C:\dev\hexascore-app\wt\hexascore-p3c51`.

Report: `.orchestration/reports/P4.A.2.3.md`.

Acceptance, from that worktree, with `TEST_DATABASE_URL` pointing at the compose Postgres (`127.0.0.1:5433`, database `postgres`, sslmode disable) and `APP_ENV=test`:

- `go test ./app/community/... -count=1 -run "TestTier|TestAccuracy|TestCooldown"`
- Laplace accuracy is 0.500 at zeros, 0.750 at 8 approved and 2 rejected, and 0.045 at 20 rejections.
- `blocked` at decided ≥ 20 and accuracy < 0.20, with `blocked_until` 30 days out. 19 rejections stay `new`. Decided 10 and accuracy < 0.40 is `low`. Decided 9 with one approval is `standard`.
- A future `blocked_until` on a healthy author stays `blocked`.
- `trusted_eligible` is true at 50 approvals and accuracy ≥ 0.90, false when a `spam` rejection is inside 90 days, true again when that spam is older.
- Three qualifying rejections in 24 h set a 48 h cooldown. One `spam` sets 7 days. `duplicate` does not. `Submit` during a cooldown returns 429 `suggestion-cooldown` with `Retry-After` > 0 and `CooldownUntil`, and the suggestion stays `draft`.
- The hourly job expires a draft idle 15 days and deletes `community/covers/old.webp`, keeps a 13-day draft and an https cover, releases a claim idle 3 h, and keeps a claim idle 30 min.
- Seeding `contributor_stats` to the live counters and running `RebuildStats` again changes no column.

Reject if the implementer wrote `PRONTO_PARA_MERGE`. Reject a cooldown refusal that changes the suggestion status. Reject a second rebuild that rewrites an already correct row. Reject deleting an https cover. Reject `[no tests to run]`.

On approval, append `PRONTO_PARA_MERGE agent/implementer/P4.A.2.3` to the report and commit that mark on the branch with message `docs: mark P4.A.2.3 ready to merge`. Do not merge. Do not tick ROADMAP. Do not push.
