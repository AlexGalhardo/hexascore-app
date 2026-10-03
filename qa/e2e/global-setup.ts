import { request } from "@playwright/test";

// Catalog reads go through unstable_cache (stale-while-revalidate): after a fresh seed the first hit per locale may
// still serve an empty cached list. Two warm-up passes make every spec see the real data.
export default async function globalSetup(): Promise<void> {
	const baseURL = process.env.E2E_BASE_URL ?? "http://localhost:3000";
	const bypass = process.env.VERCEL_AUTOMATION_BYPASS_SECRET;
	const context = await request.newContext({
		baseURL,
		extraHTTPHeaders: bypass ? { "x-vercel-protection-bypass": bypass } : {},
	});
	const paths = ["", "/pt-br", "/es"].flatMap((prefix) =>
		["/", "/games", "/movies", "/series", "/books"].map((p) => `${prefix}${p}`),
	);
	for (let pass = 0; pass < 2; pass++) {
		await Promise.all(paths.map((path) => context.get(path, { timeout: 60_000 }).catch(() => null)));
	}
	await context.dispose();
}
