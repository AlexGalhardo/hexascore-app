import { expect, test } from "bun:test";

import { evaluate, parseState, projectedUsd, shouldAlert } from "./projected-spend.ts";

test("day 20 with US$13 in a 30-day cycle projects US$19.5 and alerts", () => {
	expect(projectedUsd(13, 20, 30)).toBe(19.5);
	expect(shouldAlert(13, 20, 30)).toBe(true);
});

test("day 20 with US$11 in a 30-day cycle does not alert", () => {
	expect(projectedUsd(11, 20, 30)).toBe(16.5);
	expect(shouldAlert(11, 20, 30)).toBe(false);
});

test("a projection of exactly US$18 does not alert", () => {
	expect(projectedUsd(12, 20, 30)).toBe(18);
	expect(shouldAlert(12, 20, 30)).toBe(false);
});

test("a projection above US$18 before day 20 does not alert", () => {
	expect(shouldAlert(13, 19, 30)).toBe(false);
});

const start = "2026-10-17T12:47:54.000Z";
const end = "2026-11-16T12:47:54.000Z";
const day = (n: number): Date => new Date(Date.parse(start) + n * 86_400_000);
const usage = (usd: number) => ({ usd, start, end });

test("first run above US$5 with no state sends the US$5 threshold", () => {
	const r = evaluate(usage(7.57), parseState(""), day(15));
	expect(r.messages).toEqual(["HexaScore A17: Railway usage US$7.57 crossed US$5 (hard limit US$20)."]);
	expect(r.state).toEqual({ period: start, level: 5, projected: false });
});

test("the same threshold is not sent twice in a period", () => {
	const r = evaluate(usage(9.9), { period: start, level: 5, projected: false }, day(16));
	expect(r.messages).toEqual([]);
	expect(r.state.level).toBe(5);
});

test("jumping over several thresholds sends only the highest", () => {
	const r = evaluate(usage(16), { period: start, level: 5, projected: false }, day(10));
	expect(r.messages).toEqual(["HexaScore A17: Railway usage US$16.00 crossed US$15 (hard limit US$20)."]);
	expect(r.state.level).toBe(15);
});

test("reaching US$20 says the hard limit is hit", () => {
	const r = evaluate(usage(20), { period: start, level: 15, projected: false }, day(25));
	expect(r.messages[0]).toBe(
		"HexaScore A17: Railway usage US$20.00 reached the US$20 hard limit; Railway stops the services.",
	);
});

test("a new billing period resets the state", () => {
	const r = evaluate(usage(5.01), { period: "2026-09-17T12:47:54.000Z", level: 20, projected: true }, day(3));
	expect(r.state).toEqual({ period: start, level: 5, projected: false });
	expect(r.messages).toHaveLength(1);
});

test("under US$5 sends nothing", () => {
	const r = evaluate(usage(4.99), parseState(""), day(2));
	expect(r.messages).toEqual([]);
	expect(r.state.level).toBe(0);
});

test("the projection alert uses the billing period and is sent once", () => {
	// 30-day period, day 20, US$13 -> US$19.5
	const first = evaluate(usage(13), { period: start, level: 10, projected: false }, day(20));
	expect(first.messages).toEqual(["HexaScore A17: projected Railway spend US$19.50 for this period is above US$18."]);
	expect(first.state.projected).toBe(true);
	const again = evaluate(usage(13.2), first.state, day(20.5));
	expect(again.messages).toEqual([]);
});

test("corrupt state is treated as empty", () => {
	expect(parseState("not json")).toEqual({ period: "", level: 0, projected: false });
});
