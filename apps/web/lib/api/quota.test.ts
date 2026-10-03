import { describe, expect, test } from "bun:test";
import { consumeRequest, memoryStore, refundRequest } from "./quota";

const noCredits = async (): Promise<number> => 0;

describe("consumeRequest", () => {
	test("free plan allows exactly 100 lifetime requests, then 402", async () => {
		const s = memoryStore();
		s.throttle = async () => 0;
		for (let i = 0; i < 100; i++) expect((await consumeRequest(s, "u1", "free", noCredits)).ok).toBe(true);
		const denied = await consumeRequest(s, "u1", "free", noCredits);
		expect(denied).toMatchObject({ ok: false, reason: "quota_exhausted", limit: 100 });
	});

	test("free plan throttles at 1 request per second", async () => {
		const s = memoryStore();
		expect((await consumeRequest(s, "u2", "free", noCredits)).ok).toBe(true);
		expect(await consumeRequest(s, "u2", "free", noCredits)).toMatchObject({ ok: false, reason: "rate_limited" });
	});

	test("paid plan spends credits after the monthly quota and stops at zero", async () => {
		const s = memoryStore();
		s.throttle = async () => 0;
		const now = new Date("2026-10-15T00:00:00Z");
		await s.setIfAbsent("hx:q:u3:202610", 50_000);
		const first = await consumeRequest(s, "u3", "pro", async () => 2, now);
		expect(first).toMatchObject({ ok: true, usedCredit: true });
		expect((await consumeRequest(s, "u3", "pro", async () => 2, now)).ok).toBe(true);
		expect(await consumeRequest(s, "u3", "pro", async () => 2, now)).toMatchObject({
			ok: false,
			reason: "quota_exhausted",
		});
	});

	test("free plan never spends credits", async () => {
		const s = memoryStore();
		s.throttle = async () => 0;
		await s.setIfAbsent("hx:q:u4:life", 100);
		expect(await consumeRequest(s, "u4", "free", async () => 999)).toMatchObject({ ok: false });
	});

	test("refund gives the request back", async () => {
		const s = memoryStore();
		s.throttle = async () => 0;
		await consumeRequest(s, "u5", "free", noCredits);
		await refundRequest(s, "u5", "free", false);
		expect(await s.get("hx:q:u5:life")).toBe(0);
	});
});
