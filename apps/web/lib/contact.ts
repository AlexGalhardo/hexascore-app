import type { z } from "zod";
import { type ActionResult, fail, ok } from "./action-result";
import { contactSchema } from "./validation";

export type ContactMessage = z.infer<typeof contactSchema>;

export const CONTACT_LIMIT = { max: 5, windowMs: 60 * 60 * 1000 } as const;

export type ContactDeps = {
	user: { id: string; name: string; email: string } | null;
	ip: string;
	isBot: () => Promise<boolean>;
	allow: (key: string) => Promise<boolean>;
	store: (message: ContactMessage, userId: string | null) => Promise<void>;
	notify: (message: ContactMessage) => Promise<void>;
};

// Pure orchestration so the authorization rules are unit-testable without Next or a database.
export async function submitContact(form: FormData, deps: ContactDeps): Promise<ActionResult> {
	if (await deps.isBot()) return fail("bot");

	// A logged-in user cannot speak for someone else: name/email always come from the session.
	const parsed = contactSchema.safeParse({
		name: deps.user?.name ?? form.get("name"),
		email: deps.user?.email ?? form.get("email"),
		subject: form.get("subject"),
		message: form.get("message"),
	});
	if (!parsed.success) {
		return fail(
			"invalid",
			parsed.error.issues.map((i) => String(i.path[0])),
		);
	}

	const key = deps.user ? `contact:user:${deps.user.id}` : `contact:ip:${deps.ip}`;
	if (!(await deps.allow(key))) return fail("rateLimited");

	try {
		await deps.store(parsed.data, deps.user?.id ?? null);
	} catch (error) {
		console.error("[contact] failed to store message:", error);
		return fail("server");
	}
	// The message is already stored: a failed notification must not make the user resend it.
	await deps.notify(parsed.data).catch((error: unknown) => console.error("[contact] notify failed:", error));
	return ok;
}
