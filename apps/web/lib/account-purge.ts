import { and, inArray, isNotNull, lt, sql } from "drizzle-orm";
import { purgeCutoff } from "./account-deletion";
import type { Database } from "./db/client";
import { account, passkey, session, twoFactor, user, verification } from "./db/schema";

// Anonymizes instead of deleting the row: ratings keep their user_id so the metascore is not distorted,
// but nothing left on the row identifies the person and no credential can sign in again.
export async function purgeScheduledAccounts(db: Database, now = new Date()): Promise<string[]> {
	const due = await db
		.select({ id: user.id, email: user.email })
		.from(user)
		.where(and(isNotNull(user.deletionScheduledAt), lt(user.deletionScheduledAt, purgeCutoff(now))));
	if (due.length === 0) return [];
	const ids = due.map((u) => u.id);

	await db.transaction(async (tx) => {
		await tx.delete(session).where(inArray(session.userId, ids));
		await tx.delete(account).where(inArray(account.userId, ids));
		await tx.delete(passkey).where(inArray(passkey.userId, ids));
		await tx.delete(twoFactor).where(inArray(twoFactor.userId, ids));
		await tx.delete(verification).where(
			inArray(
				verification.identifier,
				due.map((u) => u.email),
			),
		);
		await tx
			.update(user)
			.set({
				email: sql`'deleted-' || ${user.id} || '@invalid'`,
				name: "Deleted user",
				username: null,
				displayUsername: null,
				image: null,
				emailVerified: false,
				twoFactorEnabled: false,
				profilePrivate: true,
				deletionScheduledAt: null,
			})
			.where(inArray(user.id, ids));
	});
	return ids;
}
