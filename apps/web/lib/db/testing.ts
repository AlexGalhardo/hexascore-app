import { fileURLToPath } from "node:url";
import { PGlite } from "@electric-sql/pglite";
import { pg_trgm } from "@electric-sql/pglite/contrib/pg_trgm";
import { unaccent } from "@electric-sql/pglite/contrib/unaccent";
import { drizzle } from "drizzle-orm/pglite";
import { migrate } from "drizzle-orm/pglite/migrator";
import type { Database } from "./client";
import * as schema from "./schema";

// Fresh in-memory Postgres with every migration applied. Does not import ./client, so tests never touch .pglite/.
export async function createTestDatabase(): Promise<{ db: Database; close: () => Promise<void> }> {
	const client = new PGlite({ extensions: { pg_trgm, unaccent } });
	const db = drizzle(client, { schema, casing: "snake_case" });
	await migrate(db, { migrationsFolder: fileURLToPath(new URL("./migrations", import.meta.url)) });
	return { db: db as unknown as Database, close: () => client.close() };
}
