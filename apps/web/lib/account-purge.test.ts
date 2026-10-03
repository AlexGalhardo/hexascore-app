import { beforeAll, describe, expect, test } from "bun:test";
import { PGlite } from "@electric-sql/pglite";
import { pg_trgm } from "@electric-sql/pglite/contrib/pg_trgm";
import { unaccent } from "@electric-sql/pglite/contrib/unaccent";
import { eq } from "drizzle-orm";
import { drizzle } from "drizzle-orm/pglite";
import { migrate } from "drizzle-orm/pglite/migrator";
import { deletionDate, purgeCutoff, secondsLeft } from "./account-deletion";
import { purgeScheduledAccounts } from "./account-purge";
import type { Database } from "./db/client";
import * as schema from "./db/schema";

const DAY = 24 * 60 * 60 * 1000;
const now = new Date("2026-10-02T12:00:00Z");
let db: Database;

async function addUser(id: string, scheduledDaysAgo: number | null): Promise<void> {
	await db.insert(schema.user).values({
		id,
		name: `Name ${id}`,
		email: `${id}@example.com`,
		username: id,
		emailVerified: true,
		deletionScheduledAt: scheduledDaysAgo === null ? null : new Date(now.getTime() - scheduledDaysAgo * DAY),
	});
	await db.insert(schema.account).values({
		id: `acc-${id}`,
		accountId: id,
		providerId: "credential",
		userId: id,
		password: "hash",
		updatedAt: now,
	});
	await db.insert(schema.session).values({
		id: `ses-${id}`,
		token: `tok-${id}`,
		userId: id,
		expiresAt: new Date(now.getTime() + DAY),
		updatedAt: now,
	});
	await db.insert(schema.twoFactor).values({ id: `2fa-${id}`, secret: "s", backupCodes: "b", userId: id });
}

beforeAll(async () => {
	const client = new PGlite({ extensions: { pg_trgm, unaccent } });
	db = drizzle(client, { schema, casing: "snake_case" }) as unknown as Database;
	const migrationsFolder = new URL("./db/migrations", import.meta.url).pathname.replace(/^\/([A-Za-z]:)/, "$1");
	await migrate(db as unknown as Parameters<typeof migrate>[0], { migrationsFolder });
	await addUser("due", 31);
	await addUser("grace", 29);
	await addUser("active", null);
});

describe("purgeScheduledAccounts", () => {
	test("anonymizes only accounts past the 30-day grace period", async () => {
		const purged = await purgeScheduledAccounts(db, now);
		expect(purged).toEqual(["due"]);

		const [due] = await db.select().from(schema.user).where(eq(schema.user.id, "due"));
		expect(due).toMatchObject({
			email: "deleted-due@invalid",
			name: "Deleted user",
			username: null,
			emailVerified: false,
			deletionScheduledAt: null,
		});
		for (const table of [schema.session, schema.account, schema.twoFactor]) {
			expect(await db.select().from(table).where(eq(table.userId, "due"))).toHaveLength(0);
		}

		const [grace] = await db.select().from(schema.user).where(eq(schema.user.id, "grace"));
		expect(grace?.email).toBe("grace@example.com");
		expect(await db.select().from(schema.session).where(eq(schema.session.userId, "grace"))).toHaveLength(1);
		expect(await db.select().from(schema.account).where(eq(schema.account.userId, "active"))).toHaveLength(1);
	});

	test("is idempotent", async () => {
		expect(await purgeScheduledAccounts(db, now)).toEqual([]);
	});
});

describe("deletion helpers", () => {
	test("deletion date is 30 days after scheduling", () => {
		expect(deletionDate(now).getTime() - now.getTime()).toBe(30 * DAY);
		expect(purgeCutoff(now).getTime()).toBe(now.getTime() - 30 * DAY);
	});

	test.each([
		[0, 10],
		[999, 10],
		[1000, 9],
		[9999, 1],
		[10_000, 0],
		[60_000, 0],
	])("secondsLeft after %d ms -> %d", (elapsed, left) => {
		expect(secondsLeft(1_000, 1_000 + elapsed)).toBe(left);
	});
});
