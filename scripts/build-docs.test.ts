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

describe("developer guides", () => {
	const page = (slug: string): string => readFileSync(join(docsDir, `${slug}.md`), "utf8");
	const pages = docsJSON(docsDir);

	it("has an errors anchor for every problem code of the catalogue", () => {
		const catalog = readFileSync(join(docsDir, "..", "..", "app", "http", "problems", "catalog.go"), "utf8");
		const codes = [...catalog.matchAll(/\{"([a-z-]+)", http\./g)].map((m) => m[1]);
		expect(codes.length).toBeGreaterThan(15);
		for (const code of codes) expect(page("errors"), code).toContain(`### ${code}\n`);
	});

	it("documents changes, exports and webhooks, in sidebar order before errors", () => {
		const slugs = (JSON.parse(pages) as { slug: string }[]).map((p) => p.slug);
		expect(slugs.slice(slugs.indexOf("rate-limits"))).toEqual([
			"rate-limits",
			"changes",
			"exports",
			"webhooks",
			"errors",
			"coverage",
			"changelog",
		]);
	});

	it("shows Standard Webhooks verification in TypeScript and Python", () => {
		const md = page("webhooks");
		for (const needle of ["```ts", "```python", "webhook-signature", "webhook-timestamp", "whsec_", "timingSafeEqual", "compare_digest"]) {
			expect(md, needle).toContain(needle);
		}
	});

	it("uses the published SDKs, not raw fetch, in the changes guide", () => {
		const md = page("changes");
		expect(md).toContain("listChanges");
		expect(md).toContain("list_changes");
		expect(md).not.toContain("fetch(");
	});
});

it("documents the cursor as meta.next_cursor, never top-level", () => {
	for (const f of ["index.md", "pagination.md"]) {
		const md = readFileSync(join(docsDir, f), "utf8");
		expect(md).toContain("meta.next_cursor");
		expect(md).not.toMatch(/page.next_cursor|`next_cursor` (for|is `null`)/);
	}
});
