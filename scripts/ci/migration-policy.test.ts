import { expect, test } from "bun:test";

import { contractIsOldEnough, violations } from "./migration-policy.ts";

const current = "v0.8.0";

test("rejects destructive changes on hot tables without a contract marker", () => {
	expect(violations("ALTER TABLE items DROP COLUMN title", current)).toContain("DROP COLUMN");
	expect(violations("DROP TABLE ratings", current)).toContain("DROP TABLE");
	expect(violations("ALTER TABLE item_translations RENAME COLUMN title TO name", current)).toContain(
		"RENAME",
	);
	expect(violations("ALTER TABLE items ALTER COLUMN title SET NOT NULL", current)).toContain(
		"SET NOT NULL",
	);
	expect(violations("CREATE INDEX items_title_idx ON items (title)", current)).toContain(
		"CREATE INDEX",
	);
	expect(violations("CREATE INDEX CONCURRENTLY items_title_idx ON items (title)", current)).not.toContain(
		"CREATE INDEX",
	);
});

test("accepts a contract marker two minors older", () => {
	const source = "contract: v0.6.0\nALTER TABLE items DROP COLUMN title";
	expect(contractIsOldEnough("v0.6.0", current)).toBe(true);
	expect(contractIsOldEnough("v0.7.0", current)).toBe(false);
	expect(violations(source, current)).toEqual([]);
});

test("ignores DROP COLUMN inside Down", () => {
	const source = "Up: func() {}\nDown: ALTER TABLE items DROP COLUMN title";
	expect(violations(source, current)).toEqual([]);
});
