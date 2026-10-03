import type { PgliteDatabase } from "drizzle-orm/pglite";
import { migrate as migratePglite } from "drizzle-orm/pglite/migrator";
import type { PostgresJsDatabase } from "drizzle-orm/postgres-js";
import { migrate as migratePostgres } from "drizzle-orm/postgres-js/migrator";
import { env } from "../env";
import { db } from "./client";

const migrationsFolder = new URL("./migrations", import.meta.url).pathname.replace(/^\/([A-Za-z]:)/, "$1");

export async function runMigrations(): Promise<void> {
	if (env.DATABASE_URL) await migratePostgres(db as unknown as PostgresJsDatabase, { migrationsFolder });
	else await migratePglite(db as unknown as PgliteDatabase, { migrationsFolder });
}

if (import.meta.main) {
	const started = performance.now();
	try {
		await runMigrations();
		console.info(`[db] migrations applied in ${Math.round(performance.now() - started)} ms`);
		process.exit(0);
	} catch (error) {
		console.error("[db] migration failed:", error);
		process.exit(1);
	}
}
