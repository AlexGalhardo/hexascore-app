import { createHash } from "node:crypto";
import { describe, expect, test } from "vitest";
import { type Deps, exportCovers, coverKey, type ManifestLine } from "./blob-export";

const blob = "https://store1.public.blob.vercel-storage.com/";
const full = `${blob}covers/01900000-0000-7000-8000-000000000001-0123456789abcdef.webp`;
const thumb = `${blob}covers/01900000-0000-7000-8000-000000000001-0123456789abcdef-t.webp`;
const tmdb = "https://image.tmdb.org/t/p/w500/poster.jpg";

function fakes(objects = new Map<string, Uint8Array>()): Deps & { lines: ManifestLine[]; fetched: string[] } {
	const lines: ManifestLine[] = [];
	const fetched: string[] = [];
	return {
		lines,
		fetched,
		store: {
			exists: async (key) => objects.has(key),
			write: async (key, body) => {
				objects.set(key, body);
			},
		},
		fetch: async (url) => {
			fetched.push(url);
			return url.endsWith("ffffffffffffffff.webp")
				? new Response("gone", { status: 404 })
				: new Response(new TextEncoder().encode(url));
		},
		done: new Set(lines.map((l) => l.key)),
		record: (line) => lines.push(line),
		warn: () => {},
	};
}

describe("blob-export", () => {
	test("coverKey keeps the Blob pathname and rejects other hosts", () => {
		expect(coverKey(full)).toBe("covers/01900000-0000-7000-8000-000000000001-0123456789abcdef.webp");
		expect(coverKey(thumb)).toMatch(/-t\.webp$/);
		expect(coverKey(tmdb)).toBeNull();
		expect(coverKey(`${full}?x=1`)).toBeNull();
	});

	test("copies each referenced Blob cover once, then a re-run copies nothing", async () => {
		const objects = new Map<string, Uint8Array>();
		const first = fakes(objects);

		const counts = await exportCovers([full, thumb, full, tmdb], first);
		expect(counts).toEqual({ referenced: 2, notBlob: 1, copied: 2, skipped: 0, failed: 0 });
		expect(first.lines).toEqual([
			{
				legacyUrl: full,
				key: coverKey(full),
				sha256: createHash("sha256").update(new TextEncoder().encode(full)).digest("hex"),
			},
			expect.objectContaining({ legacyUrl: thumb, key: coverKey(thumb) }),
		]);

		const again = fakes(objects);
		again.done = new Set(first.lines.map((l) => l.key));
		expect(await exportCovers([full, thumb], again)).toMatchObject({ copied: 0, skipped: 2 });
		expect(again.fetched).toEqual([]);
	});

	test("a key in the manifest but missing from the bucket is copied again; failures are counted", async () => {
		const deps = fakes();
		deps.done = new Set([coverKey(full) ?? ""]);
		const broken = `${blob}covers/01900000-0000-7000-8000-000000000002-ffffffffffffffff.webp`;

		const counts = await exportCovers([full, broken], deps);
		expect(counts).toMatchObject({ copied: 1, failed: 1 });
	});
});
