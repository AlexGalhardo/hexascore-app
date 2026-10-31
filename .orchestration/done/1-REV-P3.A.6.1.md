# REV-P3.A.6.1

Branch: `agent/implementer/P3.A.6.1` in worktree `C:\dev\hexascore-app\wt\hexascore-p3a61`.

Report: `.orchestration/reports/P3.A.6.1.md`. Do not treat that report as a review. The implementer did not append a merge mark.

Do not start until P3.A.5.3 is merged and ticked on local main. This branch was cut from `3cf4a7e`. If main has moved, merge local main into this branch before reviewing, and re-run acceptance.

Acceptance, from the feature worktree, then again on local main after the merge:

```
go test ./tests -count=1 -run TestAPIDeprecationHeaders -v
```

Scope: `v1.Deprecate` rejects `sunset - since < 183 days`. A test-only route returns `Deprecation`, `Sunset`, and `Link`. No v1 route is deprecated. The owner mail stays a runbook step. No new dependency. Do not implement P3.A.6.2. Do not run Schemathesis, k6, or ZAP.

On approval, append `PRONTO_PARA_MERGE` and merge `--no-ff` into local `main` at `C:\dev\hexascore-app\hexascore-vukem`. Do not push. Re-run the command on that main checkout, then tick ROADMAP P3.A.6.1 only if it passes. Write `C:\Users\aleex\AppData\Local\Temp\grok-goal-bff5b625ad4e\implementer\acceptance\P3.A.6.1.log` as UTF-8 without BOM (PowerShell `Out-File` is UTF-16; use .NET `UTF8Encoding($false)`). The log must include the merge commit hash. Do not stage `.env.example`, heartbeats, `status.md`, `test-results/`, `framework-prs.md`, or `__pycache__`. Do not tick P5. Do not edit other worktrees.
