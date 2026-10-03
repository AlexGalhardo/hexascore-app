import { boolean, date, index, integer, pgEnum, pgTable, primaryKey, text, timestamp, uuid } from "drizzle-orm/pg-core";
import { user } from "./auth";

export const plan = pgEnum("plan", ["free", "pro", "business"]);
export const paymentProvider = pgEnum("payment_provider", ["stripe", "abacatepay"]);
export const subscriptionStatus = pgEnum("subscription_status", ["active", "past_due", "canceled", "expired"]);

export const apiKeys = pgTable(
	"api_keys",
	{
		id: uuid()
			.primaryKey()
			.$defaultFn(() => Bun.randomUUIDv7()),
		userId: text()
			.notNull()
			.references(() => user.id, { onDelete: "cascade" }),
		name: text().notNull(),
		prefix: text().notNull(),
		hash: text().notNull().unique(),
		lastUsedAt: timestamp({ withTimezone: true }),
		revokedAt: timestamp({ withTimezone: true }),
		createdAt: timestamp({ withTimezone: true }).notNull().defaultNow(),
	},
	(t) => [index("api_keys_user_idx").on(t.userId)],
);

export const subscriptions = pgTable("subscriptions", {
	userId: text()
		.primaryKey()
		.references(() => user.id, { onDelete: "cascade" }),
	plan: plan().notNull(),
	provider: paymentProvider().notNull(),
	status: subscriptionStatus().notNull(),
	currentPeriodEnd: timestamp({ withTimezone: true }).notNull(),
	cancelAtPeriodEnd: boolean().notNull().default(false),
	providerCustomerId: text(),
	providerRef: text(),
	updatedAt: timestamp({ withTimezone: true })
		.notNull()
		.defaultNow()
		.$onUpdate(() => new Date()),
});

export const creditLedger = pgTable(
	"credit_ledger",
	{
		id: uuid()
			.primaryKey()
			.$defaultFn(() => Bun.randomUUIDv7()),
		userId: text()
			.notNull()
			.references(() => user.id, { onDelete: "cascade" }),
		delta: integer().notNull(),
		reason: text().notNull(),
		ref: text().unique(),
		createdAt: timestamp({ withTimezone: true }).notNull().defaultNow(),
	},
	(t) => [index("credit_ledger_user_idx").on(t.userId)],
);

export const pixChargeStatus = pgEnum("pix_charge_status", ["pending", "paid", "expired", "failed"]);

export const pixCharges = pgTable("pix_charges", {
	id: uuid()
		.primaryKey()
		.$defaultFn(() => Bun.randomUUIDv7()),
	userId: text()
		.notNull()
		.references(() => user.id, { onDelete: "cascade" }),
	providerRef: text().notNull().unique(),
	purpose: text().notNull(),
	amountCents: integer().notNull(),
	status: pixChargeStatus().notNull().default("pending"),
	brCode: text(),
	brCodeBase64: text(),
	expiresAt: timestamp({ withTimezone: true }),
	paidAt: timestamp({ withTimezone: true }),
	createdAt: timestamp({ withTimezone: true }).notNull().defaultNow(),
});

export const apiUsageDaily = pgTable(
	"api_usage_daily",
	{
		userId: text()
			.notNull()
			.references(() => user.id, { onDelete: "cascade" }),
		day: date().notNull(),
		count: integer().notNull(),
	},
	(t) => [primaryKey({ columns: [t.userId, t.day] })],
);

export const processedEvents = pgTable("processed_events", {
	id: text().primaryKey(),
	provider: paymentProvider().notNull(),
	processedAt: timestamp({ withTimezone: true }).notNull().defaultNow(),
});

/** One Stripe Customer per user, reused by every Checkout (subscriptions and credit packs). */
export const stripeCustomers = pgTable("stripe_customers", {
	userId: text()
		.primaryKey()
		.references(() => user.id, { onDelete: "cascade" }),
	customerId: text().notNull().unique(),
	createdAt: timestamp({ withTimezone: true }).notNull().defaultNow(),
});
