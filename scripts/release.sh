#!/usr/bin/env bash
# Usage: scripts/release.sh vX.Y.Z
# Replaces the body of [Unreleased] in CHANGELOG.md with the git-cliff section for the new tag,
# commits `chore(release): vX.Y.Z` and creates the tag. Push main and the tag separately (SDD 11 7.1).
set -euo pipefail

cd "$(dirname "$0")/.."

version="${1:-}"
if ! [[ "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
	echo "usage: scripts/release.sh vX.Y.Z" >&2
	exit 1
fi
command -v git-cliff >/dev/null || { echo "release: git-cliff not found (https://git-cliff.org/docs/installation)" >&2; exit 1; }
[ -z "$(git status --porcelain)" ] || { echo "release: working tree not clean" >&2; exit 1; }
! git rev-parse -q --verify "refs/tags/$version" >/dev/null || { echo "release: tag $version exists" >&2; exit 1; }

section=$(mktemp)
trap 'rm -f "$section"' EXIT
git-cliff --config .config/cliff.toml --unreleased --tag "$version" >"$section"
grep -q "^## \[${version#v}\]" "$section" || { echo "release: no releasable commits since the last tag" >&2; exit 1; }

awk -v sec="$section" '
	/^## \[Unreleased\]/ { print; print ""; while ((getline l < sec) > 0) print l; print ""; skip = 1; next }
	skip && /^## \[/ { skip = 0 }
	!skip { print }
' CHANGELOG.md >CHANGELOG.md.new
mv CHANGELOG.md.new CHANGELOG.md

git add CHANGELOG.md
git commit -m "chore(release): $version"
git tag -a "$version" -m "$version"
echo "tagged $version; next: git push origin main && git push origin $version"
