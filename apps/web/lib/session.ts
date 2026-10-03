import { headers } from "next/headers";
import { cache } from "react";
import { auth } from "./auth";

export type SessionUser = {
	id: string;
	name: string;
	email: string;
	username: string | null;
	emailVerified: boolean;
	role: string | null;
	banned: boolean;
	image: string | null;
	twoFactorEnabled: boolean;
	profilePrivate: boolean;
	deletionScheduledAt: Date | null;
};

// The only source of "who is acting": never trust a user id from a request body.
export const getSessionUser = cache(async (): Promise<SessionUser | null> => {
	const session = await auth.api.getSession({ headers: await headers() });
	if (!session) return null;
	const u = session.user;
	return {
		id: u.id,
		name: u.name,
		email: u.email,
		username: u.username ?? null,
		emailVerified: u.emailVerified,
		role: u.role ?? null,
		banned: u.banned ?? false,
		image: u.image ?? null,
		twoFactorEnabled: u.twoFactorEnabled ?? false,
		profilePrivate: u.profilePrivate ?? false,
		deletionScheduledAt: u.deletionScheduledAt ? new Date(u.deletionScheduledAt) : null,
	};
});
