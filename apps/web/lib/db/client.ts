import { PGlite } from "@electric-sql/pglite";
import { pg_trgm } from "@electric-sql/pglite/contrib/pg_trgm";
import { unaccent } from "@electric-sql/pglite/contrib/unaccent";
import type { PgDatabase, PgQueryResultHKT } from "drizzle-orm/pg-core";
import { drizzle as drizzlePglite } from "drizzle-orm/pglite";
import { drizzle as drizzlePostgres } from "drizzle-orm/postgres-js";
import postgres from "postgres";
import { env } from "../env";
import * as schema from "./schema";

export type Database = PgDatabase<PgQueryResultHKT, typeof schema>;

function createDatabase(): Database {
	const url = env.DATABASE_URL;
	if (url) return drizzlePostgres(postgres(url, { max: 10 }), { schema, casing: "snake_case" });
	// No DATABASE_URL: in-process Postgres (PGlite) so `bun run dev` works without Docker.
	// `next build` imports this module in many parallel workers; they cannot share one on-disk PGlite
	// (WASM abort), and the build never queries it, so each worker gets a throwaway in-memory instance.
	const dataDir =
		process.env.NEXT_PHASE === "phase-production-build" ? undefined : (process.env.PGLITE_DIR ?? ".pglite");
	const client = new PGlite({ dataDir, extensions: { pg_trgm, unaccent } });
	return drizzlePglite(client, { schema, casing: "snake_case" }) as unknown as Database;
}

const globalForDb = globalThis as unknown as { db?: Database };

export const db: Database = globalForDb.db ?? createDatabase();
if (process.env.NODE_ENV !== "production") globalForDb.db = db;
