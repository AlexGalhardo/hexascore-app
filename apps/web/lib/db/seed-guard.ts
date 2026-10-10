import { z } from "zod";

// Pure safety checks for lib/db/seed.ts: fake users/ratings/lists land in a local database or in the sandbox Neon
// branch. Production is allowed only with an explicit double opt-in (SEED_TARGET=production and
// --i-know-this-is-production), for the pre-launch demo data the owner removes later with `--purge`.

export const seedEnvSchema = z.object({
	DATABASE_URL: z.string().trim().optional(),
	VERCEL_ENV: z.string().optional(),
	SEED_TARGET: z
		.enum(["sandbox", "production"], { message: 'SEED_TARGET only accepts "sandbox" or "production"' })
		.optional(),
});
export type SeedEnv = z.infer<typeof seedEnvSchema>;

/** Verified accounts allowed outside the seed marker before the run needs `--i-know-this-is-sandbox`. */
export const REAL_USER_THRESHOLD = 50;

const LOCAL_HOSTS = new Set(["localhost", "127.0.0.1", "::1", "[::1]"]);

/** True for PGlite (no DATABASE_URL) or a Postgres on this machine. */
export function isLocalDatabase(databaseUrl: string | undefined): boolean {
	if (!databaseUrl) return true;
	try {
		return LOCAL_HOSTS.has(new URL(databaseUrl).hostname);
	} catch {
		return false;
	}
}

/** True when both production opt-ins are present. */
export function isProductionSeed(env: SeedEnv, productionAcknowledged: boolean): boolean {
	return env.SEED_TARGET === "production" && productionAcknowledged;
}

/** Checked before any connection is opened. Returns the refusal reason, or null when the target is allowed. */
export function checkSeedTarget(env: SeedEnv, productionAcknowledged = false): string | null {
	if (isProductionSeed(env, productionAcknowledged)) return null;
	if (env.SEED_TARGET === "production") {
		return "Refusing to seed: SEED_TARGET=production also needs --i-know-this-is-production.";
	}
	if (env.VERCEL_ENV === "production") {
		return [
			"Refusing to seed: VERCEL_ENV=production (fake users/ratings would distort real metascores).",
			"To seed production on purpose, run with SEED_TARGET=production and --i-know-this-is-production.",
		].join("\n");
	}
	if (isLocalDatabase(env.DATABASE_URL) || env.SEED_TARGET === "sandbox") return null;
	return [
		"Refusing to seed: DATABASE_URL points to a remote database.",
		"Seed data is only allowed locally (PGlite, localhost, 127.0.0.1) or in the sandbox Neon branch.",
		"If this really is the sandbox branch, run again with SEED_TARGET=sandbox.",
	].join("\n");
}

/** Checked after connecting: a copy of production holds real verified accounts, a dev database does not. */
export function checkRealUsers(realVerifiedUsers: number, acknowledged: boolean): string | null {
	if (acknowledged || realVerifiedUsers <= REAL_USER_THRESHOLD) return null;
	return [
		`Refusing to seed: the database has ${realVerifiedUsers} verified users outside the seed marker`,
		`(id seed-user-*, e-mail on a .test domain), more than ${REAL_USER_THRESHOLD}. It looks like production data.`,
		"Seed only the sandbox branch copied from production; if this is it, pass --i-know-this-is-sandbox.",
	].join("\n");
}
