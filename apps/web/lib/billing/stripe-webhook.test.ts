import { beforeAll, describe, expect, test } from "bun:test";
import { eq } from "drizzle-orm";
import Stripe from "stripe";
import { memoryStore } from "../api/quota";
import type { Database } from "../db/client";
import { creditLedger, subscriptions } from "../db/schema";
import { STRIPE_LOOKUP_KEYS } from "./catalog";
import { handleStripeWebhook, type StripeWebhookDeps } from "./stripe-webhook";
import { createTestDb, createTestUser } from "./test-db";

const secret = "whsec_test_secret";
const stripe = new Stripe("sk_test_offline");
let db: Database;

beforeAll(async () => {
	db = await createTestDb();
	await createTestUser(db, "u_credits");
	await createTestUser(db, "u_sub");
});

async function signed(event: Record<string, unknown>): Promise<{ body: string; signature: string }> {
	const body = JSON.stringify({ object: "event", api_version: Stripe.API_VERSION, ...event });
	return { body, signature: await stripe.webhooks.generateTestHeaderStringAsync({ payload: body, secret }) };
}

function subscriptionFixture(status: Stripe.Subscription.Status, periodEnd: number): Stripe.Subscription {
	return {
		id: "sub_1",
		object: "subscription",
		customer: "cus_1",
		status,
		cancel_at_period_end: false,
		cancel_at: null,
		ended_at: null,
		metadata: { user_id: "u_sub" },
		items: {
			object: "list",
			data: [{ current_period_end: periodEnd, price: { lookup_key: STRIPE_LOOKUP_KEYS.business } }],
		},
	} as unknown as Stripe.Subscription;
}

function deps(over: Partial<StripeWebhookDeps> = {}): StripeWebhookDeps {
	return {
		stripe,
		secret,
		db,
		store: memoryStore(),
		retrieveSubscription: async () => subscriptionFixture("active", 1_900_000_000),
		...over,
	};
}

const creditsSession = (eventId: string) =>
	signed({
		id: eventId,
		type: "checkout.session.completed",
		data: {
			object: {
				id: "cs_test_credits",
				object: "checkout.session",
				mode: "payment",
				payment_status: "paid",
				customer: "cus_credits",
				client_reference_id: "u_credits",
				metadata: { user_id: "u_credits", kind: "credits" },
			},
		},
	});

describe("handleStripeWebhook", () => {
	test("rejects a missing or invalid signature with 400", async () => {
		const { body } = await creditsSession("evt_bad");
		expect((await handleStripeWebhook(body, null, deps())).status).toBe(400);
		const forged = await stripe.webhooks.generateTestHeaderStringAsync({ payload: body, secret: "whsec_other" });
		expect((await handleStripeWebhook(body, forged, deps())).status).toBe(400);
		const { signature } = await creditsSession("evt_bad");
		expect((await handleStripeWebhook(body.replace("u_credits", "u_sub"), signature, deps())).status).toBe(400);
	});

	test("grants a credit pack once, even when the event or the session is delivered again", async () => {
		const store = memoryStore();
		await store.setIfAbsent("hx:credits:u_credits", 10);
		const first = await creditsSession("evt_credits_1");
		expect((await handleStripeWebhook(first.body, first.signature, deps({ store }))).status).toBe(200);

		const again = await handleStripeWebhook(first.body, first.signature, deps({ store }));
		expect(again).toMatchObject({ status: 200, body: { duplicate: true } });

		// Same Checkout Session under a new event id (e.g. async_payment_succeeded): the ledger ref dedupes it.
		const other = await creditsSession("evt_credits_2");
		expect((await handleStripeWebhook(other.body, other.signature, deps({ store }))).status).toBe(200);

		const rows = await db.select().from(creditLedger).where(eq(creditLedger.userId, "u_credits"));
		expect(rows).toHaveLength(1);
		expect(rows[0]).toMatchObject({ delta: 25_000, ref: "cs_test_credits", reason: "purchase" });
		expect(await store.get("hx:credits:u_credits")).toBe(25_010);
	});

	test("does not create the Redis balance when it is not cached", async () => {
		const store = memoryStore();
		await db.delete(creditLedger);
		const e = await creditsSession("evt_credits_3");
		await handleStripeWebhook(e.body, e.signature, deps({ store }));
		expect(await store.get("hx:credits:u_credits")).toBeNull();
	});

	test("upserts the subscription from the retrieved Stripe state", async () => {
		const e = await signed({
			id: "evt_sub_1",
			type: "customer.subscription.updated",
			data: { object: subscriptionFixture("incomplete", 1) },
		});
		expect((await handleStripeWebhook(e.body, e.signature, deps())).status).toBe(200);
		const [row] = await db.select().from(subscriptions).where(eq(subscriptions.userId, "u_sub"));
		expect(row).toMatchObject({
			plan: "business",
			provider: "stripe",
			status: "active",
			providerRef: "sub_1",
			providerCustomerId: "cus_1",
		});
		expect(row?.currentPeriodEnd.getTime()).toBe(1_900_000_000_000);

		const canceled = await signed({ id: "evt_sub_2", type: "customer.subscription.deleted", data: { object: {} } });
		const retrieveCanceled = async () => subscriptionFixture("canceled", 1_900_000_000);
		const res = await handleStripeWebhook(
			canceled.body,
			canceled.signature,
			deps({ retrieveSubscription: retrieveCanceled }),
		);
		expect(res.status).toBe(200);
		const [after] = await db.select().from(subscriptions).where(eq(subscriptions.userId, "u_sub"));
		expect(after?.status).toBe("canceled");
	});

	test("returns 500 so Stripe retries when processing fails, and does not mark the event as processed", async () => {
		const failing = deps({
			retrieveSubscription: async () => {
				throw new Error("stripe down");
			},
		});
		const e = await signed({
			id: "evt_fail",
			type: "invoice.paid",
			data: { object: { id: "in_1", parent: { subscription_details: { subscription: "sub_1" } } } },
		});
		expect((await handleStripeWebhook(e.body, e.signature, failing)).status).toBe(500);
		expect((await handleStripeWebhook(e.body, e.signature, deps())).body).not.toHaveProperty("duplicate");
	});

	test("acknowledges unhandled event types", async () => {
		const e = await signed({ id: "evt_other", type: "charge.refunded", data: { object: {} } });
		expect(await handleStripeWebhook(e.body, e.signature, deps())).toMatchObject({ status: 200 });
	});
});
