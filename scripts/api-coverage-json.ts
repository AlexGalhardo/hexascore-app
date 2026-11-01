// Converts the `psql --csv` output of qa/sql/api_coverage.sql into resources/data/api_coverage.json, which
// /pricing and /docs/coverage publish (P3.B.3.2, D-33):
//   bun scripts/api-coverage-json.ts /tmp/api_coverage.csv 2026-11-01 > resources/data/api_coverage.json
import { readFileSync } from "node:fs";

export interface CoverageRow {
	type: string;
	field: string;
	items: number;
	pct: number;
}

export function coverageJSON(csv: string, measuredAt: string): string {
	if (!/^\d{4}-\d{2}-\d{2}$/.test(measuredAt)) throw new Error(`measuredAt must be a YYYY-MM-DD date, got ${measuredAt}`);
	const [header, ...lines] = csv.replace(/\r\n/g, "\n").trim().split("\n");
	if (header !== "type,field,items,pct") throw new Error(`unexpected header: ${header}`);

	const rows: CoverageRow[] = lines.map((line) => {
		const [type = "", field = "", items, pct] = line.split(",");
		const row = { type, field, items: Number(items), pct: Number(pct) };
		if (!(row.pct >= 0 && row.pct <= 100) || !Number.isInteger(row.items)) throw new Error(`bad pct or items: ${line}`);
		return row;
	});

	return `${JSON.stringify({ measuredAt, rows }, null, "\t")}\n`;
}

if (import.meta.main) {
	const [file, date] = process.argv.slice(2);
	if (!file || !date) throw new Error("usage: bun scripts/api-coverage-json.ts <csv> <YYYY-MM-DD>");
	process.stdout.write(coverageJSON(readFileSync(file, "utf8"), date));
}
