---
name: hexascore-release
description: Cut or backfill Hexascore releases (semver tags, GitHub releases, CHANGELOG sections, release-please manifest). Use after merging dev into main, when release-please proposes a wrong version, or when releases are missing.
---

# Releasing Hexascore

Every `feat`/`fix` that reaches `main` must end up in a GitHub release. Versions started at `v0.0.1`; before 1.0 every release bumps the patch (`bump-patch-for-minor-pre-major`).

## Normal flow

1. Merge `dev` into `main` (fast-forward) and push.
2. `release-please.yml` opens/updates `chore(main): release X.Y.Z`. Check it: `gh pr list --label "autorelease: pending"`.
3. Verify the proposed version is the next patch after `.release-please-manifest.json`. If it says `1.0.0` or skips versions, the manifest is wrong — fix it, don't merge.
4. Merge the release PR (`gh pr merge <n> --squash`). release-please tags `vX.Y.Z`, creates the GitHub release and writes the CHANGELOG section.
5. Fast-forward `dev` from `main` so the CHANGELOG/manifest commit is not lost.

## Backfill (releases missing)

1. List first-parent commits on main: `git log --first-parent --oneline main`. Each user-visible `feat`/`fix`/merge point becomes one version, in order.
2. Write notes per version (Added / Changed / Fixed / Removed).
3. Rewrite `CHANGELOG.md`: empty `[Unreleased]`, one `## [X.Y.Z] - YYYY-MM-DD` section per version, compare links at the bottom.
4. Set `.release-please-manifest.json` to the last version; commit as `chore(release): ...` (hidden type, no bump) and push dev + main.
5. Tag the exact commits with annotated tags (`GIT_COMMITTER_DATE` = commit date), `git push origin --tags`.
6. `gh release create vX.Y.Z --notes-file notes.md --verify-tag --latest=false` for all but the newest; newest with `--latest=true`.
7. Close any stale release-please PR with `--delete-branch`.
