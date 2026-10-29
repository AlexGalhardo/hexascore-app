// Renders resources/docs/*.md to resources/docs/docs.json (embedded by the Go
// binary). Every character of the source is HTML-escaped before any tag is
// emitted and only the tags below are ever produced, so the output needs no
// separate sanitiser; link targets are restricted to "/" paths and https.
import { readFileSync, readdirSync, writeFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

export interface TocEntry {
	id: string;
	text: string;
	level: number;
}

export interface DocPage {
	slug: string;
	title: string;
	body: string;
	toc: TocEntry[];
}

const escapeHtml = (s: string): string =>
	s.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;");

const slugify = (s: string): string =>
	s
		.toLowerCase()
		.replace(/[^a-z0-9\s-]/g, "")
		.trim()
		.replace(/\s+/g, "-");

function inline(src: string): string {
	const codes: string[] = [];
	// Code spans are lifted out first so their content is never re-interpreted.
	let out = escapeHtml(src).replace(/`([^`]+)`/g, (_, c: string) => {
		codes.push(`<code>${c}</code>`);
		return `\u0000${codes.length - 1}\u0000`;
	});
	out = out.replace(/\*\*([^*]+)\*\*/g, "<strong>$1</strong>");
	out = out.replace(/\[([^\]]+)\]\(([^)\s]+)\)/g, (m, text: string, href: string) =>
		/^(\/(?!\/)|https:\/\/)/.test(href) ? `<a href="${href}">${text}</a>` : m,
	);
	// biome-ignore lint/suspicious/noControlCharactersInRegex: placeholder delimiter
	return out.replace(/\u0000(\d+)\u0000/g, (_, i: string) => codes[Number(i)] ?? "");
}

const cells = (line: string): string[] =>
	line
		.trim()
		.replace(/^\||\|$/g, "")
		.split("|")
		.map((c) => c.trim());

export function renderMarkdown(md: string): { title: string; body: string; toc: TocEntry[] } {
	const lines = md.replace(/\r\n/g, "\n").split("\n");
	const html: string[] = [];
	const toc: TocEntry[] = [];
	let title = "";
	let i = 0;

	while (i < lines.length) {
		const line = lines[i] ?? "";

		if (line.startsWith("```")) {
			const lang = line.slice(3).trim().replace(/[^a-z0-9-]/gi, "");
			const code: string[] = [];
			for (i++; i < lines.length && !(lines[i] ?? "").startsWith("```"); i++) code.push(lines[i] ?? "");
			i++;
			html.push(
				`<pre><code${lang ? ` class="language-${lang}"` : ""}>${escapeHtml(code.join("\n"))}</code></pre>`,
			);
			continue;
		}

		const h = /^(#{1,3}) (.+)$/.exec(line);
		if (h) {
			const level = (h[1] ?? "").length;
			const text = (h[2] ?? "").replace(/`/g, "");
			const id = slugify(text);
			if (level === 1) title = text;
			else toc.push({ id, text, level });
			html.push(`<h${level} id="${id}">${inline(h[2] ?? "")}</h${level}>`);
			i++;
			continue;
		}

		if (line.startsWith("|")) {
			const rows: string[][] = [];
			for (; i < lines.length && (lines[i] ?? "").startsWith("|"); i++) rows.push(cells(lines[i] ?? ""));
			const [head = [], , ...body] = rows;
			html.push(
				`<table><thead><tr>${head.map((c) => `<th scope="col">${inline(c)}</th>`).join("")}</tr></thead><tbody>${body
					.map((r) => `<tr>${r.map((c) => `<td>${inline(c)}</td>`).join("")}</tr>`)
					.join("")}</tbody></table>`,
			);
			continue;
		}

		if (line.startsWith("- ")) {
			const items: string[] = [];
			for (; i < lines.length && (lines[i] ?? "").startsWith("- "); i++) {
				items.push(`<li>${inline((lines[i] ?? "").slice(2))}</li>`);
			}
			html.push(`<ul>${items.join("")}</ul>`);
			continue;
		}

		if (line.trim() === "") {
			i++;
			continue;
		}

		const para: string[] = [];
		for (; i < lines.length; i++) {
			const l = lines[i] ?? "";
			if (l.trim() === "" || /^(#{1,3} |```|\||- )/.test(l)) break;
			para.push(l);
		}
		html.push(`<p>${inline(para.join(" "))}</p>`);
	}

	return { title, body: html.join("\n"), toc };
}

// index.md is /docs; the rest keep their file name as slug. Order drives the sidebar.
const ORDER = ["index", "authentication", "pagination", "rate-limits", "errors", "changelog"];

export function buildDocs(dir: string): DocPage[] {
	const files = readdirSync(dir).filter((f) => f.endsWith(".md"));
	const slugs = files.map((f) => f.slice(0, -3)).sort((a, b) => ORDER.indexOf(a) - ORDER.indexOf(b));

	return slugs.map((slug) => ({ slug, ...renderMarkdown(readFileSync(join(dir, `${slug}.md`), "utf8")) }));
}

export const docsJSON = (dir: string): string => `${JSON.stringify(buildDocs(dir), null, "\t")}\n`;

if (import.meta.main) {
	const dir = join(dirname(fileURLToPath(import.meta.url)), "..", "resources", "docs");
	writeFileSync(join(dir, "docs.json"), docsJSON(dir));
}
