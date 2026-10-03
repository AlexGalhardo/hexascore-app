import { afterAll, beforeAll, expect, test } from "bun:test";
import { readFile } from "node:fs/promises";
import { asc, sql } from "drizzle-orm";
import type { Database } from "./client";
import { user } from "./schema/auth";
import { items } from "./schema/catalog";
import { listItems, lists } from "./schema/social";
import { createTestDatabase } from "./testing";

let db: Database;
let close: () => Promise<void>;
const id = (n: number) => `01900000-0000-7000-8000-${String(n).padStart(12, "0")}`;

beforeAll(async () => {
	({ db, close } = await createTestDatabase());
	// Roll `lists` back to its pre-0005 shape so the migration's backfill runs against existing rows.
	await db.execute(sql`drop index lists_user_type_idx`);
	await db.execute(sql`alter table lists drop column type`);
	await db.execute(sql`create index lists_user_idx on lists (user_id)`);
});

afterAll(async () => {
	await close();
});

test("0005 backfills the list type from its first item, 'game' when empty", async () => {
	await db.insert(user).values({ id: "u", name: "U", email: "u@x.io" });
	await db.insert(items).values([
		{ id: id(1), type: "book", slug: "b", wikidataQid: "Q1" },
		{ id: id(2), type: "movie", slug: "m", wikidataQid: "Q2" },
	]);
	await db.execute(sql`insert into lists (id, user_id, title) values
		(${id(10)}, 'u', 'books first'), (${id(11)}, 'u', 'movie only'), (${id(12)}, 'u', 'empty')`);
	await db.insert(listItems).values([
		{ listId: id(10), itemId: id(2), position: 2 },
		{ listId: id(10), itemId: id(1), position: 1 },
		{ listId: id(11), itemId: id(2), position: 1 },
	]);

	const file = await readFile(new URL("./migrations/0005_list_type.sql", import.meta.url), "utf8");
	for (const statement of file.split("--> statement-breakpoint")) await db.execute(sql.raw(statement));

	const rows = await db.select({ id: lists.id, type: lists.type }).from(lists).orderBy(asc(lists.id));
	expect(rows).toEqual([
		{ id: id(10), type: "book" },
		{ id: id(11), type: "movie" },
		{ id: id(12), type: "game" },
	]);
	const untyped = db
		.execute(sql`insert into lists (id, user_id, title) values (${id(13)}, 'u', 'no type')`)
		.then(() => "inserted");
	await expect(untyped).rejects.toThrow(); // NOT NULL is enforced after the backfill
});
