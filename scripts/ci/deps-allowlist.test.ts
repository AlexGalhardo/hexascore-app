import { expect, test } from "bun:test";
import { audit, load } from "./deps-allowlist";

const today = new Date("2026-11-02T00:00:00Z");

test("a future entry passes and only GHSA ids become bun audit flags", () => {
	const r = audit(
		[
			{ id: "GHSA-aaaa-bbbb-cccc", ignoreUntil: "2027-01-01", reason: "dev only" },
			{ id: "GO-2026-0001", ignoreUntil: "2027-01-01", reason: "not called" },
		],
		today,
	);
	expect(r.errors).toEqual([]);
	expect(r.flags).toEqual(["--ignore GHSA-aaaa-bbbb-cccc"]);
});

test("an expired, undated or unexplained entry fails", () => {
	const r = audit(
		[
			{ id: "GHSA-old", ignoreUntil: "2026-10-01", reason: "x" },
			{ id: "GHSA-nodate", reason: "x" },
			{ id: "GHSA-noreason", ignoreUntil: "2027-01-01" },
		],
		today,
	);
	expect(r.errors).toHaveLength(3);
});

test("osv-scanner.toml entries are all dated and explained", () => {
	const entries = load("osv-scanner.toml");
	expect(entries.length).toBeGreaterThan(0);
	for (const e of entries) {
		expect(String(e.ignoreUntil)).toMatch(/^\d{4}-\d{2}-\d{2}/);
		expect(e.reason?.length).toBeGreaterThan(0);
	}
});
