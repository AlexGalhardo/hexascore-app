// A17 Railway spend alerts (SDD 12 §7.5). No new dependency.
// Reads the billing-period usage from the Railway public API (the query the CLI's `railway usage` uses),
// alerts once per period at each threshold and once for the day-20 projection.

import { appendFileSync, existsSync, readFileSync, writeFileSync } from "node:fs";

// Below the US$20 hard limit, so the owner hears of a projected overrun while there is time to act.
export const alertThresholdUsd = 18;
export const alertFromDay = 20;
export const hardLimitUsd = 20;
export const thresholdsUsd = [5, 10, 15, hardLimitUsd];

const dayMs = 86_400_000;

export function projectedUsd(monthToDateUsd: number, elapsedDays: number, cycleDays: number): number {
	if (elapsedDays <= 0 || cycleDays <= 0) {
		throw new Error("elapsedDays and cycleDays must be positive");
	}

	return (monthToDateUsd * cycleDays) / elapsedDays;
}

// Alerts from day 20 onward when the projection is above alertThresholdUsd; the threshold itself does not alert.
export function shouldAlert(monthToDateUsd: number, elapsedDays: number, cycleDays: number): boolean {
	if (elapsedDays < alertFromDay) {
		return false;
	}

	return projectedUsd(monthToDateUsd, elapsedDays, cycleDays) > alertThresholdUsd;
}

export type Usage = { usd: number; start: string; end: string };
// level = highest threshold already sent in `period`; projected = projection alert already sent.
export type State = { period: string; level: number; projected: boolean };

export function parseState(raw: string): State {
	try {
		const v: unknown = JSON.parse(raw);
		if (
			typeof v === "object" &&
			v !== null &&
			"period" in v &&
			typeof v.period === "string" &&
			"level" in v &&
			typeof v.level === "number" &&
			"projected" in v &&
			typeof v.projected === "boolean"
		) {
			return { period: v.period, level: v.level, projected: v.projected };
		}
	} catch {
		// fall through: a lost or corrupt state at worst resends the current threshold once
	}

	return { period: "", level: 0, projected: false };
}

export function evaluate(usage: Usage, prev: State, now: Date): { messages: string[]; state: State } {
	const state: State = prev.period === usage.start ? { ...prev } : { period: usage.start, level: 0, projected: false };
	const messages: string[] = [];
	const usd = usage.usd.toFixed(2);

	const crossed = Math.max(0, ...thresholdsUsd.filter((t) => usage.usd >= t));
	if (crossed > state.level) {
		messages.push(
			crossed === hardLimitUsd
				? `HexaScore A17: Railway usage US$${usd} reached the US$${hardLimitUsd} hard limit; Railway stops the services.`
				: `HexaScore A17: Railway usage US$${usd} crossed US$${crossed} (hard limit US$${hardLimitUsd}).`,
		);
		state.level = crossed;
	}

	const start = Date.parse(usage.start);
	const elapsed = (now.getTime() - start) / dayMs;
	const cycle = (Date.parse(usage.end) - start) / dayMs;
	if (!state.projected && elapsed > 0 && shouldAlert(usage.usd, elapsed, cycle)) {
		const projected = projectedUsd(usage.usd, elapsed, cycle).toFixed(2);
		messages.push(`HexaScore A17: projected Railway spend US$${projected} for this period is above US$${alertThresholdUsd}.`);
		state.projected = true;
	}

	return { messages, state };
}

// Documented in the Railway CLI source (src/commands/usage.rs, WorkspaceUsageContext) and verified with `railway api`.
const usageQuery = `query($w: String!) { workspace(workspaceId: $w) { customer { currentUsage billingPeriod { start end } } } }`;

async function fetchUsage(token: string, workspaceId: string): Promise<Usage> {
	const res = await fetch("https://backboard.railway.com/graphql/v2", {
		method: "POST",
		headers: { "content-type": "application/json", authorization: `Bearer ${token}` },
		body: JSON.stringify({ query: usageQuery, variables: { w: workspaceId } }),
	});
	if (!res.ok) {
		throw new Error(`railway api: HTTP ${res.status}`);
	}

	const body = (await res.json()) as {
		errors?: { message: string }[];
		data?: { workspace?: { customer?: { currentUsage?: unknown; billingPeriod?: { start?: unknown; end?: unknown } } } };
	};
	const c = body.data?.workspace?.customer;
	const usd = c?.currentUsage;
	const start = c?.billingPeriod?.start;
	const end = c?.billingPeriod?.end;
	if (typeof usd !== "number" || typeof start !== "string" || typeof end !== "string") {
		throw new Error(`railway api: unexpected response ${body.errors?.map((e) => e.message).join("; ") ?? ""}`);
	}

	return { usd, start, end };
}

async function main(): Promise<void> {
	const token = process.env.RAILWAY_API_TOKEN ?? "";
	const workspaceId = process.env.RAILWAY_WORKSPACE_ID ?? "";
	const statePath = process.env.STATE_FILE ?? ".spend-state/state.json";
	if (token === "" || workspaceId === "") {
		console.log("::warning::RAILWAY_API_TOKEN or RAILWAY_WORKSPACE_ID unset; no usage read");

		return;
	}

	const usage = await fetchUsage(token, workspaceId);
	const prev = parseState(existsSync(statePath) ? readFileSync(statePath, "utf8") : "");
	const { messages, state } = evaluate(usage, prev, new Date());
	writeFileSync(statePath, JSON.stringify(state));
	console.log(`usage=${usage.usd.toFixed(2)} period=${usage.start} level=${state.level} projected=${state.projected}`);

	const out = process.env.GITHUB_OUTPUT;
	if (out && messages.length > 0) {
		appendFileSync(out, `message=${messages.join(" ")}\n`);
	}
}

if (import.meta.main) {
	await main();
}
