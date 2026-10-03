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
	const elapsed = () => `${((performance.now() - started) / 1000).toFixed(1)} s`;

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
	console.info(`[seed] ${opts.users} users (${elapsed()})`);

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
				// Fixture data: skipping the WAL flush wait per commit is safe (a crash just means re-running the seed).
				await tx.execute(sql`set local synchronous_commit to off`);
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

				const columns: RatingColumns = { user: [], item: [], c: [[], [], [], [], [], []], at: [] };
				batch.forEach((item, i) => {
					const itemOrdinal = ordinal + i;
					const itemId = idOf(item);
					raterIndexes(itemOrdinal, item.ratingCount, opts.users).forEach((u, k) => {
						columns.user.push(seedUserId(u));
						columns.item.push(itemId);
						generateRating(item, u).forEach((score, c) => {
							columns.c[c]?.push(score);
						});
						columns.at.push(
							new Date(Date.UTC(2026, 0, 1) - (itemOrdinal * 31 + k) * 3_600_000).toISOString(),
						);
					});
				});
				for (let r = 0; r < columns.user.length; r += RATING_CHUNK) {
					await tx.execute(insertRatings(columns, r, r + RATING_CHUNK));
				}
				ratingTotal += columns.user.length;
			});
			ordinal += batch.length;
		}
		console.info(`[seed] ${perType} ${type} items (${elapsed()})`);
	}

	console.info(`[seed] ${ratingTotal} ratings written (${elapsed()}), recomputing stats…`);
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

	// Time to beat: about a third of each game's raters "completed" it and report hours around a per-game base
	// (5–64 h main story, ~1.5× with extras, ~2.5× completionist; fewer reporters for the longer styles).
	// One set-based statement, deterministic through hashtext, so reruns are idempotent and stay fast.
	await db.execute(sql`
		with picked as (
			select r.user_id, r.item_id,
				(abs(hashtext(i.wikidata_qid)) % 60 + 5) * 60 as base,
				abs(hashtext(r.user_id || i.wikidata_qid)) % 100 as jitter
			from ratings r join items i on i.id = r.item_id
			where i.type = 'game' and abs(hashtext(i.wikidata_qid || r.user_id)) % 3 = 0
		), statuses as (
			insert into item_status (user_id, item_id, status)
			select user_id, item_id, 'completed' from picked
			on conflict (user_id, item_id) do update set status = 'completed'
			returning user_id, item_id
		)
		insert into time_to_beat (user_id, item_id, main_minutes, extras_minutes, completionist_minutes)
		select p.user_id, p.item_id,
			round(p.base * (0.8 + p.jitter / 250.0) / 30) * 30,
			case when p.jitter % 2 = 0 then round(p.base * 1.5 * (0.8 + p.jitter / 250.0) / 30) * 30 end,
			case when p.jitter % 4 = 0 then round(p.base * 2.5 * (0.8 + p.jitter / 250.0) / 30) * 30 end
		from picked p join statuses s on s.user_id = p.user_id and s.item_id = p.item_id
		on conflict (user_id, item_id) do update set
			main_minutes = excluded.main_minutes, extras_minutes = excluded.extras_minutes,
			completionist_minutes = excluded.completionist_minutes`);

	console.info(
		`[seed] ${ratingTotal} ratings, stats and time to beat — done in ${Math.round(performance.now() - started)} ms`,
	);
}

type RatingColumns = { user: string[]; item: string[]; c: number[][]; at: string[] };

/** Rows per ratings statement: one array parameter per column, so the size is not bound by the 65k-parameter cap. */
const RATING_CHUNK = 20_000;

// Postgres array literal. Only generated values go in (seed ids, uuids, integers, ISO dates): none contain commas,
// quotes, braces or spaces, so no element quoting is needed.
const pgArray = (values: (string | number)[]): string => `{${values.join(",")}}`;

/**
 * Columnar upsert via `unnest`: ~10× faster than multi-row VALUES (one parse/plan per 20k rows, 9 parameters total)
 * and portable across postgres-js and PGlite.
 */
function insertRatings(cols: RatingColumns, from: number, to: number): SQL {
	const slice = <T extends string | number>(a: T[]) => pgArray(a.slice(from, to));
	const [c1, c2, c3, c4, c5, c6] = cols.c.map((c) => slice(c)) as [string, string, string, string, string, string];
	return sql`
		insert into ratings (user_id, item_id, c1, c2, c3, c4, c5, c6, created_at)
		select * from unnest(
			${slice(cols.user)}::text[], ${slice(cols.item)}::uuid[],
			${c1}::smallint[], ${c2}::smallint[], ${c3}::smallint[], ${c4}::smallint[], ${c5}::smallint[], ${c6}::smallint[],
			${slice(cols.at)}::timestamptz[]
		)
		on conflict (user_id, item_id) do update set
			c1 = excluded.c1, c2 = excluded.c2, c3 = excluded.c3, c4 = excluded.c4, c5 = excluded.c5, c6 = excluded.c6,
			created_at = excluded.created_at`;
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
