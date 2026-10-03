import { createHash, timingSafeEqual } from "node:crypto";

/**
 * `Authorization: Bearer <CRON_SECRET>` check in constant time. Both sides are hashed first so lengths always
 * match and the comparison leaks nothing about the secret's length. No secret configured = always denied.
 */
export function isCronAuthorized(header: string | null, secret: string | undefined): boolean {
	if (!secret || !header?.startsWith("Bearer ")) return false;
	const digest = (value: string) => createHash("sha256").update(value).digest();
	return timingSafeEqual(digest(header.slice("Bearer ".length)), digest(secret));
}
