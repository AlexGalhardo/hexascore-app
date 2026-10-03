import { hashPassword } from "better-auth/crypto";
import { sql } from "drizzle-orm";
import { z } from "zod";
import { passwordSchema } from "@/lib/validation";
import type { Database } from "./client";
import * as schema from "./schema";

// Default admin for local and sandbox only (the seed refuses production). The `.test` domain (RFC 2606) never
// receives mail, so nobody can request a password reset for it. Documented in docs/admin.md.
export const SEED_ADMIN_ID = "seed-admin";
export const SEED_ADMIN_DEFAULTS = { email: "admin@hexascore.test", password: "hexascoreBR@123", username: "admin" };

const adminEnvSchema = z.object({
	SEED_ADMIN_EMAIL: z.email().default(SEED_ADMIN_DEFAULTS.email),
	SEED_ADMIN_PASSWORD: passwordSchema.default(SEED_ADMIN_DEFAULTS.password),
});

/** Upserts the admin user and its email+password credential; re-running resets the password to the configured one. */
export async function seedAdmin(db: Database, env: Record<string, string | undefined> = process.env): Promise<string> {
	const { SEED_ADMIN_EMAIL: email, SEED_ADMIN_PASSWORD: password } = adminEnvSchema.parse(env);
	const hash = await hashPassword(password);
	await db.transaction(async (tx) => {
		await tx
			.insert(schema.user)
			.values({
				id: SEED_ADMIN_ID,
				name: "Admin",
				email,
				emailVerified: true,
				username: SEED_ADMIN_DEFAULTS.username,
				displayUsername: SEED_ADMIN_DEFAULTS.username,
				role: "admin",
			})
			.onConflictDoUpdate({
				target: schema.user.id,
				set: { email: sql`excluded.email`, emailVerified: true, role: "admin", banned: false },
			});
		await tx
			.insert(schema.account)
			.values({
				id: `${SEED_ADMIN_ID}-credential`,
				accountId: SEED_ADMIN_ID,
				providerId: "credential",
				userId: SEED_ADMIN_ID,
				password: hash,
				updatedAt: new Date(),
			})
			.onConflictDoUpdate({ target: schema.account.id, set: { password: hash, updatedAt: new Date() } });
	});
	return email;
}
