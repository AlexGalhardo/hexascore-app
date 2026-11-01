// Release-history notes in the vukem/git-cliff format (.config/cliff.toml), for the one-off
// rebuild of the v0.N.0 series. Two modes, both read-only on the repository:
//   bun scripts/release/history-notes.ts range <from>..<to> [--prev vA --tag vB]
//   bun scripts/release/history-notes.ts series <ref> --out <dir> [--step 35]
// `series` splits <ref>'s first-parent line into releases of ~step non-merge commits each,
// writes <dir>/vX.0.md per release and <dir>/tags.tsv (tag, full sha, commits, previous tag).
import { mkdirSync, writeFileSync } from "node:fs";
import { join } from "node:path";

const REPO_URL = "https://github.com/AlexGalhardo/hexascore-app";

export interface Commit {
	sha: string;
	subject: string;
	body: string;
}

interface Entry {
	group: string;
	line: string;
}

// Same order and skips as .config/cliff.toml; Maintenance is used only when a release
// would otherwise have an empty body.
const GROUPS = ["Added", "Changed", "Removed", "Fixed", "Maintenance"] as const;
const HEADER = /^(\w+)(?:\(([^)]+)\))?(!)?: (.+)$/;

export function classify(c: Commit): Entry | null {
	const m = HEADER.exec(c.subject);
	if (!m) return null;
	const [, type, scope, bang, desc] = m;
	if (type === "chore" && (scope === "release" || scope === "orchestration")) return null;
	let group: string;
	if (type === "feat") group = "Added";
	else if (type === "perf" || type === "refactor") group = "Changed";
	else if (type === "revert") group = "Removed";
	else if (type === "fix") group = "Fixed";
	else if (["docs", "test", "ci", "build", "chore", "style"].includes(type)) group = "Maintenance";
	else return null;
	const breaking = bang === "!" || /^BREAKING[ -]CHANGE:/m.test(c.body);
	const text = desc.charAt(0).toUpperCase() + desc.slice(1);
	const line = `- ${breaking ? "**BREAKING:** " : ""}${scope ? `**${scope}:** ` : ""}${text} ([${c.sha.slice(0, 7)}](${REPO_URL}/commit/${c.sha}))`;
	return { group, line };
}

export function render(commits: Commit[], prev?: string, tag?: string): string {
	const entries = commits.map(classify).filter((e): e is Entry => e !== null);
	const userFacing = entries.some((e) => e.group !== "Maintenance");
	const out: string[] = [];
	for (const g of GROUPS) {
		if (g === "Maintenance" && userFacing) continue;
		const lines = entries.filter((e) => e.group === g).map((e) => e.line);
		if (lines.length) out.push(`### ${g}\n\n${lines.join("\n")}\n`);
	}
	if (tag) {
		out.push(
			prev
				? `**Full Changelog**: ${REPO_URL}/compare/${prev}...${tag}`
				: `**Full Changelog**: ${REPO_URL}/commits/${tag}`,
		);
	}
	return `${out.join("\n").trim()}\n`;
}

function git(...args: string[]): string {
	const r = Bun.spawnSync(["git", ...args], { stdout: "pipe", stderr: "pipe" });
	if (r.exitCode !== 0) throw new Error(`git ${args.join(" ")}: ${r.stderr.toString()}`);
	return r.stdout.toString();
}

// Oldest first; merge commits are left out (their side commits are listed instead).
function commitsIn(range: string): Commit[] {
	const raw = git("log", "--no-merges", "--reverse", "--format=%H%x1f%s%x1f%b%x1e", range);
	return raw
		.split("\x1e")
		.map((r) => r.trim())
		.filter(Boolean)
		.map((r) => {
			const [sha, subject, body = ""] = r.split("\x1f");
			return { sha, subject, body };
		});
}

export interface Cut {
	tag: string;
	sha: string;
	commits: number;
	prev?: string;
}

// weights[i] = non-merge commits the i-th first-parent commit brings in (1, or a merge's side).
export function cuts(fp: string[], weights: number[], step: number): Cut[] {
	const total = weights.reduce((a, b) => a + b, 0);
	const releases = Math.max(1, Math.round(total / step));
	const out: Cut[] = [];
	let acc = 0;
	let last = 0;
	for (let i = 0; i < fp.length && out.length < releases - 1; i++) {
		acc += weights[i];
		if (acc >= step * (out.length + 1)) {
			out.push({ tag: `v0.${out.length + 1}.0`, sha: fp[i], commits: acc - last });
			last = acc;
		}
	}
	out.push({ tag: `v0.${out.length + 1}.0`, sha: fp[fp.length - 1], commits: total - last });
	out.forEach((c, i) => {
		if (i > 0) c.prev = out[i - 1].tag;
	});
	return out;
}

function series(ref: string, dir: string, step: number): void {
	const fp: string[] = [];
	const weights: number[] = [];
	for (const line of git("rev-list", "--first-parent", "--reverse", "--parents", ref).trim().split("\n")) {
		const [sha, first, ...rest] = line.split(" ");
		fp.push(sha);
		weights.push(rest.length ? Number(git("rev-list", "--count", "--no-merges", `${first}..${sha}`).trim()) : 1);
	}
	const plan = cuts(fp, weights, step);
	mkdirSync(dir, { recursive: true });
	let from = "";
	const tsv: string[] = [];
	for (const c of plan) {
		const notes = render(commitsIn(from ? `${from}..${c.sha}` : c.sha), c.prev, c.tag);
		writeFileSync(join(dir, `${c.tag}.md`), notes);
		tsv.push([c.tag, c.sha, c.commits, c.prev ?? ""].join("\t"));
		from = c.sha;
	}
	writeFileSync(join(dir, "tags.tsv"), `${tsv.join("\n")}\n`);
	const total = weights.reduce((a, b) => a + b, 0);
	console.log(`${ref}: ${total} non-merge commits, ${fp.length} first-parent, ${plan.length} releases -> ${dir}`);
}

function flag(args: string[], name: string): string | undefined {
	const i = args.indexOf(name);
	return i >= 0 ? args[i + 1] : undefined;
}

if (import.meta.main) {
	const [mode, target, ...rest] = Bun.argv.slice(2);
	if (mode === "range" && target) {
		process.stdout.write(render(commitsIn(target), flag(rest, "--prev"), flag(rest, "--tag")));
	} else if (mode === "series" && target && flag(rest, "--out")) {
		series(target, flag(rest, "--out") as string, Number(flag(rest, "--step") ?? 35));
	} else {
		console.error(
			"usage: history-notes.ts range <from>..<to> [--prev vA --tag vB]\n       history-notes.ts series <ref> --out <dir> [--step 35]",
		);
		process.exit(2);
	}
}
