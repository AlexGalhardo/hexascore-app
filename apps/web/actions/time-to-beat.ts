"use server";

import { updateTag } from "next/cache";
import { z } from "zod";
import { type ActionResult, fail, ok } from "@/lib/action-result";
import { itemTag } from "@/lib/catalog/cached";
import { db } from "@/lib/db/client";
import { asMember } from "@/lib/member";
import { hitRateLimit } from "@/lib/rate-limit";
import { getSessionUser } from "@/lib/session";
import { removeTimeToBeat, timeToBeatInputSchema, writeTimeToBeat } from "@/lib/time-to-beat";

async function guard(): Promise<string | ActionResult> {
	const member = asMember(await getSessionUser());
	if (typeof member === "string") return fail(member);
	if (!(await hitRateLimit(db, `ttb:${member.id}`, 30, 60_000))) return fail("rateLimited");
	return member.id;
}

export async function saveTimeToBeat(input: unknown): Promise<ActionResult> {
	const parsed = timeToBeatInputSchema.safeParse(input);
	if (!parsed.success) return fail("invalid");
	const userId = await guard();
	if (typeof userId !== "string") return userId;
	let result: Awaited<ReturnType<typeof writeTimeToBeat>>;
	try {
		result = await writeTimeToBeat(db, userId, parsed.data);
	} catch (error) {
		console.error("[time-to-beat] save failed:", error);
		return fail("server");
	}
	if (result !== "ok") return fail(result);
	// Medians live in the cached item detail; listings do not show them, so `catalog` stays untouched.
	updateTag(itemTag(parsed.data.itemId));
	return ok;
}

export async function deleteTimeToBeat(itemId: unknown): Promise<ActionResult> {
	const parsed = z.uuid().safeParse(itemId);
	if (!parsed.success) return fail("invalid");
	const userId = await guard();
	if (typeof userId !== "string") return userId;
	try {
		await removeTimeToBeat(db, userId, parsed.data);
	} catch (error) {
		console.error("[time-to-beat] delete failed:", error);
		return fail("server");
	}
	updateTag(itemTag(parsed.data));
	return ok;
}
