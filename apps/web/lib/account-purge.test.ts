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
	await addUser("paying", 31);
	await addUser("stuck", 31);
	await db.insert(schema.apiKeys).values({ userId: "due", name: "k", prefix: "hx_due", hash: "h-due" });
	for (const id of ["paying", "stuck"]) {
		await db.insert(schema.subscriptions).values({
			userId: id,
			plan: "pro",
			provider: "stripe",
			status: "active",
			currentPeriodEnd: new Date(now.getTime() + DAY),
			providerRef: `sub_${id}`,
		});
	}
});

const canceled: string[] = [];
async function cancelStripe(id: string): Promise<void> {
	if (id === "sub_stuck") throw new Error("stripe down");
	canceled.push(id);
}

describe("purgeScheduledAccounts", () => {
	test("anonymizes only accounts past the 30-day grace period", async () => {
		const purged = await purgeScheduledAccounts(db, now, cancelStripe);
		expect(purged.toSorted()).toEqual(["due", "paying"]);

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

	test("revokes API keys and cancels the live Stripe subscription", async () => {
		const [key] = await db.select().from(schema.apiKeys).where(eq(schema.apiKeys.userId, "due"));
		expect(key?.revokedAt).toEqual(now);
		expect(canceled).toEqual(["sub_paying"]);
		const [sub] = await db.select().from(schema.subscriptions).where(eq(schema.subscriptions.userId, "paying"));
		expect(sub?.status).toBe("canceled");
	});

	test("keeps the account scheduled when Stripe cancellation fails", async () => {
		const [stuck] = await db.select().from(schema.user).where(eq(schema.user.id, "stuck"));
		expect(stuck?.email).toBe("stuck@example.com");
		expect(stuck?.deletionScheduledAt).not.toBeNull();
		const [sub] = await db.select().from(schema.subscriptions).where(eq(schema.subscriptions.userId, "stuck"));
		expect(sub?.status).toBe("active");
	});

	test("without Stripe configured, accounts with a live subscription are not purged", async () => {
		expect(await purgeScheduledAccounts(db, now)).toEqual([]);
	});

	test("is idempotent", async () => {
		canceled.length = 0;
		expect(await purgeScheduledAccounts(db, now, async (id) => void canceled.push(id))).toEqual(["stuck"]);
		expect(await purgeScheduledAccounts(db, now, cancelStripe)).toEqual([]);
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
