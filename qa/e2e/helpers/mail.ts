import { spawnSync } from "node:child_process";
import { dbContainer } from "./db";
import { escapeRegExp } from "./t";

// The container logs mail bodies at debug level (docker-compose.yml, LOG_LEVEL). Against an environment
// without docker access the mail-dependent specs skip instead of failing.
const container = process.env.E2E_APP_CONTAINER ?? "hexascore-app-1";

export function canReadMail(): boolean {
	return spawnSync("docker", ["inspect", container], { stdio: "ignore" }).status === 0;
}

function logLines(): string[] {
	const out = spawnSync("docker", ["logs", "--since", "5m", container], {
		encoding: "utf8",
		maxBuffer: 256 * 1024 * 1024,
	});
	return `${out.stdout}\n${out.stderr}`.split("\n");
}

// Returns the path and query of the newest link containing `fragment` in a mail sent to `email`.
// The host is dropped on purpose: APP_URL inside the container is not the URL the browser uses.
export async function mailLink(
	email: string,
	fragment: string,
	timeoutMs = 15_000,
): Promise<string> {
	const deadline = Date.now() + timeoutMs;
	while (Date.now() < deadline) {
		const lines = logLines();
		for (let i = lines.length - 1; i >= 0; i--) {
			const line = lines[i] ?? "";
			if (!line.includes('"mail (log transport)"') || !line.includes(`<${email}>`)) continue;
			for (const next of lines.slice(i + 1, i + 6)) {
				if (!next.includes('"mail body"')) continue;
				const text = (JSON.parse(next) as { text: string }).text;
				const match = text.match(
					new RegExp(`https?://[^\\s)]*${escapeRegExp(fragment)}[^\\s)]*`),
				);
				if (match) {
					// The log transport redacts the email query parameter; the recipient is known, so restore it.
					const url = new URL(match[0].replace("[REDACTED]", encodeURIComponent(email)));
					return `${url.pathname}${url.search}`;
				}
			}
		}
		await new Promise((resolve) => setTimeout(resolve, 500));
	}
	throw new Error(`no mail with a ${fragment} link for ${email} in ${container} logs`);
}

// Inserts a user_notifications row for the account with this email (see helpers/db.ts).
export function seedNotification(email: string, kind: string, dedupeKey: string): void {
	const sql = `INSERT INTO user_notifications (id, user_id, kind, payload, dedupe_key)
		SELECT gen_random_uuid(), id, '${kind}', '{}', '${dedupeKey}' FROM users WHERE email = '${email}'`;
	const out = spawnSync(
		"docker",
		[
			"exec",
			dbContainer(),
			"psql",
			"-U",
			"hexascore",
			"-d",
			process.env.E2E_DB_NAME ?? "hexascore",
			"-v",
			"ON_ERROR_STOP=1",
			"-c",
			sql,
		],
		{ encoding: "utf8" },
	);
	if (out.status !== 0) throw new Error(`seedNotification failed: ${out.stderr}`);
}
