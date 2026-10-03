import { z } from "zod";

const optional = z
	.string()
	.trim()
	.optional()
	.transform((v) => (v === "" ? undefined : v));
const optionalUrl = optional.pipe(z.url().optional());

const schema = z
	.object({
		NODE_ENV: z.enum(["development", "test", "production"]).default("development"),
		VERCEL_ENV: z.enum(["development", "preview", "production"]).optional(),
		NEXT_PUBLIC_APP_NAME: optional,
		NEXT_PUBLIC_APP_URL: optionalUrl,
		DATABASE_URL: optional.pipe(
			z
				.string()
				.regex(/^postgres(ql)?:\/\//, "must start with postgres:// or postgresql://")
				.optional(),
		),
		BETTER_AUTH_SECRET: optional.pipe(z.string().min(32, "must have at least 32 characters").optional()),
		BETTER_AUTH_URL: optionalUrl,
		GOOGLE_CLIENT_ID: optional,
		GOOGLE_CLIENT_SECRET: optional,
		UPSTASH_REDIS_REST_URL: optionalUrl,
		UPSTASH_REDIS_REST_TOKEN: optional,
		RESEND_API_KEY: optional.pipe(z.string().startsWith("re_", "must start with re_").optional()),
		EMAIL_FROM: optional,
		CONTACT_INBOX: optional.pipe(z.email().optional()),
		STRIPE_SECRET_KEY: optional.pipe(
			z
				.string()
				.regex(/^(sk|rk)_(test|live)_/, "must be a Stripe secret key")
				.optional(),
		),
		STRIPE_WEBHOOK_SECRET: optional.pipe(z.string().startsWith("whsec_", "must start with whsec_").optional()),
		ABACATEPAY_API_KEY: optional,
		ABACATEPAY_WEBHOOK_SECRET: optional,
		BLOB_READ_WRITE_TOKEN: optional,
		CRON_SECRET: optional.pipe(z.string().min(32, "must have at least 32 characters").optional()),
		TELEGRAM_BOT_TOKEN: optional,
		TELEGRAM_CHAT_ID: optional,
	})
	.superRefine((env, ctx) => {
		const pairs = [
			["GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET"],
			["UPSTASH_REDIS_REST_URL", "UPSTASH_REDIS_REST_TOKEN"],
			["STRIPE_SECRET_KEY", "STRIPE_WEBHOOK_SECRET"],
			["ABACATEPAY_API_KEY", "ABACATEPAY_WEBHOOK_SECRET"],
			["TELEGRAM_BOT_TOKEN", "TELEGRAM_CHAT_ID"],
		] as const;
		for (const [a, b] of pairs) {
			if (Boolean(env[a]) !== Boolean(env[b])) {
				ctx.addIssue({
					code: "custom",
					path: [env[a] ? b : a],
					message: `required together with ${env[a] ? a : b}`,
				});
			}
		}
		// Any production build/server (including Vercel previews, which are publicly reachable) must have its own
		// secrets: without them Better Auth would sign sessions with a known default.
		const required: (keyof typeof env)[] = [];
		if (env.NODE_ENV === "production") required.push("BETTER_AUTH_SECRET", "CRON_SECRET");
		if (env.VERCEL_ENV) required.push("DATABASE_URL");
		if (env.VERCEL_ENV === "production") required.push("BETTER_AUTH_URL");
		for (const key of required) {
			if (!env[key]) {
				ctx.addIssue({
					code: "custom",
					path: [key],
					message: `required when NODE_ENV=${env.NODE_ENV}${env.VERCEL_ENV ? ` / VERCEL_ENV=${env.VERCEL_ENV}` : ""}`,
				});
			}
		}
	});

export type Env = z.infer<typeof schema>;

function format(error: z.ZodError): string {
	return error.issues.map((i) => `  ✗ ${i.path.join(".") || "(root)"}: ${i.message}`).join("\n");
}

function load(): Env {
	const result = schema.safeParse(process.env);
	if (!result.success) {
		const message = `\n[env] Invalid environment variables:\n${format(result.error)}\n\nFix them in apps/web/.env.local (see .env.example) or in the Vercel project settings.\n`;
		console.error(message);
		throw new Error("Invalid environment variables — see the list above.");
	}
	return result.data;
}

export const env: Env = load();

export const features = {
	google: Boolean(env.GOOGLE_CLIENT_ID),
	redis: Boolean(env.UPSTASH_REDIS_REST_URL),
	email: Boolean(env.RESEND_API_KEY),
	stripe: Boolean(env.STRIPE_SECRET_KEY),
	pix: Boolean(env.ABACATEPAY_API_KEY),
	blob: Boolean(env.BLOB_READ_WRITE_TOKEN),
	telegram: Boolean(env.TELEGRAM_BOT_TOKEN),
} as const;

export function reportDisabledFeatures(): void {
	const off = Object.entries(features)
		.filter(([, on]) => !on)
		.map(([name]) => name);
	if (off.length) console.warn(`[env] Optional integrations disabled (missing keys): ${off.join(", ")}`);
}
