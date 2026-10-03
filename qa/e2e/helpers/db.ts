import { randomBytes } from "node:crypto";
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

/** Same as `bun run admin:grant`; run it before logging in so the new session carries the role. */
export async function grantAdmin(email: string): Promise<void> {
	await sql()`update "user" set role = 'admin' where email = ${email.toLowerCase()}`;
}

export type TestItem = { id: string; title: string; href: string };

/**
 * A fresh game with no ratings, so threshold and feed assertions do not depend on the seed. Kept afterwards like
 * E2E users (slug and QID start with `e2e-`).
 */
export async function createItem(): Promise<TestItem> {
	const id = randomBytes(4).toString("hex");
	const slug = `e2e-game-${id}`;
	const title = `E2E Game ${id}`;
	const [row] = await sql()<{ id: string }[]>`
		insert into items (id, type, slug, wikidata_qid) values (gen_random_uuid(), 'game', ${slug}, ${`e2e-${id}`})
		returning id`;
	if (!row) throw new Error("item insert returned no row");
	await sql()`insert into item_translations (item_id, locale, title) values (${row.id}, 'en', ${title})`;
	return { id: row.id, title, href: `/games/${slug}` };
}

/** Random lowercase letters: nicknames accept only `a-z`. */
export function letters(n: number): string {
	return Array.from(randomBytes(n), (b) => String.fromCharCode(97 + (b % 26))).join("");
}

export type DbMember = { id: string; nickname: string };

/** A verified, public member written straight to the DB: no credentials, it only owns ratings others look at. */
export async function createMember(tag: string): Promise<DbMember> {
	const id = randomBytes(4).toString("hex");
	const nickname = `ee${tag.replace(/[^a-z]/g, "").slice(0, 6)}${letters(8)}`;
	await sql()`
		insert into "user" (id, name, email, email_verified, username, display_username, updated_at)
		values (${`e2e-${id}`}, ${`E2E ${tag}`}, ${`e2e-${tag}-${id}@${E2E_EMAIL_DOMAIN}`}, true, ${nickname},
			${nickname}, now())`;
	return { id: `e2e-${id}`, nickname };
}

/**
 * Writes a rating and recomputes the item stats the way `lib/ratings.ts` does (lock the item, recompute from
 * `ratings`; Postgres `round` on numeric = half away from zero). Item pages are cached per item: insert before the
 * first visit, or the page keeps the old stats until the next write through the app.
 */
export async function insertRating(userId: string, itemId: string, scores: readonly number[]): Promise<void> {
	const [c1 = 0, c2 = 0, c3 = 0, c4 = 0, c5 = 0, c6 = 0] = scores;
	await sql().begin(async (tx) => {
		await tx`select id from items where id = ${itemId} for update`;
		await tx`
			insert into ratings (user_id, item_id, c1, c2, c3, c4, c5, c6)
			values (${userId}, ${itemId}, ${c1}, ${c2}, ${c3}, ${c4}, ${c5}, ${c6})`;
		await tx`
			update items set (rating_count, score, score_c_1, score_c_2, score_c_3, score_c_4, score_c_5, score_c_6) = (
				select count(*), round(avg(overall)), round(avg(c1)), round(avg(c2)), round(avg(c3)), round(avg(c4)),
					round(avg(c5)), round(avg(c6))
				from ratings where item_id = ${itemId})
			where id = ${itemId}`;
	});
}
