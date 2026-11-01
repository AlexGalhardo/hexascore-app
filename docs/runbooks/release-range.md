# Release range

How `scripts/release.sh vX.Y.Z [--since <rev>]` picks the commits of a new changelog section (ADR 0013, SDD 11 §7.1, PROMPT-WORKERS 8.1).

## Why `git describe` is wrong here

Releases reach GitHub as one squash commit (`git commit-tree main^{tree} -p origin/main`), and the squash is merged back into local `main`. The tags can stay on commits `main` never reaches. Example: v0.20.0 to v0.39.0 sit on `release/history-linear`, and v0.39.0 (`e75094e9`) has the same tree as the squash `152c3faf` of PR #65. `git describe --tags main` therefore answers v0.19.0. A range from there would put about 940 commits into one section. Even a range from the squash commit is wrong, because the merge-back brings in local history that the squash does not reach.

## Rule

1. Previous release = the highest `vX.Y.Z` tag by version order (`sort -V`). The new version must be higher.
2. Base = the newest commit on `HEAD`'s first-parent line whose tree equals that tag's tree. After a squash release, this is the merge-back commit (`3b1f1b49` for v0.39.0). When the tag was created on `main`, it is the tag commit itself.
3. The section is `git-cliff --tag vX.Y.Z <base>..HEAD`, without `--unreleased`, because git-cliff 2.14 ignores the range when `--unreleased` is also given. `--unreleased` is used only for a first release, when no tag exists.
4. `--since <rev>` overrides the base. Use it when no commit has the tag's tree, for example after a hotfix branch, and the script stops and asks for it.
5. The `[Unreleased]` compare link at the bottom of `CHANGELOG.md` moves to `vX.Y.Z...HEAD`, and a `[X.Y.Z]: .../compare/<previous>...vX.Y.Z` line is added below it. The script fails if the `[Unreleased]` link is missing.

Preview without writing anything:

```sh
prev=$(git tag -l 'v*' | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' | sort -V | tail -n 1)
base=$(git log --first-parent --format='%H %T' HEAD | awk -v t="$(git rev-parse "$prev^{tree}")" '$2 == t && !f { print $1; f = 1 }')
git-cliff --config .config/cliff.toml --tag vX.Y.Z "$base..HEAD"
```

Tests: `bun test scripts/release.test.ts`. The test uses a scratch repository that reproduces the squash, the merge-back and an unreachable tag, plus a fake git-cliff that records its arguments. Vitest also runs it as part of `bun run test`.
