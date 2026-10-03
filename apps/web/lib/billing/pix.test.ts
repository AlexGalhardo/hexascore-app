import { beforeAll, describe, expect, test } from "bun:test";
import { eq } from "drizzle-orm";
import { memoryStore } from "../api/quota";
import type { Database } from "../db/client";
import { creditLedger, pixCharges, subscriptions } from "../db/schema";
import { safeEqual, signAbacatePayload, verifyAbacateSignature } from "./abacatepay";
import { pixOffer } from "./catalog";
import { handleAbacateWebhook, settlePixCharge, simulateSandboxPayment } from "./pix";
import { createTestDb, createTestUser } from "./test-db";

const SECRET = "hook-secret";
let db: Database;

beforeAll(async () => {
	db = await createTestDb();
	for (const id of ["u_plan", "u_credit", "u_hook", "u_stripe"]) await createTestUser(db, id);
});

async function charge(userId: string, purpose: string, providerRef: string): Promise<string> {
	const amountCents = pixOffer(purpose as Parameters<typeof pixOffer>[0]).amountCents;
	const [row] = await db
		.insert(pixCharges)
		.values({ userId, purpose, providerRef, amountCents })
		.returning({ id: pixCharges.id });
	if (!row) throw new Error("insert failed");
	return row.id;
}

describe("AbacatePay signature", () => {
	const body = JSON.stringify({ id: "log_1", event: "transparent.completed", data: { id: "pix_1" } });

	test("accepts the HMAC-SHA256 base64 signature of the raw body", () => {
		expect(verifyAbacateSignature(body, signAbacatePayload(body))).toBe(true);
	});

	test("rejects a missing, invalid or tampered signature", () => {
		expect(verifyAbacateSignature(body, null)).toBe(false);
		expect(verifyAbacateSignature(body, "not-a-signature")).toBe(false);
		expect(verifyAbacateSignature(body.replace("pix_1", "pix_2"), signAbacatePayload(body))).toBe(false);
		expect(verifyAbacateSignature(body, signAbacatePayload(body, "another-key"))).toBe(false);
	});

	test("safeEqual compares exactly", () => {
		expect(safeEqual("abc", "abc")).toBe(true);
		expect(safeEqual("abc", "abd")).toBe(false);
		expect(safeEqual("abc", "abcd")).toBe(false);
	});
});

describe("settlePixCharge", () => {
	test("a paid plan starts now and a renewal stacks on the current end", async () => {
		const now = new Date("2026-10-15T00:00:00Z");
		const first = await charge("u_plan", "pro_1m", "pix_plan_1");
		expect(await settlePixCharge(db, null, first, now)).toBe(true);
		expect(await settlePixCharge(db, null, first, now)).toBe(false);
		let [row] = await db.select().from(subscriptions).where(eq(subscriptions.userId, "u_plan"));
		expect(row).toMatchObject({ plan: "pro", provider: "abacatepay", status: "active" });
		expect(row?.currentPeriodEnd.toISOString()).toBe("2026-11-15T00:00:00.000Z");

		const renewal = await charge("u_plan", "pro_12m", "pix_plan_2");
		await settlePixCharge(db, null, renewal, new Date("2026-11-01T00:00:00Z"));
		[row] = await db.select().from(subscriptions).where(eq(subscriptions.userId, "u_plan"));
		expect(row?.currentPeriodEnd.toISOString()).toBe("2027-11-15T00:00:00.000Z");
	});

	test("credits are added to the ledger once and mirrored into a cached balance", async () => {
		const store = memoryStore();
		await store.setIfAbsent("hx:credits:u_credit", 5);
		const id = await charge("u_credit", "credits", "pix_credit_1");
		await settlePixCharge(db, store, id);
		await settlePixCharge(db, store, id);
		const rows = await db.select().from(creditLedger).where(eq(creditLedger.userId, "u_credit"));
		expect(rows).toHaveLength(1);
		expect(rows[0]).toMatchObject({ delta: 25_000, ref: `pix:${id}` });
		expect(await store.get("hx:credits:u_credit")).toBe(25_005);
	});

	test("never overwrites an active Stripe subscription", async () => {
		const end = new Date("2027-01-01T00:00:00Z");
		await db.insert(subscriptions).values({
			userId: "u_stripe",
			plan: "business",
			provider: "stripe",
			status: "active",
			currentPeriodEnd: end,
			providerRef: "sub_x",
		});
		const id = await charge("u_stripe", "pro_1m", "pix_stripe_1");
		await settlePixCharge(db, null, id, new Date("2026-10-15T00:00:00Z"));
		const [row] = await db.select().from(subscriptions).where(eq(subscriptions.userId, "u_stripe"));
		expect(row).toMatchObject({ provider: "stripe", plan: "business", providerRef: "sub_x" });
	});
});

