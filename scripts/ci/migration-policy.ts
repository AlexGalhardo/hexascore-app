// Guard for expand/contract migrations (SDD 11 §7.2). It reads source text.
// It is not a SQL parser. Down functions are ignored: production never runs them.

const hotTables = ["items", "ratings", "item_translations"] as const;

export type Violation = "DROP COLUMN" | "DROP TABLE" | "RENAME" | "SET NOT NULL" | "CREATE INDEX";

export function violations(source: string, currentRelease: string): Violation[] {
	const marker = source.match(/contract:\s*(v\d+\.\d+\.\d+)/);
	if (marker?.[1] && contractIsOldEnough(marker[1], currentRelease)) {
		return [];
	}

	const up = source.split(/\bDown\s*:/)[0] ?? source;
	const found: Violation[] = [];
	if (/\bDROP\s+COLUMN\b/i.test(up)) found.push("DROP COLUMN");
	if (/\bDROP\s+TABLE\b/i.test(up)) found.push("DROP TABLE");
	if (/\bRENAME\b/i.test(up)) found.push("RENAME");
	if (/\bSET\s+NOT\s+NULL\b/i.test(up)) found.push("SET NOT NULL");
	if (plainIndexOnHotTable(up)) found.push("CREATE INDEX");
	return found;
}

function plainIndexOnHotTable(up: string): boolean {
	const pattern =
		/CREATE\s+INDEX(?!\s+CONCURRENTLY)\b[\s\S]*?\bON\s+(items|ratings|item_translations)\b/gi;
	return pattern.test(up);
}

export function contractIsOldEnough(contract: string, current: string): boolean {
	const older = parseRelease(contract);
	const now = parseRelease(current);
	if (!older || !now) return false;
	if (older.major !== now.major) return older.major < now.major;
	return older.minor <= now.minor - 2;
}

function parseRelease(value: string): { major: number; minor: number } | null {
	const match = value.match(/v?(\d+)\.(\d+)\.\d+/);
	if (!match?.[1] || !match[2]) return null;
	return { major: Number(match[1]), minor: Number(match[2]) };
}

export { hotTables };
