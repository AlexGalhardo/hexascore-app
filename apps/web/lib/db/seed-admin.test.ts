import { beforeAll, describe, expect, test } from "bun:test";
import { PGlite } from "@electric-sql/pglite";
import { pg_trgm } from "@electric-sql/pglite/contrib/pg_trgm";
import { unaccent } from "@electric-sql/pglite/contrib/unaccent";
import { verifyPassword } from "better-auth/crypto";
import { eq } from "drizzle-orm";
import { drizzle } from "drizzle-orm/pglite";
import { migrate } from "drizzle-orm/pglite/migrator";
import type { Database } from "./client";
import * as schema from "./schema";
import { SEED_ADMIN_DEFAULTS, SEED_ADMIN_ID, seedAdmin } from "./seed-admin";

let db: Database;

beforeAll(async () => {
	const client = new PGlite({ extensions: { pg_trgm, unaccent } });
	db = drizzle(client, { schema, casing: "snake_case" }) as unknown as Database;
	const migrationsFolder = new URL("./migrations", import.meta.url).pathname.replace(/^\/([A-Za-z]:)/, "$1");
	await migrate(db as unknown as Parameters<typeof migrate>[0], { migrationsFolder });
});

describe("seedAdmin", () => {
	test("creates a verified admin whose credential accepts the default password", async () => {
		expect(await seedAdmin(db, {})).toBe(SEED_ADMIN_DEFAULTS.email);
		const [user] = await db.select().from(schema.user).where(eq(schema.user.id, SEED_ADMIN_ID));
		expect(user).toMatchObject({ email: SEED_ADMIN_DEFAULTS.email, role: "admin", emailVerified: true });
		const [account] = await db.select().from(schema.account).where(eq(schema.account.userId, SEED_ADMIN_ID));
		expect(await verifyPassword({ hash: account?.password ?? "", password: SEED_ADMIN_DEFAULTS.password })).toBe(
			true,
		);
	});

	test("is idempotent and SEED_ADMIN_PASSWORD overrides the password", async () => {
		await seedAdmin(db, { SEED_ADMIN_PASSWORD: "Other#Pass2026" });
		const accounts = await db.select().from(schema.account).where(eq(schema.account.userId, SEED_ADMIN_ID));
		expect(accounts).toHaveLength(1);
		expect(await verifyPassword({ hash: accounts[0]?.password ?? "", password: "Other#Pass2026" })).toBe(true);
	});

	test("rejects a password that breaks the app's password rules", async () => {
		await expect(seedAdmin(db, { SEED_ADMIN_PASSWORD: "weak" })).rejects.toThrow();
	});
});
