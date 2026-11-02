// CHANGELOG.md and the GitHub releases must say the same thing (owner, 2026-11-02): every vX.Y.Z
// release has a `## [X.Y.Z]` section with the same body, and every section has a release.
//   bun scripts/release/check-sync.ts            (needs gh, authenticated or GH_TOKEN)
// Exit 1 lists each mismatch. Fix it by editing CHANGELOG.md and running
// `gh release edit vX.Y.Z --notes-file <section>` so both match again.
import { readFileSync } from "node:fs";

export type Release = { tag: string; body: string };

/** Section bodies of CHANGELOG.md keyed by version, without [Unreleased] and the link block. */
export function sections(changelog: string): Map<string, string> {
	const out = new Map<string, string>();
	const parts = changelog.replace(/\r\n/g, "\n").split(/^## /m).slice(1);
	for (const part of parts) {
		const m = /^\[(\d+\.\d+\.\d+)\][^\n]*\n/.exec(part);
		if (!m) continue;
		// The compare links at the end of the file belong to no section.
		const body = part.slice(m[0].length).replace(/\n\[[^\]\n]+\]: \S+(?=\n|$)/g, "");
		out.set(m[1], normalize(body));
	}
	return out;
}

/** GitHub adds a "Full Changelog" line on some releases; it duplicates the compare links. */
export function normalize(body: string): string {
	return body
		.replace(/\r\n/g, "\n")
		.replace(/^\*\*Full Changelog\*\*: \S+$/gm, "")
		.replace(/\n{3,}/g, "\n\n")
		.trim();
}

export function mismatches(changelog: string, releases: Release[]): string[] {
	const secs = sections(changelog);
	const problems: string[] = [];
	const seen = new Set<string>();
	for (const r of releases) {
		const v = r.tag.replace(/^v/, "");
		if (!/^\d+\.\d+\.\d+$/.test(v)) continue;
		seen.add(v);
		const s = secs.get(v);
		if (s === undefined) problems.push(`${r.tag}: GitHub release has no CHANGELOG.md section`);
		else if (s !== normalize(r.body)) problems.push(`${r.tag}: GitHub release body differs from the CHANGELOG.md section`);
	}
	for (const v of secs.keys()) {
		if (!seen.has(v)) problems.push(`v${v}: CHANGELOG.md section has no GitHub release`);
	}
	return problems;
}

if (import.meta.main) {
	const gh = Bun.spawnSync(
		["gh", "api", "--paginate", "repos/{owner}/{repo}/releases?per_page=100", "--jq", ".[] | select(.draft | not) | {tag: .tag_name, body: .body}"],
		{ stderr: "inherit" },
	);
	if (gh.exitCode !== 0) process.exit(gh.exitCode ?? 1);
	const releases = gh.stdout
		.toString()
		.split("\n")
		.filter(Boolean)
		.map((l) => JSON.parse(l) as Release);
	const problems = mismatches(readFileSync("CHANGELOG.md", "utf8"), releases);
	for (const p of problems) console.error(p);
	if (problems.length > 0) process.exit(1);
	console.log(`CHANGELOG.md and ${releases.length} GitHub releases are in sync`);
}
