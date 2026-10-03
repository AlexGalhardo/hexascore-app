import { passkey } from "@better-auth/passkey";
import { APIError, betterAuth } from "better-auth";
import { drizzleAdapter } from "better-auth/adapters/drizzle";
import { createAuthMiddleware } from "better-auth/api";

import { nextCookies } from "better-auth/next-js";
import { admin, twoFactor, username } from "better-auth/plugins";
import { createTranslator } from "next-intl";
import { createElement } from "react";
import { ActionEmail } from "@/emails/action-email";
import { localeFromRequest } from "@/i18n/locale";
import type { Locale } from "@/i18n/routing";
import enAuth from "@/messages/en/auth.json";
import esAuth from "@/messages/es/auth.json";
import ptAuth from "@/messages/pt-BR/auth.json";
import { BOT_PROTECTED_AUTH_PATHS, isBotRequest } from "./bot";
import { APP_NAME, APP_URL } from "./config";
import { db } from "./db/client";
import * as schema from "./db/schema";
import { sendEmail } from "./email";
import { env } from "./env";
import {
	NICKNAME_MAX,
	NICKNAME_MIN,
	NICKNAME_PATTERN,
	PASSWORD_MAX,
	PASSWORD_MIN,
	passwordSchema,
	signupEmailSchema,
} from "./validation";

// Every endpoint that accepts a new password, including the admin plugin's, enforces the same rules.
const PASSWORD_PATHS = new Set([
	"/sign-up/email",
	"/reset-password",
	"/change-password",
	"/set-password",
	"/admin/set-user-password",
	"/admin/create-user",
]);

const BOT_PATHS = new Set<string>(BOT_PROTECTED_AUTH_PATHS);

const google =
	env.GOOGLE_CLIENT_ID && env.GOOGLE_CLIENT_SECRET
		? { google: { clientId: env.GOOGLE_CLIENT_ID, clientSecret: env.GOOGLE_CLIENT_SECRET } }
		: undefined;

const EMAIL_MESSAGES = { en: enAuth, "pt-BR": ptAuth, es: esAuth } satisfies Record<Locale, typeof enAuth>;

// Auth emails are sent from /api/auth/*, outside next-intl's request scope.
async function sendActionEmail(to: string, kind: "verify" | "reset", url: string, request?: Request): Promise<void> {
	const locale = localeFromRequest(request);
	const t = createTranslator({ locale, messages: EMAIL_MESSAGES[locale], namespace: `auth.emails.${kind}` });
	await sendEmail(
		to,
		`${APP_NAME}: ${t("subject")}`,
		// createElement (not a direct call) keeps `url` on the element props, which the dev email log prints.
		createElement(ActionEmail, {
			appName: APP_NAME,
			heading: t("heading"),
			body: t("body"),
			actionLabel: t("action"),
			url,
		}),
	);
}

export const auth = betterAuth({
	appName: APP_NAME,
	baseURL: env.BETTER_AUTH_URL ?? (process.env.VERCEL_URL ? `https://${process.env.VERCEL_URL}` : APP_URL),
	secret: env.BETTER_AUTH_SECRET,
	database: drizzleAdapter(db, { provider: "pg", schema }),
	emailAndPassword: {
		enabled: true,
		requireEmailVerification: true,
		minPasswordLength: PASSWORD_MIN,
		maxPasswordLength: PASSWORD_MAX,
		revokeSessionsOnPasswordReset: true,
		sendResetPassword: async ({ user, url }, request) => {
			await sendActionEmail(user.email, "reset", url, request);
		},
	},
	emailVerification: {
		sendOnSignUp: true,
		autoSignInAfterVerification: true,
		sendVerificationEmail: async ({ user, url }, request) => {
			await sendActionEmail(user.email, "verify", url, request);
		},
	},
	socialProviders: google,
	user: {
		additionalFields: {
			profilePrivate: { type: "boolean", defaultValue: false, input: false },
			deletionScheduledAt: { type: "date", required: false, input: false },
		},
	},
	rateLimit: { enabled: true, storage: "database" },
	hooks: {
		before: createAuthMiddleware(async (ctx) => {
			// Server-side auth.api calls carry no browser request, so only real HTTP requests are bot-checked.
			if (BOT_PATHS.has(ctx.path) && ctx.request && (await isBotRequest(ctx.headers))) {
				throw new APIError("FORBIDDEN", { message: "Request blocked.", code: "BOT_DETECTED" });
			}
			if (
				ctx.path === "/sign-up/email" &&
				!signupEmailSchema.safeParse((ctx.body as { email?: unknown })?.email).success
			) {
				throw new APIError("BAD_REQUEST", { message: "Invalid email.", code: "INVALID_EMAIL" });
			}
			if (!PASSWORD_PATHS.has(ctx.path)) return;
			const body = ctx.body as { password?: unknown; newPassword?: unknown } | undefined;
			const candidate = body?.newPassword ?? body?.password;
			if (!passwordSchema.safeParse(candidate).success) {
				throw new APIError("BAD_REQUEST", {
					message: "Password does not meet the requirements.",
					code: "WEAK_PASSWORD",
				});
			}
		}),
	},
	plugins: [
		username({
			minUsernameLength: NICKNAME_MIN,
			maxUsernameLength: NICKNAME_MAX,
			usernameValidator: (value) => NICKNAME_PATTERN.test(value),
		}),
		// Passwordless lets Google/passkey-only accounts use 2FA; accounts with a password must still confirm it.
		twoFactor({ issuer: APP_NAME, allowPasswordless: true }),
		passkey({ rpName: APP_NAME }),
		admin(),
		nextCookies(),
	],
});

export type Session = typeof auth.$Infer.Session;
