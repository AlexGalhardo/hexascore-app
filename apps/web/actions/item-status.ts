"use server";

import { refresh } from "next/cache";
import { z } from "zod";
import { type ActionResult, fail, ok } from "@/lib/action-result";
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
	try {
		if (!(await setItemStatus(db, member.id, parsed.data.itemId, parsed.data.status))) return fail("notFound");
	} catch (error) {
		console.error("[item-status] save failed:", error);
		return fail("server");
	}
	refresh();
	return ok;
}
