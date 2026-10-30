# REV-P4.A.1.3

Branch: `agent/implementer/P4.A.1.3` in worktree `C:\dev\hexascore-app\wt\hexascore-p4a13`.

Report: `.orchestration/reports/P4.A.1.3.md`.

Acceptance, with `TEST_DATABASE_URL=postgres://hexascore:hexascore@127.0.0.1:5433/postgres?sslmode=disable`:

- `go test -tags=integration ./app/community/... -count=1 -run TestTransitions`
- `go test -tags=integration ./app/community/... -count=1 -run TestTransitionWritesEvent`

On approval, append `PRONTO_PARA_MERGE` and merge `--no-ff` into local `main` at `C:\dev\hexascore-app\hexascore-vukem`. Do not push. Re-run both commands on `main`, then tick ROADMAP P4.A.1.3 only if they pass. Do not stage `.env.example`.
