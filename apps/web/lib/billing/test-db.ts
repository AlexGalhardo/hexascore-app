import { PGlite } from "@electric-sql/pglite";
import { pg_trgm } from "@electric-sql/pglite/contrib/pg_trgm";
import { unaccent } from "@electric-sql/pglite/contrib/unaccent";
import { drizzle } from "drizzle-orm/pglite";
import { migrate } from "drizzle-orm/pglite/migrator";
import type { Database } from "../db/client";
import * as schema from "../db/schema";

const migrationsFolder = new URL("../db/migrations", import.meta.url).pathname.replace(/^\/([A-Za-z]:)/, "$1");

/** Fresh in-memory Postgres with every migration applied; tests never touch the dev database. */
export async function createTestDb(): Promise<Database> {
	const client = new PGlite({ extensions: { pg_trgm, unaccent } });
	const db = drizzle(client, { schema, casing: "snake_case" });
	await migrate(db, { migrationsFolder });
	return db as unknown as Database;
}

export async function createTestUser(db: Database, id: string): Promise<void> {
	await db.insert(schema.user).values({ id, name: id, email: `${id}@example.com`, emailVerified: true });
}
