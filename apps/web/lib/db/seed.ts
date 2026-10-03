import { parseArgs } from "node:util";
import { and, count, eq, gt, inArray, like, notLike, type SQL, sql } from "drizzle-orm";
import { z } from "zod";
import type { Database } from "./client";
import * as schema from "./schema";
import { ITEM_TYPES } from "./schema/catalog";
import {
	existingItemPlan,
	generateItem,
	generateRating,
	generateUser,
	raterIndexes,
	type SeedItem,
	seedUserId,
} from "./seed-data";
import { checkRealUsers, checkSeedTarget, seedEnvSchema } from "./seed-guard";

// Fixture data: `bun run db:seed [--users 50] [--items 400] [--batch 250]` builds a fake catalog plus raters.
// `--existing-items` rates the catalog already in the database instead (sandbox Neon branch copied from production)
// and adds statuses, time to beat, follows and lists; `--purge` deletes every seed user and recomputes the stats.
// Idempotent (upserts keyed by deterministic ids) and streaming (keyset batches), so it scales to 10k users / 100k
// items / millions of ratings without holding the catalog in memory.
//
// Marker: seed users have id `seed-user-NNNNNN` and e-mail `seed-user-NNNNNN@example.test`; every fake rating, status,
// time to beat, follow and list belongs to one of them, so deleting them (cascade) removes all seed social data.

const argsSchema = z.object({
	users: z.coerce.number().int().min(0).max(1_000_000).default(50),
	items: z.coerce.number().int().min(4).max(1_000_000).default(400),
	batch: z.coerce.number().int().min(1).max(5_000).default(250),
	existingItems: z.boolean().default(false),
	purge: z.boolean().default(false),
	sandboxAcknowledged: z.boolean().default(false),
});
type SeedOptions = z.infer<typeof argsSchema>;

async function main(): Promise<void> {
	const target = seedEnvSchema.safeParse(process.env);
	if (!target.success) throw new Error(`Refusing to seed: ${target.error.issues.map((i) => i.message).join("; ")}`);
	const refusal = checkSeedTarget(target.data);
	if (refusal) throw new Error(refusal);

	const { values } = parseArgs({
		options: {
			users: { type: "string" },
			items: { type: "string" },
			batch: { type: "string" },
			"existing-items": { type: "boolean" },
			purge: { type: "boolean" },
			"i-know-this-is-sandbox": { type: "boolean" },
		},
	});
	const opts = argsSchema.parse({
		users: values.users,
		items: values.items,
		batch: values.batch,
		existingItems: values["existing-items"],
		purge: values.purge,
		sandboxAcknowledged: values["i-know-this-is-sandbox"],
	});

	// Imported lazily so the target guard above runs before any connection is opened.
	const { db } = await import("./client");
	const started = performance.now();

	if (opts.purge) {
		const purged = await purgeSeedUsers(db, opts.batch);
		console.info(`[seed] purged ${purged} seed users and recomputed stats in ${elapsedSince(started)}`);
		return;
	}

	const [real] = await db
		.select({ n: count() })
		.from(schema.user)
		.where(
			and(
				eq(schema.user.emailVerified, true),
				notLike(schema.user.id, "seed-user-%"),
				// Reserved TLD (RFC 2606): seed and E2E accounts, never a real inbox.
				notLike(schema.user.email, "%.test"),
			),
		);
	const realRefusal = checkRealUsers(real?.n ?? 0, opts.sandboxAcknowledged);
	if (realRefusal) throw new Error(realRefusal);

	if (opts.existingItems) {
		const ratings = await seedExistingItems(db, opts);
		console.info(`[seed] ${ratings} ratings on existing items + social data — done in ${elapsedSince(started)}`);
		return;
	}
	await seedFixtureCatalog(db, opts, started);
}

const elapsedSince = (started: number): string => `${((performance.now() - started) / 1000).toFixed(1)} s`;

