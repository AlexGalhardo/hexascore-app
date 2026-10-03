"use server";

import { refresh } from "next/cache";
import { z } from "zod";
import { type ActionResult, fail, ok } from "@/lib/action-result";
import { db } from "@/lib/db/client";
import { asMember } from "@/lib/member";
import { hitRateLimit } from "@/lib/rate-limit";
import { getSessionUser } from "@/lib/session";
import { setFollow } from "@/lib/social/repository";

const inputSchema = z.object({ nickname: z.string().trim().min(1).max(64), follow: z.boolean() });

export async function toggleFollow(input: unknown): Promise<ActionResult> {
	const parsed = inputSchema.safeParse(input);
	if (!parsed.success) return fail("invalid");
	const member = asMember(await getSessionUser());
	if (typeof member === "string") return fail(member);
	if (!(await hitRateLimit(db, `follow:${member.id}`, 30, 60_000))) return fail("rateLimited");
	try {
		const result = await setFollow(db, member.id, parsed.data.nickname, parsed.data.follow);
		if (result === "self") return fail("forbidden");
		if (result === "notFound") return fail("notFound");
	} catch (error) {
		console.error("[follows] toggle failed:", error);
		return fail("server");
	}
	refresh();
	return ok;
}
