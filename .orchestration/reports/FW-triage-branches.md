# FW-triage-branches
status: parcial (PR not opened: git push was denied by the permission classifier)
branch: fw: fix/railway-otlp-traces @ e3eca299 (local in $FW worktree wt/senior-FW-railway-obs, not pushed)
files: $FW drivers/otel/{provider,provider_test}.go, docs/subsystems.md, docs/upgrade/1.0.md, site docs (deployment/docker-and-railway, deployment/logging, digging-deeper/opentelemetry)
acceptance:
- backup/specialist-feedback-a: `git cherry main backup/specialist-feedback-a` -> 0 unique commits (57 patch-equivalent on main, PR #2 merged). Deleted after tag `checkpoint/backup-specialist-feedback-a-2026-10-31` (-> e43b4f01).
- agent/implementer/railway-observability: 1 unique commit (1a8e2d3d, no PR). Cherry-picked onto current vukem main (only CHANGELOG.md conflicted; kept main's, changelog is generated), `go test ./drivers/otel/...` ok, golangci-lint 0 issues in drivers/otel.
validation: lint ok (drivers/otel), tests ok (drivers/otel); `task ci` / `task api:check` not run
changelog: none in app repo
roadmap: no box
risks / follow-ups: needs `git push -u origin fix/railway-otlp-traces` + `gh pr create` (orchestrator or owner), then specialist review. Old branch agent/implementer/railway-observability left in place until the PR exists.

## Review (especialista, 2026-10-31)
verdict: APROVADO locally; PR review pending the push (Q-A7)
re-run: `go test ./drivers/otel/...` in wt/senior-FW-railway-obs @ e3eca299 -> ok
checked: backup/specialist-feedback-a deleted only after `git cherry` showed 0 unique commits and a checkpoint tag; otel change: on Railway (RAILWAY_ENVIRONMENT or RAILWAY_SERVICE_ID) with no OTEL_EXPORTER_OTLP_METRICS_ENDPOINT, metrics stay off (the shared receiver is traces-only); 4 table cases cover on/off Railway and the explicit metrics endpoint; docs updated in docs/ and site/
next: once pushed, open the PR; I comment the same verdict on GitHub and the senior merges on green CI
