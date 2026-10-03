"use server";

import { headers } from "next/headers";
import { ContactEmail } from "@/emails/contact-email";
import type { ActionResult } from "@/lib/action-result";
import { isBotRequest } from "@/lib/bot";
import { APP_NAME } from "@/lib/config";
import { CONTACT_LIMIT, submitContact } from "@/lib/contact";
import { db } from "@/lib/db/client";
import { contactMessages } from "@/lib/db/schema";
import { sendEmail } from "@/lib/email";
import { env } from "@/lib/env";
import { clientIp, hitRateLimit } from "@/lib/rate-limit";
import { getSessionUser } from "@/lib/session";

export async function sendContactMessage(_prev: ActionResult | null, form: FormData): Promise<ActionResult> {
	const requestHeaders = await headers();
	return submitContact(form, {
		user: await getSessionUser(),
		ip: clientIp(requestHeaders),
		isBot: () => isBotRequest(),
		allow: (key) => hitRateLimit(db, key, CONTACT_LIMIT.max, CONTACT_LIMIT.windowMs),
		store: async (message, userId) => {
			await db.insert(contactMessages).values({ ...message, userId });
		},
		notify: async (message) => {
			if (!env.CONTACT_INBOX) return;
			await sendEmail(env.CONTACT_INBOX, `${APP_NAME} contact: ${message.subject}`, ContactEmail(message));
		},
	});
}
