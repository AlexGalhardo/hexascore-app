import { describe, expect, test } from "bun:test";
import {
	checkRealUsers,
	checkSeedTarget,
	isLocalDatabase,
	isProductionSeed,
	REAL_USER_THRESHOLD,
	seedEnvSchema,
} from "./seed-guard";

const NEON = "postgresql://owner:pw@ep-cool-name-123456.us-east-2.aws.neon.tech/hexascore?sslmode=require";

describe("checkSeedTarget", () => {
	test("allows PGlite and local Postgres", () => {
		for (const url of [undefined, "", "postgres://u:p@localhost:5432/db", "postgres://u:p@127.0.0.1:55432/db"]) {
			expect(checkSeedTarget({ DATABASE_URL: url })).toBeNull();
		}
		expect(isLocalDatabase("postgres://u:p@[::1]:5432/db")).toBe(true);
	});

	test("refuses a remote database unless SEED_TARGET=sandbox", () => {
		expect(checkSeedTarget({ DATABASE_URL: NEON })).toContain("SEED_TARGET=sandbox");
		expect(checkSeedTarget({ DATABASE_URL: NEON, SEED_TARGET: "sandbox" })).toBeNull();
		// A host that merely contains "localhost" is still remote.
		expect(isLocalDatabase("postgres://u:p@localhost.evil.example/db")).toBe(false);
		expect(isLocalDatabase("not a url")).toBe(false);
	});

	test("always refuses VERCEL_ENV=production, even locally or with SEED_TARGET", () => {
		expect(checkSeedTarget({ VERCEL_ENV: "production" })).toContain("Refusing to seed: VERCEL_ENV=production");
		expect(checkSeedTarget({ VERCEL_ENV: "production", DATABASE_URL: NEON, SEED_TARGET: "sandbox" })).toContain(
			"VERCEL_ENV=production",
		);
		expect(checkSeedTarget({ VERCEL_ENV: "preview", DATABASE_URL: NEON, SEED_TARGET: "sandbox" })).toBeNull();
	});

	test("SEED_TARGET only accepts sandbox or production", () => {
		expect(seedEnvSchema.safeParse({ SEED_TARGET: "prod" }).success).toBe(false);
		expect(seedEnvSchema.safeParse({ SEED_TARGET: "production" }).success).toBe(true);
		expect(seedEnvSchema.safeParse({ SEED_TARGET: "sandbox" }).success).toBe(true);
	});

	test("production needs SEED_TARGET=production and the acknowledgement flag together", () => {
		const prod = { VERCEL_ENV: "production", DATABASE_URL: NEON } as const;
		expect(checkSeedTarget({ ...prod, SEED_TARGET: "production" }, true)).toBeNull();
		expect(checkSeedTarget({ ...prod, SEED_TARGET: "production" })).toContain("--i-know-this-is-production");
		expect(checkSeedTarget({ DATABASE_URL: NEON, SEED_TARGET: "production" })).toContain(
			"--i-know-this-is-production",
		);
		expect(checkSeedTarget(prod, true)).toContain("VERCEL_ENV=production");
		expect(checkSeedTarget({ ...prod, SEED_TARGET: "sandbox" }, true)).toContain("VERCEL_ENV=production");
		expect(isProductionSeed({ SEED_TARGET: "production" }, true)).toBe(true);
		expect(isProductionSeed({ SEED_TARGET: "sandbox" }, true)).toBe(false);
	});
});

describe("checkRealUsers", () => {
	test("allows up to the threshold, refuses above it unless acknowledged", () => {
		expect(checkRealUsers(0, false)).toBeNull();
		expect(checkRealUsers(REAL_USER_THRESHOLD, false)).toBeNull();
		expect(checkRealUsers(REAL_USER_THRESHOLD + 1, false)).toContain("--i-know-this-is-sandbox");
		expect(checkRealUsers(10_000, true)).toBeNull();
	});
});
