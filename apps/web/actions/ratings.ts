"use server";

import { revalidateTag, updateTag } from "next/cache";
import { z } from "zod";
import { type ActionResult, fail, ok } from "@/lib/action-result";
import { CATALOG_TAG, itemTag } from "@/lib/catalog/cached";
import { db } from "@/lib/db/client";
import { asMember } from "@/lib/member";
import { hitRateLimit } from "@/lib/rate-limit";
import { ratingInputSchema, removeRating, writeRating } from "@/lib/ratings";
import { getSessionUser } from "@/lib/session";

const RATE = { max: 30, windowMs: 60_000 };

function refreshItem(itemId: string): void {
	// The rater sees the new metascore right away; listings and the landing refresh in the background.
	updateTag(itemTag(itemId));
	revalidateTag(CATALOG_TAG, "max");
}

async function guard(): Promise<string | ActionResult> {
	const member = asMember(await getSessionUser());
	if (typeof member === "string") return fail(member);
	if (!(await hitRateLimit(db, `rating:${member.id}`, RATE.max, RATE.windowMs))) return fail("rateLimited");
	return member.id;
}

export async function saveRating(input: unknown): Promise<ActionResult> {
	const parsed = ratingInputSchema.safeParse(input);
	if (!parsed.success) return fail("invalid");
	const userId = await guard();
	if (typeof userId !== "string") return userId;
	try {
		if (!(await writeRating(db, userId, parsed.data))) return fail("notFound");
	} catch (error) {
		console.error("[ratings] save failed:", error);
		return fail("server");
	}
	refreshItem(parsed.data.itemId);
	return ok;
}

export async function deleteRating(itemId: unknown): Promise<ActionResult> {
	const parsed = z.uuid().safeParse(itemId);
	if (!parsed.success) return fail("invalid");
	const userId = await guard();
	if (typeof userId !== "string") return userId;
	try {
		if (!(await removeRating(db, userId, parsed.data))) return fail("notFound");
	} catch (error) {
		console.error("[ratings] delete failed:", error);
		return fail("server");
	}
	refreshItem(parsed.data);
	return ok;
}
