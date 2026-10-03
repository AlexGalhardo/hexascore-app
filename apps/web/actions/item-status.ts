"use server";

import { refresh, updateTag } from "next/cache";
import { z } from "zod";
import { type ActionResult, fail, ok } from "@/lib/action-result";
import { itemTag } from "@/lib/catalog/cached";
import { db } from "@/lib/db/client";
import { asMember } from "@/lib/member";
import { hitRateLimit } from "@/lib/rate-limit";
import { getSessionUser } from "@/lib/session";
import { itemStatusSchema, setItemStatus } from "@/lib/social/repository";

const inputSchema = z.object({ itemId: z.uuid(), status: itemStatusSchema.nullable() });

/** `status: null` clears it. */
export async function saveItemStatus(input: unknown): Promise<ActionResult> {
	const parsed = inputSchema.safeParse(input);
	if (!parsed.success) return fail("invalid");
	const member = asMember(await getSessionUser());
	if (typeof member === "string") return fail(member);
	if (!(await hitRateLimit(db, `status:${member.id}`, 60, 60_000))) return fail("rateLimited");
	let result: Awaited<ReturnType<typeof setItemStatus>>;
	try {
		result = await setItemStatus(db, member.id, parsed.data.itemId, parsed.data.status);
	} catch (error) {
		console.error("[item-status] save failed:", error);
		return fail("server");
	}
	if (result === "notFound") return fail("notFound");
	// Leaving `completed` deleted a time-to-beat report: the item's cached medians must drop it.
	if (result === "cleared") updateTag(itemTag(parsed.data.itemId));
	refresh();
	return ok;
}
