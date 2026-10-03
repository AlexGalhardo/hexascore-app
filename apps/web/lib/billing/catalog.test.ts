import { describe, expect, test } from "bun:test";
import {
	addMonths,
	CREDIT_PACK_PRICE_CENTS,
	currencyForLocale,
	extendPeriod,
	MONTHLY_PRICE_CENTS,
	PIX_PURPOSES,
	pixOffer,
	planFromLookupKey,
	STRIPE_LOOKUP_KEYS,
} from "./catalog";
import { pixRefusal } from "./pix";

describe("catalog", () => {
	test("monthly prices match PLAN.md", () => {
		expect(MONTHLY_PRICE_CENTS.pro).toEqual({ brl: 2900, usd: 600, eur: 600 });
		expect(MONTHLY_PRICE_CENTS.business).toEqual({ brl: 14900, usd: 2900, eur: 2700 });
		expect(CREDIT_PACK_PRICE_CENTS).toEqual({ brl: 1900, usd: 400, eur: 400 });
	});

	test("PIX annual costs 10x the monthly BRL price for 12 months", () => {
		expect(pixOffer("pro_12m")).toEqual({ kind: "plan", plan: "pro", months: 12, amountCents: 29_000 });
		expect(pixOffer("business_12m")).toEqual({ kind: "plan", plan: "business", months: 12, amountCents: 149_000 });
		expect(pixOffer("pro_1m")).toEqual({ kind: "plan", plan: "pro", months: 1, amountCents: 2900 });
		expect(pixOffer("credits")).toEqual({ kind: "credits", requests: 25_000, amountCents: 1900 });
		for (const p of PIX_PURPOSES) expect(pixOffer(p).amountCents).toBeGreaterThan(0);
	});

	test("currency follows the locale", () => {
		expect(currencyForLocale("pt-BR")).toBe("brl");
		expect(currencyForLocale("es")).toBe("eur");
		expect(currencyForLocale("en")).toBe("usd");
		expect(currencyForLocale("xx")).toBe("usd");
	});

	test("lookup keys map back to plans", () => {
		expect(planFromLookupKey(STRIPE_LOOKUP_KEYS.pro)).toBe("pro");
		expect(planFromLookupKey(STRIPE_LOOKUP_KEYS.business)).toBe("business");
		expect(planFromLookupKey(STRIPE_LOOKUP_KEYS.credits)).toBeNull();
		expect(planFromLookupKey(null)).toBeNull();
	});
});

describe("PIX period math", () => {
	const now = new Date("2026-10-15T12:00:00Z");

	test("addMonths clamps to the end of shorter months", () => {
		expect(addMonths(new Date("2026-01-31T00:00:00Z"), 1).toISOString()).toBe("2026-02-28T00:00:00.000Z");
		expect(addMonths(new Date("2028-01-31T00:00:00Z"), 1).toISOString()).toBe("2028-02-29T00:00:00.000Z");
		expect(addMonths(new Date("2026-10-15T12:00:00Z"), 12).toISOString()).toBe("2027-10-15T12:00:00.000Z");
	});

	test("no current period starts from now", () => {
		expect(extendPeriod(null, now, 1).toISOString()).toBe("2026-11-15T12:00:00.000Z");
	});

	test("an expired period starts from now, not from the old end", () => {
		expect(extendPeriod(new Date("2026-09-01T00:00:00Z"), now, 1).toISOString()).toBe("2026-11-15T12:00:00.000Z");
	});

	test("an active period stacks on its end", () => {
		expect(extendPeriod(new Date("2026-10-20T00:00:00Z"), now, 12).toISOString()).toBe("2027-10-20T00:00:00.000Z");
	});
});

describe("pixRefusal", () => {
	const now = new Date("2026-10-15T00:00:00Z");
	const future = new Date("2026-11-01T00:00:00Z");
	const pixPro = { provider: "abacatepay", plan: "pro", status: "active", currentPeriodEnd: future } as const;

	test("anyone can start a plan; credits need an active paid plan", () => {
		expect(pixRefusal("pro_1m", null, now)).toBeNull();
		expect(pixRefusal("credits", null, now)).toBe("plan_required");
		expect(pixRefusal("credits", pixPro, now)).toBeNull();
	});

	test("an active Stripe subscription blocks PIX plans but not credits", () => {
		const stripePro = { ...pixPro, provider: "stripe" } as const;
		expect(pixRefusal("pro_12m", stripePro, now)).toBe("stripe_active");
		expect(pixRefusal("credits", stripePro, now)).toBeNull();
	});

	test("PIX renews the same plan only", () => {
		expect(pixRefusal("pro_12m", pixPro, now)).toBeNull();
		expect(pixRefusal("business_1m", pixPro, now)).toBe("plan_conflict");
		expect(
			pixRefusal("business_1m", { ...pixPro, currentPeriodEnd: new Date("2026-10-01T00:00:00Z") }, now),
		).toBeNull();
	});
});
