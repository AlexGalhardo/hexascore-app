import { and, count, eq, sql } from "drizzle-orm";
import type { AnyPgColumn } from "drizzle-orm/pg-core";
import { z } from "zod";
import type { TimeToBeat } from "./catalog/types";
import type { Database } from "./db/client";
import { items } from "./db/schema/catalog";
import { itemStatus, timeToBeat } from "./db/schema/social";
import { MIN_REPORTS_FOR_TIME_TO_BEAT, TIME_TO_BEAT_HOURS, type TimeToBeatStyle } from "./social/constants";

// Crowdsourced time to beat for games. Only a user whose item status is `completed` can report; leaving that status
// deletes the report (lib/social/repository.ts#setItemStatus). Medians are computed on read, never stored.

const hours = z
	.number()
	.min(TIME_TO_BEAT_HOURS.min)
	.max(TIME_TO_BEAT_HOURS.max)
	.multipleOf(TIME_TO_BEAT_HOURS.step)
	.nullable();

export const timeToBeatInputSchema = z
	.object({ itemId: z.uuid(), main: hours, extras: hours, completionist: hours })
	.refine((v) => v.main !== null || v.extras !== null || v.completionist !== null, { message: "empty" });
export type TimeToBeatInput = z.infer<typeof timeToBeatInputSchema>;
export type UserTimeToBeat = Record<TimeToBeatStyle, number | null>;

const toMinutes = (h: number | null): number | null => (h === null ? null : Math.round(h * 60));
const toHours = (m: number | null): number | null => (m === null ? null : m / 60);

export type TimeToBeatWrite = "ok" | "notFound" | "forbidden";

/** Upserts the user's report. "forbidden" = not a game or the user's status is not `completed`. */
export function writeTimeToBeat(db: Database, userId: string, input: TimeToBeatInput): Promise<TimeToBeatWrite> {
	return db.transaction(async (tx) => {
		const [item] = await tx.select({ type: items.type }).from(items).where(eq(items.id, input.itemId)).limit(1);
		if (!item) return "notFound";
		// Locks the status row, so a concurrent status change (which deletes the report) waits for this write.
		const [status] = await tx
			.select({ status: itemStatus.status })
			.from(itemStatus)
			.where(and(eq(itemStatus.userId, userId), eq(itemStatus.itemId, input.itemId)))
			.for("update");
		if (item.type !== "game" || status?.status !== "completed") return "forbidden";
		const values = {
			mainMinutes: toMinutes(input.main),
			extrasMinutes: toMinutes(input.extras),
			completionistMinutes: toMinutes(input.completionist),
		};
		await tx
			.insert(timeToBeat)
			.values({ userId, itemId: input.itemId, ...values })
			.onConflictDoUpdate({
				target: [timeToBeat.userId, timeToBeat.itemId],
				set: { ...values, updatedAt: new Date() },
			});
		return "ok";
	});
}

/** No-op when the user has no report. */
export async function removeTimeToBeat(db: Database, userId: string, itemId: string): Promise<void> {
	await db.delete(timeToBeat).where(and(eq(timeToBeat.userId, userId), eq(timeToBeat.itemId, itemId)));
}

export async function getUserTimeToBeat(db: Database, userId: string, itemId: string): Promise<UserTimeToBeat | null> {
	const [row] = await db
		.select({
			main: timeToBeat.mainMinutes,
			extras: timeToBeat.extrasMinutes,
			completionist: timeToBeat.completionistMinutes,
		})
		.from(timeToBeat)
		.where(and(eq(timeToBeat.userId, userId), eq(timeToBeat.itemId, itemId)))
		.limit(1);
	return row
		? { main: toHours(row.main), extras: toHours(row.extras), completionist: toHours(row.completionist) }
		: null;
}

// percentile_cont ignores nulls, so each style's median and count only see the users who reported that style.
const median = (column: AnyPgColumn) =>
	sql<number | string | null>`percentile_cont(0.5) within group (order by ${column})`;

/** Median hours (1 decimal) and report count per style; the median is null below MIN_REPORTS_FOR_TIME_TO_BEAT. */
export async function getTimeToBeat(db: Database, itemId: string): Promise<TimeToBeat> {
	const [row] = await db
		.select({
			main: median(timeToBeat.mainMinutes),
			mainN: count(timeToBeat.mainMinutes),
			extras: median(timeToBeat.extrasMinutes),
			extrasN: count(timeToBeat.extrasMinutes),
			completionist: median(timeToBeat.completionistMinutes),
			completionistN: count(timeToBeat.completionistMinutes),
		})
		.from(timeToBeat)
		.where(eq(timeToBeat.itemId, itemId));
	const style = (minutes: number | string | null | undefined, reports = 0) => ({
		medianHours:
			reports >= MIN_REPORTS_FOR_TIME_TO_BEAT && minutes !== null && minutes !== undefined
				? Math.round(Number(minutes) / 6) / 10
				: null,
		reports,
	});
	return {
		main: style(row?.main, row?.mainN),
		extras: style(row?.extras, row?.extrasN),
		completionist: style(row?.completionist, row?.completionistN),
	};
}
