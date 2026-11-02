import { describe, expect, it } from "vitest";

import { mismatches, sections } from "./check-sync";

const commit = (n: string): string => `https://github.com/AlexGalhardo/hexascore-app/commit/${n.repeat(40)}`;
const two = `- **ui:** Two ([${"1".repeat(7)}](${commit("1")}))`;
const one = `- One ([${"2".repeat(7)}](${commit("2")}))`;

const log = [
	"# Changelog",
	"",
	"## [Unreleased]",
	"",
	"## [0.2.0] - 2026-11-02",
	"",
	"### Added",
	"",
	two,
	"",
	"## [0.1.0] - 2026-11-01",
	"",
	"### Fixed",
	"",
	one,
	"",
	"[Unreleased]: https://example.test/compare/v0.2.0...HEAD",
	"[0.2.0]: https://example.test/compare/v0.1.0...v0.2.0",
	"[0.1.0]: https://example.test/releases/tag/v0.1.0",
	"",
].join("\n");

describe("sections", () => {
	it("keeps only released sections and drops the link block", () => {
		expect([...sections(log).entries()]).toEqual([
			["0.2.0", `### Added\n\n${two}`],
			["0.1.0", `### Fixed\n\n${one}`],
		]);
	});
});

describe("mismatches", () => {
	it("passes when every body matches, ignoring the Full Changelog line and CRLF", () => {
		expect(
			mismatches(log, [
				{ tag: "v0.2.0", body: `### Added\r\n\r\n${two}\r\n\r\n**Full Changelog**: https://example.test/x` },
				{ tag: "v0.1.0", body: `### Fixed\n\n${one}\n` },
				{ tag: "mobile-v1.0.0", body: "ignored" },
			]),
		).toEqual([]);
	});

	it("reports a differing body, a missing section and a missing release", () => {
		expect(
			mismatches(log, [
				{ tag: "v0.2.0", body: "### Added\n\n- edited" },
				{ tag: "v0.3.0", body: "x" },
			]),
		).toEqual([
			"v0.2.0: GitHub release body differs from the CHANGELOG.md section",
			"v0.3.0: GitHub release has no CHANGELOG.md section",
			"v0.1.0: CHANGELOG.md section has no GitHub release",
		]);
	});

	it("rejects a hand-written line in a released section", () => {
		const prose = log.replace(two, "- Rework the item page for phones");
		expect(mismatches(prose, [])[0]).toBe(
			'v0.2.0: CHANGELOG.md section is not one commit per line: "- Rework the item page for phones"',
		);
	});
});