describe("handleAbacateWebhook", () => {
	const send = (payload: object, opts: { secret?: string | null; signature?: string | null } = {}) => {
		const body = JSON.stringify(payload);
		return handleAbacateWebhook(
			body,
			{
				webhookSecret: opts.secret === undefined ? SECRET : opts.secret,
				signature: opts.signature === undefined ? signAbacatePayload(body) : opts.signature,
			},
			{ db, store: null, secret: SECRET },
		);
	};

	test("rejects a wrong webhookSecret or signature with 401", async () => {
		const payload = { id: "log_x", event: "transparent.completed", data: { id: "pix_none" } };
		expect((await send(payload, { secret: "wrong" })).status).toBe(401);
		expect((await send(payload, { secret: null })).status).toBe(401);
		expect((await send(payload, { signature: "bad" })).status).toBe(401);
	});

	test("applies a paid charge once per event id", async () => {
		const id = await charge("u_hook", "business_1m", "pix_hook_1");
		const payload = {
			id: "log_hook_1",
			event: "transparent.completed",
			devMode: true,
			data: { id: "pix_hook_1", amount: 14_900, status: "PAID" },
		};
		expect((await send(payload)).status).toBe(200);
		expect((await send(payload)).status).toBe(200);
		const [c] = await db.select().from(pixCharges).where(eq(pixCharges.id, id));
		expect(c?.status).toBe("paid");
		const [row] = await db.select().from(subscriptions).where(eq(subscriptions.userId, "u_hook"));
		expect(row).toMatchObject({ plan: "business", provider: "abacatepay", status: "active" });
	});

	test("ignores a payment whose amount differs from the stored charge", async () => {
		const id = await charge("u_hook", "credits", "pix_hook_2");
		const payload = { id: "log_hook_2", event: "transparent.completed", data: { id: "pix_hook_2", amount: 1 } };
		expect((await send(payload)).status).toBe(200);
		const [c] = await db.select().from(pixCharges).where(eq(pixCharges.id, id));
		expect(c?.status).toBe("pending");
	});

	test("acknowledges other events without changes", async () => {
		expect((await send({ id: "log_o", event: "transparent.refunded", data: {} })).status).toBe(200);
	});
});

describe("simulateSandboxPayment", () => {
	const now = new Date("2026-10-15T00:00:00Z");
	function deps(enabled = true) {
		const calls: string[] = [];
		return {
			calls,
			deps: {
				db,
				enabled,
				now: () => now,
				simulate: async (ref: string) => {
					calls.push(ref);
				},
			},
		};
	}

	test("simulates only the owner's pending charge and never settles it locally", async () => {
		const id = await charge("u_plan", "pro_1m", "pix_sim_ok");
		const { calls, deps: d } = deps();
		expect(await simulateSandboxPayment("u_plan", id, d)).toBe("ok");
		expect(calls).toEqual(["pix_sim_ok"]);
		const [row] = await db.select().from(pixCharges).where(eq(pixCharges.id, id));
		expect(row?.status).toBe("pending");
	});

	test("is refused when sandbox is off, before touching the database or AbacatePay", async () => {
		const id = await charge("u_plan", "pro_1m", "pix_sim_off");
		const { calls, deps: d } = deps(false);
		expect(await simulateSandboxPayment("u_plan", id, d)).toBe("not_configured");
		expect(calls).toEqual([]);
	});

	test("another user's charge reads as not found", async () => {
		const id = await charge("u_plan", "pro_1m", "pix_sim_other");
		const { calls, deps: d } = deps();
		expect(await simulateSandboxPayment("u_credit", id, d)).toBe("not_found");
		expect(calls).toEqual([]);
	});

	test("paid or expired charges are not simulated again", async () => {
		const paid = await charge("u_plan", "pro_1m", "pix_sim_paid");
		await db.update(pixCharges).set({ status: "paid" }).where(eq(pixCharges.id, paid));
		const stale = await charge("u_plan", "pro_1m", "pix_sim_stale");
		await db
			.update(pixCharges)
			.set({ expiresAt: new Date(now.getTime() - 1000) })
			.where(eq(pixCharges.id, stale));
		const { calls, deps: d } = deps();
		expect(await simulateSandboxPayment("u_plan", paid, d)).toBe("not_pending");
		expect(await simulateSandboxPayment("u_plan", stale, d)).toBe("not_pending");
		expect(calls).toEqual([]);
	});
});
