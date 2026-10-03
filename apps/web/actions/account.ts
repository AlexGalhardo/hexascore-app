"use server";

import { APIError } from "better-auth/api";
import { eq } from "drizzle-orm";
import { refresh } from "next/cache";
import { headers } from "next/headers";
import { z } from "zod";
import { type ActionResult, fail, ok } from "@/lib/action-result";
import { auth } from "@/lib/auth";
import { db } from "@/lib/db/client";
import { session, user } from "@/lib/db/schema";
import { getSessionUser } from "@/lib/session";
import { nameSchema, passwordSchema } from "@/lib/validation";

export async function updateName(name: unknown): Promise<ActionResult> {
	const me = await getSessionUser();
	if (!me) return fail("unauthorized");
	const parsed = nameSchema.safeParse(name);
	if (!parsed.success) return fail("invalid", ["name"]);
	try {
		await auth.api.updateUser({ body: { name: parsed.data }, headers: await headers() });
	} catch (error) {
		console.error("[account] updateName failed:", error);
		return fail("server");
	}
	refresh();
	return ok;
}

// Better Auth exposes setPassword only on the server, for accounts created through Google or passkeys.
export async function setPassword(newPassword: unknown): Promise<ActionResult> {
	const me = await getSessionUser();
	if (!me) return fail("unauthorized");
	const parsed = passwordSchema.safeParse(newPassword);
	if (!parsed.success) return fail("weakPassword", ["password"]);
	try {
		await auth.api.setPassword({ body: { newPassword: parsed.data }, headers: await headers() });
	} catch (error) {
		if (error instanceof APIError && error.body?.code === "PASSWORD_ALREADY_SET") return fail("passwordAlreadySet");
		console.error("[account] setPassword failed:", error);
		return fail("server");
	}
	refresh();
	return ok;
}

export async function setProfilePrivate(value: unknown): Promise<ActionResult> {
	const me = await getSessionUser();
	if (!me) return fail("unauthorized");
	const parsed = z.boolean().safeParse(value);
	if (!parsed.success) return fail("invalid");
	try {
		await db.update(user).set({ profilePrivate: parsed.data }).where(eq(user.id, me.id));
	} catch (error) {
		console.error("[account] setProfilePrivate failed:", error);
		return fail("server");
	}
	refresh();
	return ok;
}

// Starts the 30-day grace period and signs the user out everywhere; logging back in shows the cancel banner.
export async function scheduleDeletion(): Promise<ActionResult> {
	const me = await getSessionUser();
	if (!me) return fail("unauthorized");
	try {
		await db.update(user).set({ deletionScheduledAt: new Date() }).where(eq(user.id, me.id));
		await auth.api.signOut({ headers: await headers() });
		await db.delete(session).where(eq(session.userId, me.id));
	} catch (error) {
		console.error("[account] scheduleDeletion failed:", error);
		return fail("server");
	}
	return ok;
}

export async function cancelDeletion(): Promise<ActionResult> {
	const me = await getSessionUser();
	if (!me) return fail("unauthorized");
	if (!me.deletionScheduledAt) return fail("notScheduled");
	try {
		await db.update(user).set({ deletionScheduledAt: null }).where(eq(user.id, me.id));
	} catch (error) {
		console.error("[account] cancelDeletion failed:", error);
		return fail("server");
	}
	refresh();
	return ok;
}
