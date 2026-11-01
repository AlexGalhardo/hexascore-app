import { readdirSync, readFileSync } from "node:fs";
import { basename, dirname, join, relative, resolve } from "node:path";
import { expect, test } from "vitest";

// The app root has no QueryClientProvider (ADR 0020): a component that calls a query hook without
// withQueryClient throws "No QueryClient set" at runtime, and render tests miss it because
// lib/testing.tsx always provides a client. So check it statically.
const ROOT = resolve(__dirname, "..");
const HOOK =
	/\buse(Query|Queries|Mutation|QueryClient|InfiniteQuery|SuspenseQuery|IsFetching|IsMutating)\s*\(/;
const WRAP = /\bwithQueryClient\s*\(/;
const IMPORT = /from\s+["'](\.{1,2}\/[^"']+)["']/g;

function sources(dir: string): string[] {
	return readdirSync(dir, { withFileTypes: true }).flatMap((e) => {
		const p = join(dir, e.name);
		if (e.isDirectory()) return sources(p);
		return /\.tsx?$/.test(e.name) && !/\.test\.tsx?$/.test(e.name) ? [p] : [];
	});
}

test("every query hook caller is wrapped in withQueryClient, or used only by wrapped components", () => {
	const files = sources(ROOT).filter((f) => !f.endsWith(join("lib", "queryClient.tsx")));
	const text = new Map(files.map((f) => [f, readFileSync(f, "utf8")]));
	const importers = (target: string): string[] =>
		files.filter((f) =>
			[...(text.get(f) ?? "").matchAll(IMPORT)].some(
				(m) =>
					resolve(dirname(f), m[1]).replace(/\.tsx?$/, "") ===
					target.replace(/\.tsx?$/, ""),
			),
		);

	// A file is safe when it wraps, or when it has importers and every importer is safe.
	const safe = (f: string, seen: Set<string> = new Set()): boolean => {
		if (WRAP.test(text.get(f) ?? "")) return true;
		if (seen.has(f)) return false;
		seen.add(f);
		const users = importers(f);
		return users.length > 0 && users.every((u) => safe(u, seen));
	};

	const callers = files.filter((f) => HOOK.test(text.get(f) ?? ""));
	expect(callers.length).toBeGreaterThan(0);
	const unsafe = callers.filter((f) => !safe(f)).map((f) => relative(ROOT, f));
	expect(unsafe, `wrap these with withQueryClient: ${unsafe.join(", ")}`).toEqual([]);
	// Sanity: the hook module behind the search components is reached through its importers.
	expect(callers.map((f) => basename(f))).toContain("useSuggest.ts");
});