async function seedUsers(db: Database, users: number, batch: number): Promise<void> {
	for (let start = 0; start < users; start += batch) {
		const rows = Array.from({ length: Math.min(batch, users - start) }, (_, i) => ({
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
}

async function seedFixtureCatalog(db: Database, opts: SeedOptions, started: number): Promise<void> {
	const elapsed = () => elapsedSince(started);
	await seedUsers(db, opts.users, opts.batch);
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
	await recomputeAllStats(db, opts.batch * 4);

	// Time to beat: about a third of each game's raters "completed" it and report hours around a per-game base
	// (5–64 h main story, ~1.5× with extras, ~2.5× completionist; fewer reporters for the longer styles).
	// One set-based statement, deterministic through hashtext, so reruns are idempotent and stay fast.
	await db.execute(sql`
		with picked as (
			select r.user_id, r.item_id,
				(abs(0::bigint + hashtext(i.wikidata_qid)) % 60 + 5) * 60 as base,
				abs(0::bigint + hashtext(r.user_id || i.wikidata_qid)) % 100 as jitter
			from ratings r join items i on i.id = r.item_id
			where i.type = 'game' and abs(0::bigint + hashtext(i.wikidata_qid || r.user_id)) % 3 = 0
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

	console.info(`[seed] ${ratingTotal} ratings, stats and time to beat — done in ${elapsed()}`);
}

/**
 * `--existing-items`: creates the seed users and rates the items already in the database without touching any item
 * column except the denormalized stats. Items are streamed in keyset batches by id; each batch's ratings and stats
 * are written in one transaction. Social data is then built per batch of users with set-based SQL.
 * Returns the number of ratings written.
 */
export async function seedExistingItems(db: Database, opts: { users: number; batch: number }): Promise<number> {
	const started = performance.now();
	await seedUsers(db, opts.users, opts.batch);
	console.info(`[seed] ${opts.users} users (${elapsedSince(started)})`);

	// Popularity rank within the type, ties broken by a hash so items without popularity data still spread out.
	// Postgres recomputes the window for each batch (a ~100k-row sort, tens of ms), which keeps JS memory flat.
	const i = schema.items;
	const ranked = db
		.select({
			id: i.id,
			rank: sql<number>`(row_number() over (partition by ${i.type} order by ${i.popularity} desc, md5(${i.id}::text)) - 1)::int`.as(
				"rank",
			),
		})
		.from(i)
		.as("ranked");

	let ratingTotal = 0;
	let itemTotal = 0;
	for (let after: string | undefined; ; ) {
		const batch = await db
			.select()
			.from(ranked)
			.where(after ? gt(ranked.id, after) : undefined)
			.orderBy(ranked.id)
			.limit(opts.batch);
		const first = batch[0]?.id;
		const last = batch.at(-1)?.id;
		if (!first || !last) break;

		const columns: RatingColumns = { user: [], item: [], c: [[], [], [], [], [], []], at: [] };
		for (const item of batch) {
			const plan = existingItemPlan(item.id, item.rank, opts.users);
			raterIndexes(plan.ordinal, plan.ratingCount, opts.users).forEach((u, k) => {
				columns.user.push(seedUserId(u));
				columns.item.push(item.id);
				// generateRating only uses the key to seed faker; the item id plays the role of the fixture QID.
				generateRating({ wikidataQid: item.id, quality: plan.quality }, u).forEach((score, c) => {
					columns.c[c]?.push(score);
				});
				columns.at.push(
					new Date(Date.UTC(2026, 0, 1) - ((plan.ordinal % 20_000) + k) * 3_600_000).toISOString(),
				);
			});
		}
		await db.transaction(async (tx) => {
			await tx.execute(sql`set local synchronous_commit to off`);
			for (let r = 0; r < columns.user.length; r += RATING_CHUNK) {
				await tx.execute(insertRatings(columns, r, r + RATING_CHUNK));
			}
			await tx.execute(recomputeStats(first, last));
		});
		ratingTotal += columns.user.length;
		itemTotal += batch.length;
		after = last;
		if (itemTotal % (opts.batch * 20) < opts.batch) {
			console.info(`[seed] ${itemTotal} items, ${ratingTotal} ratings (${elapsedSince(started)})`);
		}
	}
	console.info(`[seed] ${itemTotal} items, ${ratingTotal} ratings + stats (${elapsedSince(started)})`);

	for (let lo = 0; lo < opts.users; lo += opts.batch) {
		const hi = Math.min(opts.users, lo + opts.batch) - 1;
		await db.transaction(async (tx) => {
			await tx.execute(sql`set local synchronous_commit to off`);
			for (const statement of socialStatements(lo, hi, opts.users)) await tx.execute(statement);
		});
	}
	console.info(`[seed] statuses, time to beat, follows and lists (${elapsedSince(started)})`);
	return ratingTotal;
}

const LIST_TITLES = [
	"Favorites",
	"All-time best",
	"Hidden gems",
	"Comfort picks",
	"Worth revisiting",
	"Underrated",
	"Weekend picks",
	"Modern classics",
	"Guilty pleasures",
	"Recommended to friends",
];

/**
 * Social data for seed users lo..hi (inclusive indexes), derived from their ratings with deterministic hashes so
 * reruns converge on the same rows. Rules mirrored from the app: time to beat only for games whose status is
 * `completed` (30..600000 minutes, at least one style), at most LIST_LIMITS.perType (10) lists per user and type
 * (ids n = 1..10 per pair), list items only of the list's type, no self-follows.
 * `abs(0::bigint + hashtext(…))`: hashtext is int4 and abs(-2^31) would overflow on int4.
 */
function socialStatements(lo: number, hi: number, users: number): SQL[] {
	const from = seedUserId(lo);
	const to = seedUserId(hi);
	const titles = sql.raw(`array[${LIST_TITLES.map((t) => `'${t}'`).join(", ")}]`);
	return [
		// Rated items: 70% completed, 15% in progress, 10% dropped, 5% no status.
		sql`
			insert into item_status (user_id, item_id, status)
			select user_id, item_id,
				(case when h < 14 then 'completed' when h < 17 then 'in_progress' else 'dropped' end)::item_status_value
			from (
				select user_id, item_id, abs(0::bigint + hashtext(user_id || item_id::text)) % 20 as h
				from ratings where user_id between ${from} and ${to}
			) r
			where h < 19
			on conflict (user_id, item_id) do update set status = excluded.status
				where item_status.status is distinct from excluded.status`,
		// "Want": about 1 in 8 of the next seed user's rated items that this user has not rated.
		sql`
			insert into item_status (user_id, item_id, status)
			select s.uid, r.item_id, 'want'
			from generate_series(${lo}::int, ${hi}::int) u
			cross join lateral (
				select 'seed-user-' || lpad(u::text, 6, '0') as uid,
					'seed-user-' || lpad(((u + 1) % ${users}::int)::text, 6, '0') as nid
			) s
			join ratings r on r.user_id = s.nid
			where abs(0::bigint + hashtext(s.uid || r.item_id::text)) % 8 = 0
				and not exists (select 1 from ratings m where m.user_id = s.uid and m.item_id = r.item_id)
			on conflict (user_id, item_id) do nothing`,
		// Keep the app invariant if a status changed between runs: time to beat only while completed.
		sql`
			delete from time_to_beat t
			where t.user_id between ${from} and ${to}
				and not exists (
					select 1 from item_status s
					where s.user_id = t.user_id and s.item_id = t.item_id and s.status = 'completed'
				)`,
		// Half of the completed games get a report around a per-game base (5–64 h main, ~1.5× extras, ~2.5× 100%).
		sql`
			insert into time_to_beat (user_id, item_id, main_minutes, extras_minutes, completionist_minutes)
			select user_id, item_id,
				round(base * (0.8 + j / 250.0) / 30) * 30,
				case when j % 2 = 0 then round(base * 1.5 * (0.8 + j / 250.0) / 30) * 30 end,
				case when j % 4 = 0 then round(base * 2.5 * (0.8 + j / 250.0) / 30) * 30 end
			from (
				select s.user_id, s.item_id,
					(abs(0::bigint + hashtext(i.id::text)) % 60 + 5) * 60 as base,
					abs(0::bigint + hashtext(s.user_id || i.id::text)) % 100 as j
				from item_status s join items i on i.id = s.item_id
				where s.user_id between ${from} and ${to} and s.status = 'completed' and i.type = 'game'
					and abs(0::bigint + hashtext(i.id::text || s.user_id)) % 2 = 0
			) p
			on conflict (user_id, item_id) do update set
				main_minutes = excluded.main_minutes, extras_minutes = excluded.extras_minutes,
				completionist_minutes = excluded.completionist_minutes`,
		// 2–21 follows per user, skewed (cube of a uniform) towards low indexes, the power raters.
		sql`
			insert into follows (follower_id, followee_id)
			select 'seed-user-' || lpad(u::text, 6, '0'), 'seed-user-' || lpad(f::text, 6, '0')
			from generate_series(${lo}::int, ${hi}::int) u
			cross join lateral (
				select distinct floor(${users}::int * power((abs(0::bigint + hashtext(u::text || ':' || j::text)) % 1000000) / 1000000.0, 3))::int as f
				from generate_series(1, 2 + abs(0::bigint + hashtext('follows:' || u::text)) % 20) j
			) x
			where f <> u
			on conflict do nothing`,
		// A third of (user, type) pairs with 5+ ratings get 1..10 lists (one more per 25 ratings), ids from md5.
		sql`
			insert into lists (id, user_id, type, title, is_public)
			select md5('seed-list:' || c.user_id || ':' || c.type::text || ':' || n::text)::uuid, c.user_id, c.type,
				(${titles})[n], abs(0::bigint + hashtext(c.user_id || c.type::text || n::text)) % 5 <> 0
			from (
				select r.user_id, i.type, count(*) as cnt
				from ratings r join items i on i.id = r.item_id
				where r.user_id between ${from} and ${to}
				group by r.user_id, i.type
			) c
			cross join lateral generate_series(1, least(10, 1 + c.cnt / 25)::int) n
			where c.cnt >= 5 and abs(0::bigint + hashtext(c.user_id || c.type::text)) % 3 = 0
			on conflict (id) do update set title = excluded.title, is_public = excluded.is_public`,
		// Up to 25 of the owner's best-rated items of the list's type, a different third per list.
		sql`
			insert into list_items (list_id, item_id, position)
			select l.id, x.item_id, x.pos
			from lists l
			cross join lateral (
				select r.item_id, (row_number() over (order by r.overall desc, r.item_id))::int as pos
				from ratings r join items i on i.id = r.item_id
				where r.user_id = l.user_id and i.type = l.type and abs(0::bigint + hashtext(l.id::text || r.item_id::text)) % 3 = 0
				order by r.overall desc, r.item_id
				limit 25
			) x
			where l.user_id between ${from} and ${to}
			on conflict (list_id, item_id) do update set position = excluded.position`,
	];
}

/**
 * `--purge`: deletes every seed user (both markers: id prefix and @example.test), which cascades to their ratings,
 * statuses, time to beat, follows and lists, then recomputes the stats of every item. Returns the users deleted.
 */
export async function purgeSeedUsers(db: Database, batch: number): Promise<number> {
	const isSeedUser = and(like(schema.user.id, "seed-user-%"), like(schema.user.email, "%@example.test"));
	let purged = 0;
	for (;;) {
		const chunk = db.select({ id: schema.user.id }).from(schema.user).where(isSeedUser).limit(batch);
		const deleted = await db
			.delete(schema.user)
			.where(inArray(schema.user.id, chunk))
			.returning({ id: schema.user.id });
		if (!deleted.length) break;
		purged += deleted.length;
	}
	await recomputeAllStats(db, batch * 4);
	return purged;
}

/** Recomputes the stats of every item in keyset batches by id. */
async function recomputeAllStats(db: Database, batch: number): Promise<void> {
	for (let after: string | undefined; ; ) {
		const ids = await db
			.select({ id: schema.items.id })
			.from(schema.items)
			.where(after ? gt(schema.items.id, after) : undefined)
			.orderBy(schema.items.id)
			.limit(batch);
		const first = ids[0]?.id;
		const last = ids.at(-1)?.id;
		if (!first || !last) return;
		await db.execute(recomputeStats(first, last));
		after = last;
	}
}

/**
 * Denormalized stats for items with id in [from, to], same formula as recomputeItemStats() in lib/ratings.ts
 * (count + rounded averages, null without ratings). Raw SQL so only the stat columns change (no updated_at bump),
 * and rows whose stats are already right are not rewritten.
 */
function recomputeStats(from: string, to: string): SQL {
	return sql`
		update items i set
			rating_count = s.n, score = s.score, score_c_1 = s.c1, score_c_2 = s.c2, score_c_3 = s.c3,
			score_c_4 = s.c4, score_c_5 = s.c5, score_c_6 = s.c6
		from (
			select i2.id, count(r.item_id)::int as n,
				round(avg(r.overall))::smallint as score,
				round(avg(r.c1))::smallint as c1, round(avg(r.c2))::smallint as c2, round(avg(r.c3))::smallint as c3,
				round(avg(r.c4))::smallint as c4, round(avg(r.c5))::smallint as c5, round(avg(r.c6))::smallint as c6
			from items i2 left join ratings r on r.item_id = i2.id
			where i2.id between ${from}::uuid and ${to}::uuid
			group by i2.id
		) s
		where i.id = s.id
			and (i.rating_count, i.score, i.score_c_1, i.score_c_2, i.score_c_3, i.score_c_4, i.score_c_5, i.score_c_6)
				is distinct from (s.n, s.score, s.c1, s.c2, s.c3, s.c4, s.c5, s.c6)`;
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
