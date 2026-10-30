// Copies the covers the legacy database references from Vercel Blob to the S3 bucket under the same key
// (SDD 11 §7.6, SDD 01 §5.4: legacy key covers/<id>-<hash16>.webp kept). Run with Bun:
//
//   bun scripts/legacy/blob-export.ts --db "$NEON_RO_URL" --bucket <name> --manifest manifest.ndjson [--since <ISO time>]
//
// --since (the window's delta) limits the copy to items updated after that time; --delta is accepted as a
// label for it. The bucket credentials come from S3_ENDPOINT, S3_REGION, S3_ACCESS_KEY_ID,
// S3_SECRET_ACCESS_KEY and S3_USE_PATH_STYLE, as for the app. The manifest gets one NDJSON line
// {legacyUrl, key, sha256} per copied object; a key already in the manifest and in the bucket is skipped, so a
// re-run copies nothing. Covers outside the Blob (stored as source URLs) and unreferenced Blob objects
// (orphans) are never copied.
import { createHash } from "node:crypto";
import { appendFileSync, existsSync, readFileSync } from "node:fs";

// Same pattern as the Go catalog import (database/legacyimport/catalog.go coverKey).
const BLOB_COVER = /^https:\/\/[^/]+\.blob\.vercel-storage\.com\/(covers\/[0-9a-f-]+-[0-9a-f]{16}(?:-t)?\.webp)$/;

export function coverKey(url: string): string | null {
	return BLOB_COVER.exec(url)?.[1] ?? null;
}

export interface Store {
	exists(key: string): Promise<boolean>;
	write(key: string, body: Uint8Array): Promise<void>;
}

export interface ManifestLine {
	legacyUrl: string;
	key: string;
	sha256: string;
}

export interface Counts {
	referenced: number;
	notBlob: number;
	copied: number;
	skipped: number;
	failed: number;
}

export interface Deps {
	store: Store;
	fetch: (url: string) => Promise<Response>;
	// Keys already in the manifest.
	done: Set<string>;
	record: (line: ManifestLine) => void;
	warn: (msg: string) => void;
	concurrency?: number;
}

export async function exportCovers(urls: string[], deps: Deps): Promise<Counts> {
	const counts: Counts = { referenced: 0, notBlob: 0, copied: 0, skipped: 0, failed: 0 };
	const todo: { url: string; key: string }[] = [];
	const seen = new Set<string>();

	for (const url of urls) {
		const key = coverKey(url);
		if (!key) {
			counts.notBlob++;
			continue;
		}
		if (seen.has(key)) continue;
		seen.add(key);
		counts.referenced++;
		todo.push({ url, key });
	}

	const copy = async ({ url, key }: { url: string; key: string }): Promise<void> => {
		const attempts = 5;
		for (let attempt = 1; attempt <= attempts; attempt++) {
			try {
				if (deps.done.has(key) && (await deps.store.exists(key))) {
					counts.skipped++;
					return;
				}
				const res = await deps.fetch(url);
				if (!res.ok) throw new Error(`HTTP ${res.status}`);
				const body = new Uint8Array(await res.arrayBuffer());
				await deps.store.write(key, body);
				deps.record({ legacyUrl: url, key, sha256: createHash("sha256").update(body).digest("hex") });
				deps.done.add(key);
				counts.copied++;
				return;
			} catch (err) {
				const msg = err instanceof Error ? err.message : String(err);
				// 404 is permanent. DNS blips and the bucket's request cap are not.
				const transient = !/^HTTP 4(?!29)\d\d$/.test(msg) && /ENOTFOUND|EAI_AGAIN|ECONNRESET|ETIMEDOUT|ECONNREFUSED|HTTP 429|HTTP 5|fetch failed|request rate|unexpected error/i.test(msg);
				if (transient && attempt < attempts) {
					await new Promise((resolve) => setTimeout(resolve, 400 * attempt));
					continue;
				}
				counts.failed++;
				deps.warn(`${key}: ${msg}`);
				return;
			}
		}
	};

	let next = 0;
	const worker = async (): Promise<void> => {
		while (next < todo.length) await copy(todo[next++]);
	};
	await Promise.all(Array.from({ length: deps.concurrency ?? 8 }, worker));

	return counts;
}

function readManifest(path: string): Set<string> {
	const done = new Set<string>();
	if (!existsSync(path)) return done;
	for (const line of readFileSync(path, "utf8").split("\n")) {
		if (line.trim()) done.add((JSON.parse(line) as ManifestLine).key);
	}
	return done;
}

function flag(args: string[], name: string): string | undefined {
	const i = args.indexOf(name);
	return i >= 0 ? args[i + 1] : undefined;
}

async function main(args: string[]): Promise<number> {
	const db = flag(args, "--db");
	const bucket = flag(args, "--bucket");
	const manifest = flag(args, "--manifest");
	const since = flag(args, "--since");
	const concurrencyFlag = flag(args, "--concurrency");
	const concurrency = concurrencyFlag === undefined ? 8 : Number(concurrencyFlag);
	if (!db || !bucket || !manifest || (args.includes("--delta") && !since) || !Number.isInteger(concurrency) || concurrency < 1) {
		console.error("usage: blob-export.ts --db <url> --bucket <name> --manifest <file> [--concurrency <n>] [--delta --since <ISO time>]");
		return 2;
	}
	if (since && Number.isNaN(Date.parse(since))) {
		console.error(`--since: not a time: ${since}`);
		return 2;
	}

	const sql = new Bun.SQL(db);
	const rows: { url: string }[] = await sql`
		SELECT url FROM (
			SELECT meta->>'coverUrl' AS url, updated_at FROM items
			UNION ALL SELECT meta->>'coverThumbUrl', updated_at FROM items
		) c
		WHERE url IS NOT NULL AND (${since ?? null}::timestamptz IS NULL OR updated_at > ${since ?? null}::timestamptz)`;
	await sql.close();

	const s3 = new Bun.S3Client({
		bucket,
		endpoint: process.env.S3_ENDPOINT,
		region: process.env.S3_REGION,
		accessKeyId: process.env.S3_ACCESS_KEY_ID,
		secretAccessKey: process.env.S3_SECRET_ACCESS_KEY,
		virtualHostedStyle: process.env.S3_USE_PATH_STYLE !== "true",
	});

	const counts = await exportCovers(
		rows.map((r) => r.url),
		{
			store: {
				exists: (key) => s3.exists(key),
				write: async (key, body) => {
					await s3.write(key, body, { type: "image/webp" });
				},
			},
			fetch: (url) => fetch(url, { redirect: "error" }),
			done: readManifest(manifest),
			record: (line) => appendFileSync(manifest, `${JSON.stringify(line)}\n`),
			warn: (msg) => console.error(`blob-export: ${msg}`),
			concurrency,
		},
	);

	console.log(JSON.stringify(counts));
	return counts.failed > 0 ? 1 : 0;
}

if (import.meta.main) process.exit(await main(process.argv.slice(2)));
