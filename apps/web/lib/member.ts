import type { ActionError } from "./action-result";
import type { SessionUser } from "./session";

export type Member = SessionUser & { username: string };

/**
 * Social writes (rating, status, lists, follows, reports) need a verified, non-banned account with a nickname.
 * Google sign-ups start without one and are sent to /onboarding first.
 */
export function asMember(me: SessionUser | null): Member | ActionError {
	if (!me) return "unauthorized";
	if (me.banned) return "forbidden";
	if (!me.emailVerified) return "emailNotVerified";
	if (!me.username) return "nicknameRequired";
	return { ...me, username: me.username };
}

// Better Auth's admin plugin stores roles as a comma-separated string.
export function isAdmin(user: Pick<SessionUser, "role" | "banned"> | null): boolean {
	return Boolean(user && !user.banned && user.role?.split(",").includes("admin"));
}
