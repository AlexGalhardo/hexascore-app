import { beforeAll, expect, test } from "bun:test";
import { eq } from "drizzle-orm";
import type { Database } from "../db/client";
import { subscriptions } from "../db/schema";
import { claimReminder, daysLeft, dueReminders, expireSubscriptions } from "./reminders";
import { createTestDb, createTestUser } from "./test-db";

const now = new Date("2026-10-15T12:00:00Z");
const day = 86_400_000;
let db: Database;

beforeAll(async () => {
	db = await createTestDb();
	const subs = [
		{ userId: "in7", end: now.getTime() + 6.5 * day, provider: "abacatepay" },
		{ userId: "in1", end: now.getTime() + 0.5 * day, provider: "abacatepay" },
		{ userId: "in3", end: now.getTime() + 3 * day, provider: "abacatepay" },
		{ userId: "card7", end: now.getTime() + 6.5 * day, provider: "stripe" },
		{ userId: "past", end: now.getTime() - day, provider: "abacatepay" },
	] as const;
	for (const s of subs) {
		await createTestUser(db, s.userId);
		await db.insert(subscriptions).values({
			userId: s.userId,
			plan: "pro",
			provider: s.provider,
			status: "active",
			currentPeriodEnd: new Date(s.end),
		});
	}
});

test("daysLeft rounds partial days up", () => {
	expect(daysLeft(new Date(now.getTime() + 6.1 * day), now)).toBe(7);
	expect(daysLeft(new Date(now.getTime() + 1), now)).toBe(1);
});

test("reminds PIX subscribers 7 and 1 day(s) before the end, once each", async () => {
	const due = await dueReminders(db, now);
	expect(due.map((r) => [r.userId, r.days]).sort()).toEqual([
		["in1", 1],
		["in7", 7],
	]);
	const first = due[0];
	if (!first) throw new Error("no reminder");
	expect(await claimReminder(db, first)).toBe(true);
	expect(await claimReminder(db, first)).toBe(false);
});

test("expires ended PIX periods only", async () => {
	expect(await expireSubscriptions(db, now)).toBe(1);
	const [row] = await db.select().from(subscriptions).where(eq(subscriptions.userId, "past"));
	expect(row?.status).toBe("expired");
});
