# REV-P3.C.1.1

Branch: `agent/implementer/P3.C.1.1` in worktree `C:\dev\hexascore-app\wt\hexascore-p3c11`. Commit `040b782`.

Report: `.orchestration/reports/P3.C.1.1.md`. Do not treat that report as a review. The implementer did not append a merge mark.

Acceptance, from the feature worktree, then again on local main after the merge:

```
go test ./tests -count=1 -run "TestAppMinVersion|TestSecurity_A01_CrossSurfaceCredentials|TestSecurity_A01_AbilityRequired" -v
go test ./tests -count=1 -run TestAppIdempotency -v
```

Scope: shared `/app/v1` guard (min version, hx_app_ token, NotBanned, Verified, ability, ByToken 120/30, Idempotency-Key after auth). Logout `DELETE /app/v1/tokens/current` is exempt. Not plan-metered. No new dependency.

On approval, append `PRONTO_PARA_MERGE` and merge `--no-ff` into local `main` at `C:\dev\hexascore-app\hexascore-vukem`. Do not push. Re-run the commands on that main checkout, then tick ROADMAP P3.C.1.1 only if they pass. Write `C:\Users\aleex\AppData\Local\Temp\grok-goal-bff5b625ad4e\implementer\acceptance\P3.C.1.1.log` as UTF-8 without BOM. The log must include the merge commit hash. Do not stage `.env.example`, heartbeats, `status.md`, `test-results/`, `framework-prs.md`, or `__pycache__`. Do not tick P5 or P3.C.5.2. Do not edit other worktrees.
