import type { ReactElement } from "react";
import { Resend } from "resend";
import { env } from "./env";

const resend = env.RESEND_API_KEY ? new Resend(env.RESEND_API_KEY) : null;
const from = env.EMAIL_FROM ?? "Hexascore <no-reply@example.com>";

export async function sendEmail(to: string, subject: string, react: ReactElement): Promise<void> {
	if (!resend) {
		// Emails carry one-time tokens (verification, password reset): only print them on a developer machine.
		if (env.NODE_ENV === "production") {
			throw new Error("Email delivery is not configured (RESEND_API_KEY missing).");
		}
		const { url } = react.props as { url?: string };
		console.info(`[email:dev] to=${to} subject="${subject}" link=${url ?? "(none)"}`);
		return;
	}
	const { error } = await resend.emails.send({ from, to, subject, react });
	if (error) throw new Error(`Resend failed: ${error.name}: ${error.message}`);
}
