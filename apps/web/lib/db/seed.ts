import { parseArgs } from "node:util";
import { eq, type SQL, sql } from "drizzle-orm";
import { type AnyPgColumn, alias } from "drizzle-orm/pg-core";
import { z } from "zod";
import { ITEM_TYPES } from "./schema/catalog";
import { generateItem, generateRating, generateUser, raterIndexes, type SeedItem, seedUserId } from "./seed-data";

// Dev/preview fixture catalog: `bun run db:seed [--users 50] [--items 400] [--batch 250]`.
// Idempotent (upserts keyed by deterministic ids) and streaming: items are generated and written batch by batch,
// so the same script scales to 10k users / 100k items without holding the catalog in memory.

const argsSchema = z.object({
	users: z.coerce.number().int().min(0).max(1_000_000).default(50),
	items: z.coerce.number().int().min(4).max(1_000_000).default(400),
	batch: z.coerce.number().int().min(1).max(5_000).default(250),
});

async function main(): Promise<void> {
	if (process.env.VERCEL_ENV === "production") {
		throw new Error("Refusing to seed: VERCEL_ENV=production (fake users/ratings would distort real metascores).");
	}
	const { values } = parseArgs({
		options: { users: { type: "string" }, items: { type: "string" }, batch: { type: "string" } },
	});
	const opts = argsSchema.parse(values);

	// Imported lazily so the production guard above runs before any connection is opened.
	const { db } = await import("./client");
	const schema = await import("./schema");
	const started = performance.now();

	for (let start = 0; start < opts.users; start += opts.batch) {
		const rows = Array.from({ length: Math.min(opts.batch, opts.users - start) }, (_, i) => ({
			...generateUser(start + i),
			emailVerified: true,
		}));
		await db
			.insert(schema.user)
			.values(rows)
			.onConflictDoUpdate({
				target: schema.user.id,
				set: {
					name: sql`excluded.name`,
					username: sql`excluded.username`,
					profilePrivate: sql`excluded.profile_private`,
				},
			});
	}
	console.info(`[seed] ${opts.users} users`);

	const perType = Math.floor(opts.items / ITEM_TYPES.length);
	let ratingTotal = 0;
	let ordinal = 0;
	for (const type of ITEM_TYPES) {
		const usedSlugs = new Set<string>();
		for (let start = 0; start < perType; start += opts.batch) {
			const batch: SeedItem[] = [];
			for (let i = start; i < Math.min(perType, start + opts.batch); i++)
				batch.push(generateItem(type, i, opts.users));
			const slugs = batch.map((item) => uniqueSlug(item, usedSlugs));

			await db.transaction(async (tx) => {
				const inserted = await tx
					.insert(schema.items)
					.values(
						batch.map((item, i) => ({
							type: item.type,
							slug: slugs[i] as string,
							wikidataQid: item.wikidataQid,
							releaseDate: item.releaseDate,
							popularity: item.popularity,
							externalIds: item.externalIds,
							meta: item.meta,
						})),
					)
					.onConflictDoUpdate({
						target: schema.items.wikidataQid,
						set: {
							slug: sql`excluded.slug`,
							releaseDate: sql`excluded.release_date`,
							popularity: sql`excluded.popularity`,
							externalIds: sql`excluded.external_ids`,
							meta: sql`excluded.meta`,
						},
					})
					.returning({ id: schema.items.id, qid: schema.items.wikidataQid });
				const idByQid = new Map(inserted.map((r) => [r.qid, r.id]));
				const idOf = (item: SeedItem) => idByQid.get(item.wikidataQid) as string;

				await tx
					.insert(schema.itemTranslations)
					.values(batch.flatMap((item) => item.translations.map((t) => ({ ...t, itemId: idOf(item) }))))
					.onConflictDoUpdate({
						target: [schema.itemTranslations.itemId, schema.itemTranslations.locale],
						set: {
							title: sql`excluded.title`,
							synopsis: sql`excluded.synopsis`,
							synopsisLicense: sql`excluded.synopsis_license`,
							synopsisSourceUrl: sql`excluded.synopsis_source_url`,
						},
					});

				const scores = batch.flatMap((item) =>
					item.externalScores.map((s) => ({
						itemId: idOf(item),
						source: s.source as (typeof schema.externalScoreSource.enumValues)[number],
						value: s.value,
					})),
				);
				if (scores.length) {
					await tx
						.insert(schema.externalScores)
						.values(scores)
						.onConflictDoUpdate({
							target: [schema.externalScores.itemId, schema.externalScores.source],
							set: { value: sql`excluded.value` },
						});
				}

				const links = batch.flatMap((item) => item.affiliateLinks.map((l) => ({ ...l, itemId: idOf(item) })));
				if (links.length) {
					await tx
						.insert(schema.affiliateLinks)
						.values(links)
						.onConflictDoUpdate({
							target: [schema.affiliateLinks.itemId, schema.affiliateLinks.store],
							set: { url: sql`excluded.url` },
						});
				}

				const ratingRows = batch.flatMap((item, i) => {
					const itemOrdinal = ordinal + i;
					return raterIndexes(itemOrdinal, item.ratingCount, opts.users).map((u, k) => {
						const [c1, c2, c3, c4, c5, c6] = generateRating(item, u) as [
							number,
							number,
							number,
							number,
							number,
							number,
						];
						return {
							userId: seedUserId(u),
							itemId: idOf(item),
							c1,
							c2,
							c3,
							c4,
							c5,
							c6,
							createdAt: new Date(Date.UTC(2026, 0, 1) - (itemOrdinal * 31 + k) * 3_600_000),
						};
					});
				});
				for (let r = 0; r < ratingRows.length; r += 1000) {
					await tx
						.insert(schema.ratings)
						.values(ratingRows.slice(r, r + 1000))
						.onConflictDoUpdate({
							target: [schema.ratings.userId, schema.ratings.itemId],
							set: {
								c1: sql`excluded.c1`,
								c2: sql`excluded.c2`,
								c3: sql`excluded.c3`,
								c4: sql`excluded.c4`,
								c5: sql`excluded.c5`,
								c6: sql`excluded.c6`,
								createdAt: sql`excluded.created_at`,
							},
						});
				}
				ratingTotal += ratingRows.length;
			});
			ordinal += batch.length;
		}
		console.info(`[seed] ${perType} ${type} items`);
	}

	// Denormalized stats, same formula the rating write path (0.3) keeps up to date transactionally.
	const r = schema.ratings;
	const avg = (column: SQL | AnyPgColumn) => sql<number | null>`round(avg(${column}))::smallint`;
	const inner = alias(schema.items, "i2");
	const stats = db
		.select({
			id: inner.id,
			n: sql<number>`count(${r.itemId})::int`.as("stat_n"),
			score: avg(r.overall).as("stat_score"),
			c1: avg(r.c1).as("stat_c1"),
			c2: avg(r.c2).as("stat_c2"),
			c3: avg(r.c3).as("stat_c3"),
			c4: avg(r.c4).as("stat_c4"),
			c5: avg(r.c5).as("stat_c5"),
			c6: avg(r.c6).as("stat_c6"),
		})
		.from(inner)
		.leftJoin(r, eq(r.itemId, inner.id))
		.groupBy(inner.id)
		.as("s");
	await db
		.update(schema.items)
		.set({
			ratingCount: sql`${stats.n}`,
			score: sql`${stats.score}`,
			scoreC1: sql`${stats.c1}`,
			scoreC2: sql`${stats.c2}`,
			scoreC3: sql`${stats.c3}`,
			scoreC4: sql`${stats.c4}`,
			scoreC5: sql`${stats.c5}`,
			scoreC6: sql`${stats.c6}`,
		})
		.from(stats)
		.where(eq(schema.items.id, stats.id));

	console.info(
		`[seed] ${ratingTotal} ratings, stats recomputed — done in ${Math.round(performance.now() - started)} ms`,
	);
}

function uniqueSlug(item: SeedItem, used: Set<string>): string {
	const candidates = [item.slugBase, `${item.slugBase}-${item.releaseDate.slice(0, 4)}`];
	let slug = candidates.find((c) => !used.has(c));
	for (let n = 2; !slug; n++) {
		const c = `${item.slugBase}-${n}`;
		if (!used.has(c)) slug = c;
	}
	used.add(slug);
	return slug;
}

if (import.meta.main) {
	main().then(
		() => process.exit(0),
		(error: unknown) => {
			console.error("[seed] failed:", error);
			process.exit(1);
		},
	);
}
