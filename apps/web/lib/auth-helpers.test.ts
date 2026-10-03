import { describe, expect, test } from "bun:test";
import { authErrorKey } from "./auth-errors";
import { isAuthorizedCron } from "./cron-auth";
import { checkPassword, isPasswordValid } from "./password-rules";
import { safeNextPath } from "./safe-next";
import { passwordSchema } from "./validation";

describe("checkPassword", () => {
	test("reports each rule independently", () => {
		expect(checkPassword("")).toEqual({
			length: false,
			uppercase: false,
			lowercase: false,
			number: false,
			special: false,
		});
		expect(checkPassword("abc")).toMatchObject({ lowercase: true, uppercase: false, length: false });
		expect(checkPassword("Abcdef1!")).toEqual({
			length: true,
			uppercase: true,
			lowercase: true,
			number: true,
			special: true,
		});
	});

	test.each([
		"Abcdef1!",
		"Abcde1!",
		`Ab1!${"x".repeat(28)}`,
		`Ab1!${"x".repeat(29)}`,
		"abcdef1!",
		"ABCDEF1!",
		"Abcdefg!",
		"Abcdefg1",
		"Ação1234!",
	])("agrees with passwordSchema for %s", (value) => {
		expect(isPasswordValid(value)).toBe(passwordSchema.safeParse(value).success);
	});
});

describe("isAuthorizedCron", () => {
	const secret = "a".repeat(64);
	test.each([
		[`Bearer ${secret}`, secret, true],
		[`Bearer ${secret}x`, secret, false],
		[secret, secret, false],
		[null, secret, false],
		[`Bearer ${secret}`, undefined, false],
		["Bearer ", "", false],
	])("%s", (header, configured, expected) => {
		expect(isAuthorizedCron(header, configured)).toBe(expected);
	});
});

describe("authErrorKey", () => {
	test.each([
		[{ status: 429, code: "INVALID_EMAIL_OR_PASSWORD" }, "rateLimited"],
		[{ status: 401, code: "INVALID_EMAIL_OR_PASSWORD" }, "invalidCredentials"],
		[{ status: 401, code: "INVALID_USERNAME_OR_PASSWORD" }, "invalidCredentials"],
		[{ status: 403, code: "EMAIL_NOT_VERIFIED" }, "emailNotVerified"],
		[{ status: 403, code: "BOT_DETECTED" }, "bot"],
		[{ status: 400, code: "WEAK_PASSWORD" }, "weakPassword"],
		[{ status: 0 }, "network"],
		[{ status: 500, code: "SOMETHING_NEW" }, "generic"],
		[null, "generic"],
	] as const)("%o -> %s", (error, key) => {
		expect(authErrorKey(error)).toBe(key);
	});
});

describe("safeNextPath", () => {
	test.each([
		["/profile", "/profile"],
		["/pt-br/profile?tab=2", "/pt-br/profile?tab=2"],
		["//evil.com", "/"],
		["/\\evil.com", "/"],
		["https://evil.com", "/"],
		[null, "/"],
		["", "/"],
	])("%s -> %s", (value, expected) => {
		expect(safeNextPath(value)).toBe(expected);
	});
});
