import { createHash, timingSafeEqual } from "node:crypto";

// Hashing both sides gives equal-length buffers, so the comparison time does not leak the secret length.
export function isAuthorizedCron(authorization: string | null, secret: string | undefined): boolean {
	if (!secret || !authorization) return false;
	const digest = (value: string): Buffer => createHash("sha256").update(value).digest();
	return timingSafeEqual(digest(authorization), digest(`Bearer ${secret}`));
}
