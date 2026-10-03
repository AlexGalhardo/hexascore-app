"use server";

import { and, count, eq, isNull } from "drizzle-orm";
import { revalidatePath } from "next/cache";
import { z } from "zod";
import { generateApiKey } from "@/lib/api/keys";
import { db } from "@/lib/db/client";
import { apiKeys } from "@/lib/db/schema";
import { publicApiSignupFlag } from "@/lib/flags";
import { getSessionUser } from "@/lib/session";

const MAX_ACTIVE_KEYS = 5;

export type CreateKeyResult =
	| { ok: true; key: string }
	| { ok: false; error: "unauthenticated" | "unverified" | "invalid" | "limit" | "disabled" | "server" };

const nameSchema = z.string().trim().min(1).max(40);

export async function createApiKeyAction(name: string): Promise<CreateKeyResult> {
	const user = await getSessionUser();
	if (!user) return { ok: false, error: "unauthenticated" };
	if (!user.emailVerified) return { ok: false, error: "unverified" };
	if (!(await publicApiSignupFlag())) return { ok: false, error: "disabled" };
	const parsed = nameSchema.safeParse(name);
	if (!parsed.success) return { ok: false, error: "invalid" };

	try {
		const [active] = await db
			.select({ n: count() })
			.from(apiKeys)
			.where(and(eq(apiKeys.userId, user.id), isNull(apiKeys.revokedAt)));
		if ((active?.n ?? 0) >= MAX_ACTIVE_KEYS) return { ok: false, error: "limit" };

		const { key, prefix, hash } = generateApiKey();
		await db.insert(apiKeys).values({ userId: user.id, name: parsed.data, prefix, hash });
		revalidatePath("/[locale]/profile/api", "page");
		return { ok: true, key };
	} catch (error) {
		console.error("[api-keys] create failed", error);
		return { ok: false, error: "server" };
	}
}

export type RevokeKeyResult = { ok: true } | { ok: false; error: "unauthenticated" | "not_found" | "server" };

export async function revokeApiKeyAction(id: string): Promise<RevokeKeyResult> {
	const user = await getSessionUser();
	if (!user) return { ok: false, error: "unauthenticated" };
	if (!z.uuid().safeParse(id).success) return { ok: false, error: "not_found" };
	try {
		const updated = await db
			.update(apiKeys)
			.set({ revokedAt: new Date() })
			.where(and(eq(apiKeys.id, id), eq(apiKeys.userId, user.id), isNull(apiKeys.revokedAt)))
			.returning({ id: apiKeys.id });
		if (!updated.length) return { ok: false, error: "not_found" };
		revalidatePath("/[locale]/profile/api", "page");
		return { ok: true };
	} catch (error) {
		console.error("[api-keys] revoke failed", error);
		return { ok: false, error: "server" };
	}
}
