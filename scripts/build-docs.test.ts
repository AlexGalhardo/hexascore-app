import { readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";

import { docsJSON, renderMarkdown } from "./build-docs";

const docsDir = join(dirname(fileURLToPath(import.meta.url)), "..", "resources", "docs");

describe("renderMarkdown", () => {
	it("escapes raw HTML and refuses unsafe link targets", () => {
		const { body } = renderMarkdown(
			"# T\n\n<script>alert(1)</script> [x](javascript:alert(1)) [ok](/docs/errors#a) `<b>` [r](/\\evil.example) [p](//evil.example)",
		);
		expect(body).not.toContain("<script>");
		expect(body).toContain("&lt;script&gt;");
		expect(body).not.toContain('href="javascript:');
		expect(body).not.toContain('evil.example"');
		expect(body).toContain('<a href="/docs/errors#a">ok</a>');
		expect(body).toContain("<code>&lt;b&gt;</code>");
	});

	it("gives headings ids and one h1", () => {
		const r = renderMarkdown("# Title\n\n## quota-exhausted\n\ntext");
		expect(r.title).toBe("Title");
		expect(r.toc).toEqual([{ id: "quota-exhausted", text: "quota-exhausted", level: 2 }]);
		expect(r.body.match(/<h1/g)).toHaveLength(1);
	});
});

it("docs.json is up to date (run: bun scripts/build-docs.ts)", () => {
	expect(readFileSync(join(docsDir, "docs.json"), "utf8")).toBe(docsJSON(docsDir));
});
