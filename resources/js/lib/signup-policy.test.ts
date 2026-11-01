import { expect, test } from "vitest";

import { PASSWORD_MAX_LENGTH, PASSWORD_MIN_LENGTH } from "./password-policy";
import { isValidNickname, NICKNAME_PATTERN, normalizeNickname } from "./signup-policy";

test("nickname boundaries match the server rule", () => {
	expect(isValidNickname("abc")).toBe(false);
	expect(isValidNickname("abcd")).toBe(true);
	expect(isValidNickname("a".repeat(16))).toBe(true);
	expect(isValidNickname("a".repeat(17))).toBe(false);
	for (const bad of ["ab cd", "ab_cd", "abc1", "Abcd", "ação"]) {
		expect(isValidNickname(bad)).toBe(false);
	}
	expect(new RegExp(`^(?:${NICKNAME_PATTERN})$`).test("abcd")).toBe(true);
});

test("the nickname is lowercased and loses whitespace as it is typed", () => {
	expect(normalizeNickname(" Ana Luiza ")).toBe("analuiza");
});

test("password bounds are 8 to 32", () => {
	expect([PASSWORD_MIN_LENGTH, PASSWORD_MAX_LENGTH]).toEqual([8, 32]);
});
