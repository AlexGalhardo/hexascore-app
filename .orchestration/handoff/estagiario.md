# estagiario handoff (2026-10-29, after E-015, context heavy)

## Done this relaunch
- E-011 integrated; E-012 spike branch agent/estagiario/E-012 (local, not pushed) awaiting review.
- E-013 concluida (reports/E-013.md, E-013-table.md). openapi namespace drop, keep? en-only.
- P1.A.1.2 concluida: branch agent/estagiario/P1.A.1.2 (7fb1c52, pushed). REV queued. RustFS pinned 1.0.4 vs ROADMAP 1.0.2 (orchestrator to update).
- E-014 concluida and approved/integrated by coordinator (783ea22). 443 third-party markdownlint errors in .claude/skills left as is.
- E-015 concluida: branch agent/estagiario/E-015 (3aacf1a, pushed). 8 non-legal keys reworded in 3 locales. Legal keys left for owner (list in reports/E-015.md). REV queued.

## Queues
- queue/estagiario, reserve/estagiario: empty. active/estagiario: empty.
- Waiter running for new task/STOP (bd0rk214a may have timed out; restart the wait if no notification).

## Pitfalls
- jq regex: `\.` inside a double-quoted jq string breaks ("Invalid escape"); use startswith or split instead.
- Locale JSON sections are objects keyed "0","1" (use `.sections["1"].body["4"]`).
- Edit tool needs Read of the file in this conversation; large JSON reads are heavy, read only the line ranges needed.
- gh api contents returns empty content for files >1 MB; use git/blobs/<sha>.
- Bash heredoc with many files in one call failed to parse; write files with Write or separate calls.
- VF numbering: VF-NEW-<task ID> only.

## Branch state
- wt/estagiario: agent/estagiario/E-015 (pushed, clean).
- wt/spike-estagiario: agent/estagiario/E-012 (local).
