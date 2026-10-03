import { and, avg, count, eq } from "drizzle-orm";
import { z } from "zod";
import type { Database } from "./db/client";
import { items } from "./db/schema/catalog";
import { ratings } from "./db/schema/social";

export const scoresSchema = z.array(z.number().int().min(0).max(100)).length(6);
export const ratingInputSchema = z.object({ itemId: z.uuid(), scores: scoresSchema });
export type RatingInput = z.infer<typeof ratingInputSchema>;
export type Scores = z.infer<typeof scoresSchema>;

type Tx = Parameters<Parameters<Database["transaction"]>[0]>[0];

const roundAvg = (value: string | null): number | null => (value === null ? null : Math.round(Number(value)));

// Row lock on the item serializes every writer of its stats: two concurrent ratings can no longer both read the
// old aggregate and overwrite each other (lost update). Returns false when the item does not exist.
async function lockItem(tx: Tx, itemId: string): Promise<boolean> {
	const rows = await tx.select({ id: items.id }).from(items).where(eq(items.id, itemId)).for("update");
	return rows.length > 0;
}

/**
 * Recomputes the denormalized stats from `ratings` (never incrementally, so they self-heal). Stores raw rounded
 * averages for any count; the read side (`lib/catalog/repository.ts`) hides them below MIN_RATINGS_FOR_SCORE.
 * Must run inside the transaction that changed the ratings, after lockItem().
 */
export async function recomputeItemStats(tx: Tx, itemId: string): Promise<void> {
	const [s] = await tx
		.select({
			n: count(),
			score: avg(ratings.overall),
			c1: avg(ratings.c1),
			c2: avg(ratings.c2),
			c3: avg(ratings.c3),
			c4: avg(ratings.c4),
			c5: avg(ratings.c5),
			c6: avg(ratings.c6),
		})
		.from(ratings)
		.where(eq(ratings.itemId, itemId));
	await tx
		.update(items)
		.set({
			ratingCount: s?.n ?? 0,
			score: roundAvg(s?.score ?? null),
			scoreC1: roundAvg(s?.c1 ?? null),
			scoreC2: roundAvg(s?.c2 ?? null),
			scoreC3: roundAvg(s?.c3 ?? null),
			scoreC4: roundAvg(s?.c4 ?? null),
			scoreC5: roundAvg(s?.c5 ?? null),
			scoreC6: roundAvg(s?.c6 ?? null),
		})
		.where(eq(items.id, itemId));
}

/** Insert or edit the user's rating and refresh the item stats atomically. False = unknown item. */
export function writeRating(db: Database, userId: string, input: RatingInput): Promise<boolean> {
	const [c1, c2, c3, c4, c5, c6] = input.scores as [number, number, number, number, number, number];
	return db.transaction(async (tx) => {
		if (!(await lockItem(tx, input.itemId))) return false;
		await tx
			.insert(ratings)
			.values({ userId, itemId: input.itemId, c1, c2, c3, c4, c5, c6 })
			.onConflictDoUpdate({
				target: [ratings.userId, ratings.itemId],
				set: { c1, c2, c3, c4, c5, c6, updatedAt: new Date() },
			});
		await recomputeItemStats(tx, input.itemId);
		return true;
	});
}

/** Deletes the user's rating (no-op if none) and refreshes the stats. False = unknown item. */
export function removeRating(db: Database, userId: string, itemId: string): Promise<boolean> {
	return db.transaction(async (tx) => {
		if (!(await lockItem(tx, itemId))) return false;
		await tx.delete(ratings).where(and(eq(ratings.userId, userId), eq(ratings.itemId, itemId)));
		await recomputeItemStats(tx, itemId);
		return true;
	});
}

export async function getUserRating(db: Database, userId: string, itemId: string): Promise<Scores | null> {
	const [row] = await db
		.select({ c1: ratings.c1, c2: ratings.c2, c3: ratings.c3, c4: ratings.c4, c5: ratings.c5, c6: ratings.c6 })
		.from(ratings)
		.where(and(eq(ratings.userId, userId), eq(ratings.itemId, itemId)))
		.limit(1);
	return row ? [row.c1, row.c2, row.c3, row.c4, row.c5, row.c6] : null;
}
