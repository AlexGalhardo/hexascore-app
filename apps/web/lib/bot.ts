import { checkBotId } from "botid/server";

// Paths the browser calls through BotID (instrumentation-client.ts) and the server re-checks.
export const BOT_PROTECTED_AUTH_PATHS = ["/sign-up/email", "/sign-in/email", "/sign-in/username"] as const;

const onVercel = process.env.VERCEL === "1";

// Off Vercel (local dev, `next start`, tests) BotID cannot classify and bypasses as human.
// On Vercel any failure of the check denies the request (fails closed).
export async function isBotRequest(headers?: Headers): Promise<boolean> {
	try {
		const result = await checkBotId({
			developmentOptions: { isDevelopment: !onVercel },
			advancedOptions: headers ? { headers: Object.fromEntries(headers.entries()) } : undefined,
		});
		return result.isBot;
	} catch (error) {
		console.error("[botid] verification failed:", error);
		return onVercel;
	}
}
