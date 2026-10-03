import { createHash, randomBytes } from "node:crypto";

const PREFIX = "hx_live_";
const ALPHABET = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz";

export function generateApiKey(): { key: string; prefix: string; hash: string } {
	const bytes = randomBytes(32);
	let body = "";
	for (const b of bytes) body += ALPHABET[b % ALPHABET.length];
	const key = `${PREFIX}${body}`;
	return { key, prefix: key.slice(0, PREFIX.length + 6), hash: hashApiKey(key) };
}

export function hashApiKey(key: string): string {
	return createHash("sha256").update(key).digest("hex");
}

export function parseBearer(header: string | null): string | null {
	const match = header?.match(/^Bearer\s+(hx_live_[0-9A-Za-z]{32})$/);
	return match?.[1] ?? null;
}
