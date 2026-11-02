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

## Keeping releases in sync

CHANGELOG.md and the GitHub releases say the same thing (owner, 2026-11-02).

1. Write the release notes under `[Unreleased]` while working. `scripts/release.sh vX.Y.Z` turns them into the `## [X.Y.Z]` section; only an empty `[Unreleased]` falls back to the git-cliff list of the range above.
2. Push `main`, then the tag. `release.yml` publishes the section as the release body and fails when the section is missing.
3. `changelog-sync.yml` runs `bun scripts/release/check-sync.ts` on every release event (published, edited, deleted), every Monday and on demand. It fails on a release whose body differs from its section, a release without a section and a section without a release. GitHub's `**Full Changelog**` line and CRLF line endings are ignored.
4. To fix a released section, edit CHANGELOG.md and run `gh release edit vX.Y.Z --notes-file <section>` in the same change. Deleting or renumbering a release also removes or renames its section and compare link.
5. A release workflow that is still running recreates its release after a manual delete: wait for it to finish (`gh run list --workflow release.yml`) before deleting a release or its tag.

Tests: `bunx vitest run scripts/release/check-sync.test.ts`; locally, `bun scripts/release/check-sync.ts` (needs `gh`).
