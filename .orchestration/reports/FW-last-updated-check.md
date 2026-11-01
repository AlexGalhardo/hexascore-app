# FW-last-updated-check tolerate squash merges
status: concluida
branch: fw: fix/vf-last-updated-squash @ 78cdd7a9 (not pushed, no PR; worktree wt/fw-FW-last-updated-check)
files: site/lib/git-dates.ts, site/scripts/check-last-updated.ts, site/scripts/check-last-updated.test.ts, site/README.md
acceptance:
- `node --test scripts/check-last-updated.test.ts` (site/) → pass 5, fail 0 (new: stale date explained by squash, squashExplainsDate cases)
- `node scripts/check-last-updated.ts` → 155 doc pages checked
- `task ci` → ci=0; `task api:check` → api=0
validation: lint ok, tests ok
changelog: none (CI tooling)
roadmap: no box
risks / follow-ups: a stale entry is accepted only when the page's newest commit subject ends with `(#N)` (squash merge) and the recorded date is not older than the change before it (or there is none). A page edited in a squash PR whose branch never refreshed its entry still fails, because the entry is then older than the previous change. Merge commits and rebase merges are not tolerated (the repo squashes). No page under site/content/docs changed, so last-updated.json needed no refresh. Specialist: push, PR, merge on green CI; after merge, `main` no longer needs a refresh PR.
vukem-feedback status line: n/a (CI tooling, no VF entry).
PR: vukem #114 open (pushed, awaiting GitHub CI)
