# Role: framework

Read `briefs/common.md` for the stop file and the heartbeat. `<role>` is `framework`. You edit `C:\dev\hexascore-app\vukem` only, in worktree `C:\dev\hexascore-app\wt\framework`.

Scope: open pull requests, GitHub Actions failures, and open issues on `github.com/AlexGalhardo/vukem`. Merge a PR with `gh pr merge <n> --squash` only when every check is green. If the branch is behind, `gh pr update-branch` and wait for the new checks. One red job may be re-run once. A second failure is a real defect: fix it on a branch or leave the PR open and write `reports/VF-<id>.md`.

Do not tag vukem and do not ship a `BREAKING CHANGE`. Do not implement HexaScore features. The only app files you may touch are a VF status line in `vukem-feedback.md` and your own files under `.orchestration/`.

Order for the PRs that were open at the last pause, after `#63` (already merged): `#71`, `#77`, `#80`, `#81`, `#82`, `#87`, `#88`, then `#89`, `#90`, `#91`, then Dependabot `#83`–`#86` and `#92`, then `#66` and `#67`. Skip a dirty or failing PR and record why.
