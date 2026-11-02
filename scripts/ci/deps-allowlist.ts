// Prints the `bun audit --ignore` flags for the GHSA advisories ignored in osv-scanner.toml, the single dated
// allowlist (audit T-21). Exits 1 when an entry has no ignoreUntil or its date has passed, so an expired
// exception fails CI instead of silently staying ignored.
import { readFileSync } from "node:fs";

export interface Ignore {
	id: string;
	// Bun's TOML parser returns a Temporal.PlainDate for a bare date; anything whose string form starts with
	// YYYY-MM-DD is accepted.
	ignoreUntil?: unknown;
	reason?: string;
}

function day(v: unknown): string | null {
	const m = v == null ? null : String(v).match(/^\d{4}-\d{2}-\d{2}/);
	return m ? m[0] : null;
}

export function audit(entries: Ignore[], today: Date): { flags: string[]; errors: string[] } {
	const errors: string[] = [];
	const flags: string[] = [];
	for (const e of entries) {
		if (!e.reason) errors.push(`${e.id}: no reason`);
		const until = day(e.ignoreUntil);
		if (until === null) {
			errors.push(`${e.id}: no ignoreUntil date`);
			continue;
		}
		if (until < today.toISOString().slice(0, 10)) {
			errors.push(`${e.id}: ignoreUntil ${until} has passed; fix it or renew it with a new reason`);
		}
		if (e.id.startsWith("GHSA-")) flags.push(`--ignore ${e.id}`);
	}
	return { flags, errors };
}

export function load(path: string): Ignore[] {
	const doc = Bun.TOML.parse(readFileSync(path, "utf8")) as { IgnoredVulns?: Ignore[] };
	return doc.IgnoredVulns ?? [];
}

if (import.meta.main) {
	const { flags, errors } = audit(load(new URL("../../osv-scanner.toml", import.meta.url).pathname.replace(/^\/([A-Za-z]:)/, "$1")), new Date());
	for (const e of errors) console.error(`deps-allowlist: ${e}`);
	if (errors.length > 0) process.exit(1);
	console.log(flags.join(" "));
}
