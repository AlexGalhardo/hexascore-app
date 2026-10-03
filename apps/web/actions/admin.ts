"use server";

import { revalidateTag, updateTag } from "next/cache";
import { headers } from "next/headers";
import { z } from "zod";
import { type ActionResult, fail, ok } from "@/lib/action-result";
import * as admin from "@/lib/admin";
import { auth } from "@/lib/auth";
import { CATALOG_TAG, itemTag } from "@/lib/catalog/cached";
import { db } from "@/lib/db/client";
import { isAdmin } from "@/lib/member";
import { getSessionUser } from "@/lib/session";

// Every admin action re-checks the role on the server (the /admin layout check only hides pages).
// Signature (prev, FormData) so the plain <form>s in /admin work through useActionState.

type FormAction = (prev: ActionResult | null, form: FormData) => Promise<ActionResult>;

async function guarded<S extends z.ZodType>(
	schema: S,
	form: FormData,
	label: string,
	run: (input: z.infer<S>) => Promise<ActionResult>,
): Promise<ActionResult> {
	if (!isAdmin(await getSessionUser())) return fail("notFound");
	const parsed = schema.safeParse(Object.fromEntries(form));
	if (!parsed.success) return fail("invalid", [...new Set(parsed.error.issues.map((i) => String(i.path[0])))]);
	try {
		return await run(parsed.data);
	} catch (error) {
		console.error(`[admin] ${label} failed:`, error);
		return fail("server");
	}
}

const checkbox = z.literal("on").optional().transform(Boolean);
const flag = z.enum(["true", "false"]).transform((v) => v === "true");

export const setContactHandled: FormAction = async (_prev, form) =>
	guarded(z.object({ id: z.uuid(), handled: flag }), form, "contact", async ({ id, handled }) =>
		(await admin.setContactHandled(db, id, handled)) ? ok : fail("notFound"),
	);

export const setReportStatus: FormAction = async (_prev, form) =>
	guarded(
		z.object({ id: z.uuid(), status: z.enum(["resolved", "dismissed"]) }),
		form,
		"report",
		async ({ id, status }) => ((await admin.setReportStatus(db, id, status)) ? ok : fail("notFound")),
	);

export const setListHidden: FormAction = async (_prev, form) =>
	guarded(z.object({ listId: z.uuid(), hidden: flag }), form, "hide list", async ({ listId, hidden }) =>
		(await admin.setListHidden(db, listId, hidden)) ? ok : fail("notFound"),
	);

export const banUser: FormAction = async (_prev, form) =>
	guarded(
		z.object({ userId: z.string().min(1).max(128), reason: z.string().trim().min(1).max(200) }),
		form,
		"ban",
		async ({ userId, reason }) => {
			// Better Auth revokes every session of the user and blocks new sign-ins.
			await auth.api.banUser({ body: { userId, banReason: reason }, headers: await headers() });
			return ok;
		},
	);

export const saveAffiliateLink: FormAction = async (_prev, form) =>
	guarded(
		z.object({ itemId: z.uuid(), store: z.enum(admin.AFFILIATE_STORES), url: z.string(), active: checkbox }),
		form,
		"affiliate save",
		async (input) => {
			const parsed = admin.affiliateInputSchema.safeParse(input);
			if (!parsed.success) return fail("invalid", ["url"]);
			if (!(await admin.upsertAffiliateLink(db, parsed.data))) return fail("notFound");
			updateTag(itemTag(parsed.data.itemId));
			return ok;
		},
	);

export const deleteAffiliateLink: FormAction = async (_prev, form) =>
	guarded(
		z.object({ itemId: z.uuid(), store: z.enum(admin.AFFILIATE_STORES) }),
		form,
		"affiliate delete",
		async ({ itemId, store }) => {
			await admin.deleteAffiliateLink(db, itemId, store);
			updateTag(itemTag(itemId));
			return ok;
		},
	);

export const mergeItems: FormAction = async (_prev, form) =>
	guarded(z.object({ sourceId: z.uuid(), targetId: z.uuid() }), form, "merge", async ({ sourceId, targetId }) => {
		const result = await admin.mergeItems(db, sourceId, targetId);
		if (result !== "ok") return fail(result === "notFound" ? "notFound" : "invalid", ["sourceId"]);
		updateTag(itemTag(sourceId));
		updateTag(itemTag(targetId));
		revalidateTag(CATALOG_TAG, "max");
		return ok;
	});
