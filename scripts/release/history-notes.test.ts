import { describe, expect, it } from "vitest";

import { classify, cuts, render } from "./history-notes";

const sha = "0123456789abcdef0123456789abcdef01234567";

describe("classify", () => {
	it("formats scope, capitalises and links the sha", () => {
		expect(classify({ sha, subject: "fix(auth): block seed users", body: "" })?.line).toBe(
			`- **auth:** Block seed users ([0123456](https://github.com/AlexGalhardo/hexascore-app/commit/${sha}))`,
		);
	});

	it("skips merges, release and orchestration noise", () => {
		for (const subject of ["Merge branch 'x'", "chore(release): v0.12.4", "chore(orchestration): board sync"]) {
			expect(classify({ sha, subject, body: "" })).toBeNull();
		}
	});

	it("marks breaking changes", () => {
		expect(classify({ sha, subject: "feat(api)!: drop v0", body: "" })?.line).toContain("**BREAKING:** **api:**");
	});
});

describe("render", () => {
	it("hides Maintenance when a release has user-facing entries", () => {
		const md = render(
			[
				{ sha, subject: "feat: add x", body: "" },
				{ sha, subject: "docs(repo): note", body: "" },
			],
			"v0.1.0",
			"v0.2.0",
		);
		expect(md).toContain("### Added\n\n- Add x");
		expect(md).not.toContain("Maintenance");
		expect(md).toContain("/compare/v0.1.0...v0.2.0");
	});

	it("falls back to Maintenance so no body is empty", () => {
		expect(render([{ sha, subject: "ci: pin bun", body: "" }])).toContain("### Maintenance");
	});
});

describe("cuts", () => {
	it("cuts every step commits and puts the last release on the tip", () => {
		const fp = Array.from({ length: 100 }, (_, i) => `c${i}`);
		const plan = cuts(fp, Array(100).fill(1), 35);
		expect(plan.map((c) => [c.tag, c.sha, c.commits])).toEqual([
			["v0.1.0", "c34", 35],
			["v0.2.0", "c69", 35],
			["v0.3.0", "c99", 30],
		]);
		expect(plan[2].prev).toBe("v0.2.0");
	});
});
