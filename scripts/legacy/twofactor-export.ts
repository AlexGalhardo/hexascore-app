// Decrypts the legacy Better Auth 2FA rows and writes them as NDJSON
// {userId, secret, backupCodes[]} to stdout, sorted by user id, for
// `hexascore:import-legacy --stage twofactor --stdin` (SDD 03 §7.6).
//
// Run it from a temporary copy of legacy/apps/web with its node_modules
// (better-auth 1.7.7), so the decryption is Better Auth's own code:
//
//   cp -r legacy/apps/web "$TMP/web" && cd "$TMP/web" && bun install --frozen-lockfile --ignore-scripts
//   mkdir -p scripts/legacy && cp <app>/scripts/legacy/twofactor-export.ts scripts/legacy/
//   BETTER_AUTH_SECRET=... LEGACY_DATABASE_URL=... bun scripts/legacy/twofactor-export.ts \
//     | <app>/app hexascore:import-legacy --stage twofactor --stdin
//
// LEGACY_DATABASE_URL is the database holding the restored legacy_import
// schema. Nothing is written to disk and no secret reaches stderr: errors name
// the user id only.
import { SQL } from "bun";
import { symmetricDecrypt } from "better-auth/crypto";

const key = process.env.BETTER_AUTH_SECRET;
const url = process.env.LEGACY_DATABASE_URL;

if (!key || !url) {
	console.error("twofactor-export: BETTER_AUTH_SECRET and LEGACY_DATABASE_URL are required");
	process.exit(2);
}

const db = new SQL(url);

try {
	const rows: { user_id: string; secret: string; backup_codes: string }[] =
		await db`SELECT user_id, secret, backup_codes FROM legacy_import.two_factor ORDER BY user_id COLLATE "C"`;

	for (const row of rows) {
		let line: string;
		try {
			const secret = await symmetricDecrypt({ key, data: row.secret });
			// Better Auth 1.7.7 default storeBackupCodes "encrypted": a JSON array of unused codes.
			const backupCodes: unknown = JSON.parse(await symmetricDecrypt({ key, data: row.backup_codes }));
			if (!Array.isArray(backupCodes) || !backupCodes.every((c) => typeof c === "string")) {
				throw new Error("backup codes are not a string array");
			}
			line = JSON.stringify({ userId: row.user_id, secret, backupCodes });
		} catch {
			console.error(`twofactor-export: cannot decrypt the row of user ${row.user_id}`);
			process.exit(1);
		}
		process.stdout.write(`${line}\n`);
	}
} finally {
	await db.close();
}
