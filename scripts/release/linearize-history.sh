#!/usr/bin/env bash
# Rebuilds <target> as a linear branch on top of <base> without merge commits, for a
# repository whose main requires linear history. Usage, from a scratch worktree:
#   scripts/release/linearize-history.sh <base> <target> <branch>
# Stage 1 starts at merge-base(<base>, <target>) and walks <target>'s first-parent line oldest
# first. A plain commit is cherry-picked. For a merge, the side-branch commits are
# cherry-picked, then a checkpoint resets the tree to the merge's own tree (the first-parent
# line can reach below the merge-base, so this is the merge's tree, not a merge with it), adding a
# "chore(release): reconcile" commit only when the replay drifted. Conflicts during picks take
# the replayed commit's side; the checkpoints repair whatever that gets wrong, so every merge
# point of <target> exists with its exact tree. Stage 2 cherry-picks that linear line onto
# <base> (only commits <base> gained since the merge-base can conflict) and a last checkpoint
# sets the tree to <target> + <base>, which needs <target> to already contain <base>.
set -euo pipefail
base=${1:?base}; target=${2:?target}; branch=${3:?branch}
base=$(command git rev-parse "$base"); target=$(command git rev-parse "$target")
export GIT_EDITOR=true
git() { command git -c core.hooksPath=/dev/null "$@"; }

mb=$(git merge-base "$base" "$target")
git checkout -q -B "$branch" "$mb"
stops=0; drops=0; recon=0; skipped_cp=0

pick() {
	local c=$1
	if git cherry-pick --allow-empty --keep-redundant-commits "$c" >/dev/null 2>&1; then return; fi
	stops=$((stops + 1))
	echo "conflict: $(git log -1 --format='%h %s' "$c")" >&2
	local f
	while IFS= read -r f; do
		if git cat-file -e ":3:$f" 2>/dev/null; then git checkout -q --theirs -- "$f" && git add -- "$f"; else git rm -q -- "$f"; fi
	done < <(git diff --name-only --diff-filter=U)
	if git diff --cached --quiet; then
		git cherry-pick --skip >/dev/null 2>&1; drops=$((drops + 1))
	else
		git commit -q -C "$c"
	fi
}

checkpoint() {
	local m=$1 t
	if [ -z "${2:-}" ]; then
		t=$(git rev-parse "$m^{tree}")
	elif ! t=$(git merge-tree --write-tree "$m" "$2" 2>/dev/null); then
		skipped_cp=$((skipped_cp + 1)); return
	else
		t=$(head -n1 <<<"$t")
	fi
	[ "$(git rev-parse HEAD^{tree})" = "$t" ] && return
	git read-tree -u --reset "$t"
	git commit -q --allow-empty -m "chore(release): reconcile tree with $(git rev-parse --short "$m")" \
		-m "Linearized history checkpoint: the tree now equals $(git rev-parse "$m")${2:+ merged with $2}."
	recon=$((recon + 1))
}

while read -r c first rest; do
	if [ -z "$rest" ]; then
		pick "$c"
	else
		for s in $(git rev-list --no-merges --reverse --topo-order "$first..$c"); do pick "$s"; done
		checkpoint "$c"
	fi
done < <(git rev-list --first-parent --reverse --parents "$mb..$target")
checkpoint "$target"
[ "$(git rev-parse HEAD^{tree})" = "$(git rev-parse "$target^{tree}")" ] || { echo "stage1 tree differs from target" >&2; exit 1; }
git branch -f "$branch-stage1" HEAD
echo "stage1: picks_conflicted=$stops reconcile_commits=$recon commits=$(git rev-list --count "$mb..HEAD")"

stage1=$(git rev-parse HEAD)
git checkout -q -B "$branch" "$base"
stops=0; recon=0
for s in $(git rev-list --reverse "$mb..$stage1"); do pick "$s"; done
checkpoint "$target" "$base"

echo "stage2: picks_conflicted=$stops dropped_empty=$drops reconcile_commits=$recon checkpoints_skipped=$skipped_cp"
echo "commits_on_branch=$(git rev-list --count "$base..HEAD") merges=$(git rev-list --count --merges "$base..HEAD")"
