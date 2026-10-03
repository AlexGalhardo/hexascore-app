import type { ReactElement } from "react";
import { Resend } from "resend";
import { env } from "./env";

const resend = env.RESEND_API_KEY ? new Resend(env.RESEND_API_KEY) : null;
const from = env.EMAIL_FROM ?? "Hexascore <no-reply@example.com>";

export async function sendEmail(to: string, subject: string, react: ReactElement): Promise<void> {
	if (!resend) {
		// Local dev without Resend: links appear in the server log instead.
		console.info(`[email] to=${to} subject="${subject}"`, JSON.stringify(react.props));
		return;
	}
	const { error } = await resend.emails.send({ from, to, subject, react });
	if (error) throw new Error(`Resend failed: ${error.message}`);
}
