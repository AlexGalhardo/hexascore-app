import { z } from "zod";

export const PASSWORD_MIN = 8;
export const PASSWORD_MAX = 32;
export const NICKNAME_MIN = 4;
export const NICKNAME_MAX = 16;
export const EMAIL_INPUT_MAX = 36;
export const NICKNAME_PATTERN = /^[a-z]+$/;

/** What the nickname inputs keep while typing: lowercase letters only (no digits, spaces or symbols). */
export function normalizeNickname(value: string): string {
	return value.toLowerCase().replace(/[^a-z]/g, "");
}
export const CONTACT_MESSAGE_MIN = 32;
export const CONTACT_MESSAGE_MAX = 512;

export const passwordSchema = z
	.string()
	.min(PASSWORD_MIN, "password.min")
	.max(PASSWORD_MAX, "password.max")
	.regex(/[A-Z]/, "password.uppercase")
	.regex(/[a-z]/, "password.lowercase")
	.regex(/[0-9]/, "password.number")
	.regex(/[^A-Za-z0-9]/, "password.special");

export const nicknameSchema = z
	.string()
	.trim()
	.min(NICKNAME_MIN, "nickname.min")
	.max(NICKNAME_MAX, "nickname.max")
	.regex(NICKNAME_PATTERN, "nickname.charset");

export const emailSchema = z.email("email.invalid").max(254);
// New accounts only: existing (e.g. Google) accounts with longer addresses must still log in and recover.
export const signupEmailSchema = emailSchema.max(EMAIL_INPUT_MAX, "email.max");

export const nameSchema = z.string().trim().min(1, "name.required").max(100, "name.max");

export const CONTACT_SUBJECTS = ["question", "feedback", "bug"] as const;

export const contactSchema = z.object({
	name: nameSchema,
	email: emailSchema,
	subject: z.enum(CONTACT_SUBJECTS),
	message: z.string().trim().min(CONTACT_MESSAGE_MIN, "message.min").max(CONTACT_MESSAGE_MAX, "message.max"),
});
