# REV-P3.A.6.2

Branch: `agent/implementer/P3.A.6.2` in worktree `C:\dev\hexascore-app\wt\hexascore-p3a62`.

Report: `.orchestration/reports/P3.A.6.2.md`. Do not treat that report as a review. The implementer did not append a merge mark.

Acceptance, from the feature worktree, then again on local main after the merge:

```
go test ./app/webhooks -count=1 -run "TestPingRejectsMalformedID|TestPublicURL"
```

The contract commands already ran green on this worktree (openapi diff empty, Redocly 2.72.1, Schemathesis 4.29.0 exit 0, k6 2.3.0 exit 0 with list p95 18ms and detail p95 14ms). Re-run those on local main only if the merge is clean. ZAP is not in the acceptance commands.

Scope: non-UUID `POST /api/v1/webhooks/{id}/test` is 404; webhook DNS lookup is bounded to 2s; OpenAPI goldens match the export; CI seed sets `created_at`. No new dependency. Servers stay `https://api.hexascore.app` without `/v1`.

On approval, append `PRONTO_PARA_MERGE` and merge `--no-ff` into local `main` at `C:\dev\hexascore-app\hexascore-vukem`. Do not push. Do not tick the ROADMAP box. Do not stage `.env.example`, heartbeats, `status.md`, `test-results/`, `framework-prs.md`, or `__pycache__`.
