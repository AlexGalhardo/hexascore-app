#!/usr/bin/env bash
# Usage: scripts/release.sh vX.Y.Z [--since <rev>]
# Replaces the body of [Unreleased] in CHANGELOG.md with the git-cliff section for the new tag, adds its
# compare link, commits `chore(release): vX.Y.Z` and creates the tag. Push separately (PROMPT-WORKERS 8.1).
#
# Range: the previous release is the highest vX.Y.Z tag, not `git describe`. Releases land on GitHub as a
# squash commit and the tag can sit on a branch main never reaches, so the range starts after the newest
# commit on main's first-parent line whose tree equals that tag's tree (the merge-back of the squash, or the
# tag commit itself). --since <rev> overrides it (docs/runbooks/release-range.md).
set -euo pipefail

cd "$(dirname "$0")/.."

version="${1:-}"
since=""
if [ "${2:-}" = "--since" ]; then since="${3:?--since needs a revision}"; fi
if ! [[ "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
	echo "usage: scripts/release.sh vX.Y.Z [--since <rev>]" >&2
	exit 1
fi
command -v git-cliff >/dev/null || { echo "release: git-cliff not found (https://git-cliff.org/docs/installation)" >&2; exit 1; }
[ -z "$(git status --porcelain)" ] || { echo "release: working tree not clean" >&2; exit 1; }
! git rev-parse -q --verify "refs/tags/$version" >/dev/null || { echo "release: tag $version exists" >&2; exit 1; }

prev=$(git tag -l 'v*' | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' | sort -V | tail -n 1 || true)
if [ -n "$prev" ] && [ "$(printf '%s\n%s\n' "$prev" "$version" | sort -V | tail -n 1)" != "$version" ]; then
	echo "release: $version is not above $prev" >&2
	exit 1
fi

if [ -n "$since" ]; then
	base=$(git rev-parse --verify "$since^{commit}")
elif [ -n "$prev" ]; then
	tree=$(git rev-parse "$prev^{tree}")
	# awk reads to the end: exiting early would SIGPIPE git log and pipefail would abort the script.
	base=$(git log --first-parent --format='%H %T' HEAD | awk -v t="$tree" '$2 == t && !found { print $1; found = 1 }')
	[ -n "$base" ] || { echo "release: no commit on HEAD has the tree of $prev; pass --since <rev>" >&2; exit 1; }
fi

section=$(mktemp)
trap 'rm -f "$section"' EXIT
# git-cliff 2.14 ignores the range when --unreleased is also given, so --unreleased is only for a first release.
if [ -n "${base:-}" ]; then
	git-cliff --config .config/cliff.toml --tag "$version" "$base..HEAD" >"$section"
else
	git-cliff --config .config/cliff.toml --unreleased --tag "$version" >"$section"
fi
grep -q "^## \[${version#v}\]" "$section" || { echo "release: no releasable commits since ${prev:-the first commit}" >&2; exit 1; }

# The [Unreleased] link moves to the new tag and the new version's link compares it with the previous tag.
awk -v sec="$section" -v new="$version" -v prev="$prev" '
	/^## \[Unreleased\]/ { print; print ""; while ((getline l < sec) > 0) print l; print ""; skip = 1; next }
	skip && /^## \[/ { skip = 0 }
	/^\[Unreleased\]: .*\/compare\/.*\.\.\.HEAD$/ {
		base = $0; sub(/^\[Unreleased\]: /, "", base); sub(/\/compare\/.*$/, "", base)
		print "[Unreleased]: " base "/compare/" new "...HEAD"
		v = new; sub(/^v/, "", v)
		if (prev != "") print "[" v "]: " base "/compare/" prev "..." new
		else print "[" v "]: " base "/releases/tag/" new
		linked = 1; next
	}
	!skip { print }
	END { if (!linked) { print "release: CHANGELOG.md has no [Unreleased] compare link" > "/dev/stderr"; exit 1 } }
' CHANGELOG.md >CHANGELOG.md.new || { rm -f CHANGELOG.md.new; exit 1; }
mv CHANGELOG.md.new CHANGELOG.md

git add CHANGELOG.md
git commit -m "chore(release): $version"
git tag -a "$version" -m "$version"
echo "tagged $version (range ${base:-root}..HEAD, previous ${prev:-none}); next: PROMPT-WORKERS 8.1 step 3"
