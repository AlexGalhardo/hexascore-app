import { passkey } from "@better-auth/passkey";
import { APIError, betterAuth } from "better-auth";
import { drizzleAdapter } from "better-auth/adapters/drizzle";
import { createAuthMiddleware } from "better-auth/api";
import { nextCookies } from "better-auth/next-js";
import { admin, twoFactor, username } from "better-auth/plugins";
import { ActionEmail } from "@/emails/action-email";
import { APP_NAME, APP_URL } from "./config";
import { db } from "./db/client";
import * as schema from "./db/schema";
import { sendEmail } from "./email";
import { env } from "./env";
import { NICKNAME_MAX, NICKNAME_MIN, PASSWORD_MAX, PASSWORD_MIN, passwordSchema } from "./validation";

// Every endpoint that accepts a new password, including the admin plugin's, enforces the same rules.
const PASSWORD_PATHS = new Set([
	"/sign-up/email",
	"/reset-password",
	"/change-password",
	"/set-password",
	"/admin/set-user-password",
	"/admin/create-user",
]);

const google =
	env.GOOGLE_CLIENT_ID && env.GOOGLE_CLIENT_SECRET
		? { google: { clientId: env.GOOGLE_CLIENT_ID, clientSecret: env.GOOGLE_CLIENT_SECRET } }
		: undefined;

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
		sendResetPassword: async ({ user, url }) => {
			await sendEmail(
				user.email,
				`${APP_NAME}: reset your password`,
				ActionEmail({
					appName: APP_NAME,
					heading: "Reset your password",
					body: "Someone asked to reset the password of this account. If it was not you, ignore this email.",
					actionLabel: "Reset password",
					url,
				}),
			);
		},
	},
	emailVerification: {
		sendOnSignUp: true,
		autoSignInAfterVerification: true,
		sendVerificationEmail: async ({ user, url }) => {
			await sendEmail(
				user.email,
				`${APP_NAME}: confirm your email`,
				ActionEmail({
					appName: APP_NAME,
					heading: "Confirm your email",
					body: "Confirm your email address to rate items and create API keys.",
					actionLabel: "Confirm email",
					url,
				}),
			);
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
			if (!PASSWORD_PATHS.has(ctx.path)) return;
			const body = ctx.body as { password?: unknown; newPassword?: unknown } | undefined;
			const candidate = body?.newPassword ?? body?.password;
			if (!passwordSchema.safeParse(candidate).success) {
				throw new APIError("BAD_REQUEST", { message: "Password does not meet the requirements." });
			}
		}),
	},
	plugins: [
		username({ minUsernameLength: NICKNAME_MIN, maxUsernameLength: NICKNAME_MAX }),
		twoFactor({ issuer: APP_NAME }),
		passkey({ rpName: APP_NAME }),
		admin(),
		nextCookies(),
	],
});

export type Session = typeof auth.$Infer.Session;
