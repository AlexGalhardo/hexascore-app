import { describe, expect, it } from "vitest";

import { coverageJSON } from "./api-coverage-json";

describe("coverageJSON", () => {
	it("turns the psql --csv output of qa/sql/api_coverage.sql into resources/data/api_coverage.json", () => {
		const csv = "type,field,items,pct\r\nbook,title,75,100.0\r\nmovie,synopsis,3,33.3\r\n";
		expect(JSON.parse(coverageJSON(csv, "2026-11-01"))).toEqual({
			measuredAt: "2026-11-01",
			rows: [
				{ type: "book", field: "title", items: 75, pct: 100 },
				{ type: "movie", field: "synopsis", items: 3, pct: 33.3 },
			],
		});
	});

	it("refuses a header or a percentage it does not expect", () => {
		expect(() => coverageJSON("type,field,pct\nbook,title,1\n", "2026-11-01")).toThrow(/header/);
		expect(() => coverageJSON("type,field,items,pct\nbook,title,1,101\n", "2026-11-01")).toThrow(/pct/);
		expect(() => coverageJSON("type,field,items,pct\n", "11/01/2026")).toThrow(/date/);
	});
});
