import postgres from "postgres";

// Test-only access to the target deployment's database (E2E_DATABASE_URL). The app has no test backdoor: flows
// that need an email link read what the app wrote (reset tokens in `verification`) or mint what it would have sent
// (see verificationUrl in auth.ts).

export const E2E_EMAIL_DOMAIN = "e2e.example.test";

let client: postgres.Sql | undefined;

export function hasDatabase(): boolean {
	return Boolean(process.env.E2E_DATABASE_URL);
}

export function sql(): postgres.Sql {
	const url = process.env.E2E_DATABASE_URL;
	if (!url) throw new Error("E2E_DATABASE_URL is not set");
	client ??= postgres(url, { max: 2, idle_timeout: 5 });
	return client;
}

export type DbUser = {
	id: string;
	emailVerified: boolean;
	deletionScheduledAt: Date | null;
	twoFactorEnabled: boolean;
};

export async function findUser(email: string): Promise<DbUser | undefined> {
	const rows = await sql()<DbUser[]>`
		select id, email_verified as "emailVerified", deletion_scheduled_at as "deletionScheduledAt",
			coalesce(two_factor_enabled, false) as "twoFactorEnabled"
		from "user" where email = ${email.toLowerCase()}`;
	return rows[0];
}

/** Latest password-reset token Better Auth stored for this user (`reset-password:<token>` → user id). */
export async function latestResetToken(email: string): Promise<string | undefined> {
	const rows = await sql()<{ identifier: string }[]>`
		select v.identifier from verification v join "user" u on u.id = v.value
		where u.email = ${email.toLowerCase()} and v.identifier like 'reset-password:%' and v.expires_at > now()
		order by v.created_at desc limit 1`;
	return rows[0]?.identifier.slice("reset-password:".length);
}

/** Removes accounts created by previous runs (cascades to sessions, accounts, keys). */
export async function deleteE2eUsers(): Promise<void> {
	await sql()`delete from "user" where email like ${`%@${E2E_EMAIL_DOMAIN}`}`;
}
