import { request } from "@playwright/test";

// Fails fast with a clear message when the container is not up, instead of every spec timing out.
export default async function globalSetup(): Promise<void> {
	const baseURL = process.env.E2E_BASE_URL ?? "http://127.0.0.1:3000";
	const context = await request.newContext({ baseURL });
	const res = await context.get("/up", { timeout: 15_000 }).catch(() => null);
	await context.dispose();
	if (!res?.ok()) {
		throw new Error(
			`${baseURL}/up is not healthy; run: docker compose --profile app up -d --wait app`,
		);
	}
}
