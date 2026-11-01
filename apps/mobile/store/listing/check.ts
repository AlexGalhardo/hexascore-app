// Checks the Play listing limits for every locale file next to this script: bun apps/mobile/store/listing/check.ts
import { readdirSync, readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const LIMITS: Record<string, number> = {
	Title: 30,
	"Short description": 80,
	"Full description": 4000,
};

function sections(md: string): Record<string, string> {
	const out: Record<string, string> = {};
	for (const part of md.split(/^## /m).slice(1)) {
		const nl = part.indexOf("\n");
		out[part.slice(0, nl).trim()] = part.slice(nl + 1).trim();
	}
	return out;
}

const dir = dirname(fileURLToPath(import.meta.url));
let failed = false;
for (const file of readdirSync(dir).filter((f) => f.endsWith(".md"))) {
	const found = sections(readFileSync(join(dir, file), "utf8"));
	for (const [name, max] of Object.entries(LIMITS)) {
		// Play counts characters, not UTF-16 units or bytes.
		const len = [...(found[name] ?? "")].length;
		const ok = len > 0 && len <= max;
		failed ||= !ok;
		console.log(`${ok ? "ok  " : "FAIL"} ${file} ${name}: ${len}/${max}`);
	}
}
process.exit(failed ? 1 : 0);
