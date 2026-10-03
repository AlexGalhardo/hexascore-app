import { expect, test } from "bun:test";

async function loadWith(vars: Record<string, string>): Promise<unknown> {
	const saved = { ...process.env };
	Object.assign(process.env, vars);
	try {
		return await import(`./env.ts?${Math.random()}`);
	} finally {
		process.env = saved;
	}
}

test("rejects a malformed DATABASE_URL", async () => {
	expect(loadWith({ DATABASE_URL: "mysql://x" })).rejects.toThrow("Invalid environment variables");
});

test("requires paired keys together", async () => {
	expect(loadWith({ GOOGLE_CLIENT_ID: "id", GOOGLE_CLIENT_SECRET: "" })).rejects.toThrow();
});

test("requires core keys in production", async () => {
	expect(loadWith({ VERCEL_ENV: "production", DATABASE_URL: "" })).rejects.toThrow();
});

test("accepts an empty local environment", async () => {
	expect(loadWith({ DATABASE_URL: "" })).resolves.toBeDefined();
});

test("requires auth and cron secrets for any production build", async () => {
	expect(loadWith({ NODE_ENV: "production", BETTER_AUTH_SECRET: "", CRON_SECRET: "" })).rejects.toThrow();
});
