import { PASSWORD_MAX, PASSWORD_MIN } from "./validation";

export const PASSWORD_RULES = ["length", "uppercase", "lowercase", "number", "special"] as const;
export type PasswordRule = (typeof PASSWORD_RULES)[number];

// Mirrors passwordSchema rule by rule so the live checklist and the server never disagree.
const tests: Record<PasswordRule, (value: string) => boolean> = {
	length: (v) => v.length >= PASSWORD_MIN && v.length <= PASSWORD_MAX,
	uppercase: (v) => /[A-Z]/.test(v),
	lowercase: (v) => /[a-z]/.test(v),
	number: (v) => /[0-9]/.test(v),
	special: (v) => /[^A-Za-z0-9]/.test(v),
};

export function checkPassword(value: string): Record<PasswordRule, boolean> {
	return Object.fromEntries(PASSWORD_RULES.map((rule) => [rule, tests[rule](value)])) as Record<
		PasswordRule,
		boolean
	>;
}

export function isPasswordValid(value: string): boolean {
	return PASSWORD_RULES.every((rule) => tests[rule](value));
}
