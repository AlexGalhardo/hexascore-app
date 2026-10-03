"use server";

import { refresh } from "next/cache";
import { z } from "zod";
import { type ActionFailure, type ActionResult, fail, ok } from "@/lib/action-result";
import { db } from "@/lib/db/client";
import { asMember, type Member } from "@/lib/member";
import { hitRateLimit } from "@/lib/rate-limit";
import { getSessionUser } from "@/lib/session";
import * as lists from "@/lib/social/lists";

async function member(): Promise<Member | ActionFailure> {
	const me = asMember(await getSessionUser());
	if (typeof me === "string") return fail(me);
	if (!(await hitRateLimit(db, `lists:${me.id}`, 60, 60_000))) return fail("rateLimited");
	return me;
}

async function run(label: string, write: () => Promise<boolean | ActionResult>): Promise<ActionResult> {
	try {
		const result = await write();
		if (result === false) return fail("notFound");
		if (result !== true && !result.ok) return result;
	} catch (error) {
		console.error(`[lists] ${label} failed:`, error);
		return fail("server");
	}
	refresh();
	return ok;
}

const listId = z.uuid();
const entrySchema = z.object({ listId, itemId: z.uuid() });

function invalidFields(error: z.ZodError): string[] {
	return [...new Set(error.issues.map((i) => String(i.path[0])))];
}

export async function createList(input: unknown, firstItemId?: unknown): Promise<ActionResult<{ id: string }>> {
	const parsed = lists.createListInputSchema.safeParse(input);
	if (!parsed.success) return fail("invalid", invalidFields(parsed.error));
	const item = z.uuid().optional().safeParse(firstItemId);
	if (!item.success) return fail("invalid");
	const me = await member();
	if (!("id" in me)) return me;
	try {
		const created = await lists.createList(db, me.id, parsed.data, item.data);
		if (typeof created === "string") return fail(created);
		refresh();
		return { ok: true, id: created.id };
	} catch (error) {
		console.error("[lists] create failed:", error);
		return fail("server");
	}
}

export async function updateList(id: unknown, input: unknown): Promise<ActionResult> {
	const parsedId = listId.safeParse(id);
	const parsed = lists.listInputSchema.safeParse(input);
	if (!parsedId.success) return fail("invalid");
	if (!parsed.success) return fail("invalid", invalidFields(parsed.error));
	const me = await member();
	if (!("id" in me)) return me;
	return run("update", () => lists.updateList(db, me.id, parsedId.data, parsed.data));
}

export async function deleteList(id: unknown): Promise<ActionResult> {
	const parsedId = listId.safeParse(id);
	if (!parsedId.success) return fail("invalid");
	const me = await member();
	if (!("id" in me)) return me;
	return run("delete", () => lists.deleteList(db, me.id, parsedId.data));
}

export async function addToList(input: unknown): Promise<ActionResult> {
	const parsed = entrySchema.safeParse(input);
	if (!parsed.success) return fail("invalid");
	const me = await member();
	if (!("id" in me)) return me;
	return run("add", async () => {
		const result = await lists.addListItem(db, me.id, parsed.data.listId, parsed.data.itemId);
		return result === "ok" ? true : result === "notFound" ? false : fail(result);
	});
}

export async function removeFromList(input: unknown): Promise<ActionResult> {
	const parsed = entrySchema.safeParse(input);
	if (!parsed.success) return fail("invalid");
	const me = await member();
	if (!("id" in me)) return me;
	return run("remove", () => lists.removeListItem(db, me.id, parsed.data.listId, parsed.data.itemId));
}

const moveSchema = entrySchema.extend({ direction: z.enum(["up", "down"]) });

export async function moveInList(input: unknown): Promise<ActionResult> {
	const parsed = moveSchema.safeParse(input);
	if (!parsed.success) return fail("invalid");
	const me = await member();
	if (!("id" in me)) return me;
	const { listId: id, itemId, direction } = parsed.data;
	return run("move", () => lists.moveListItem(db, me.id, id, itemId, direction));
}
