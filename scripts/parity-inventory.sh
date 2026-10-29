#!/usr/bin/env sh
# Lists every legacy route, Server Action and e2e test title, one per line:
#   route <path> | action <file> <name> | e2e <spec> <title>
# Used by docs/parity-checklist.md (P1.I.6.1). Read-only on the legacy tree.
set -eu

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
# legacy/ is a sibling of the main checkout and two levels up from a worktree under wt/.
LEGACY="${1:-}"
if [ -z "$LEGACY" ]; then
	LEGACY="$ROOT/../legacy"
	[ -d "$LEGACY/apps/web" ] || LEGACY="$ROOT/../../legacy"
fi

if [ ! -d "$LEGACY/apps/web" ]; then
	echo "parity-inventory: legacy tree not found at $LEGACY" >&2
	exit 1
fi

(
	cd "$LEGACY/apps/web/app/[locale]"
	find . -name page.tsx | sed -E 's#^\./?##; s#/?page\.tsx$##' | LC_ALL=C sort
) | sed -E 's#^#route /#; s#^route /$#route /#'

for f in "$LEGACY"/apps/web/actions/*.ts; do
	name="$(basename "$f")"
	grep -oE '^export (async )?function [A-Za-z0-9_]+|^export const [A-Za-z0-9_]+' "$f" \
		| sed -E "s/^export (async )?function //; s/^export const //; s#^#action $name #"
done

for f in "$LEGACY"/qa/e2e/specs/*.spec.ts; do
	spec="$(basename "$f" .spec.ts)"
	sed -nE "s/^[[:space:]]*test\(([\"'\`])(.*)\1[,)].*/\2/p" "$f" | sed "s#^#e2e $spec #"
done
