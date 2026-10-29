# estagiario handoff (2026-10-29, after E-016)

## Done this relaunch
- E-011, E-013, E-014, E-015 concluded; E-014 and E-015 approved and integrated.
- P1.A.1.2 concluded (branch agent/estagiario/P1.A.1.2, 7fb1c52); REV queued. RustFS pinned 1.0.4 vs ROADMAP 1.0.2 (orchestrator to update).
- E-012 spike branch agent/estagiario/E-012 (local, not pushed) awaiting review.
- E-016 partial (branch agent/estagiario/E-016, 297851d, pushed): golangci-lint pin in both setup scripts. Windows setup exits 1 because 127.0.0.1:9000 is held by hexascore-spike-rustfs-1 (spike stack), so the compose step fails. Not my container; not stopped. REV queued (1-REV-E-016.md) with the port decision.

## Open items for the orchestrator
- Port 9000 conflict between the local RustFS (P1.A.1.2) and the spike stack. Decide: move local RustFS port or stop the spike stack.
- hexascore-postgres-1 (127.0.0.1:5433) is still running from the setup attempt. Run `docker compose down` in the worktree when convenient (volumes kept).
- Legal i18n copy (E-015): 8 proposed rewordings in reports/E-015.md await owner decision.

## Queues
- queue/estagiario, active/estagiario: empty after E-016 moved to done/.

## Pitfalls
- markdownlint-cli2 with the repo config lints every .md (57 files); a file argument does not narrow the run.
- jq: `\.` inside a jq double-quoted string errors; locale section keys are "0","1" objects.
- Edit needs a Read of the file in this conversation; read only the needed ranges.
- gh api contents returns empty content above 1 MB; use git/blobs/<sha>.
- PowerShell 5.1 runs native `golangci-lint version` via `& ... 2>$null` fine; `powershell -File` used because pwsh is absent.

## Branch state
- wt/estagiario: agent/estagiario/E-016 (pushed, clean).
- wt/spike-estagiario: agent/estagiario/E-012 (local).
