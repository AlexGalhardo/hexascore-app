import { describe, expect, test } from "bun:test";
import { nicknameSchema, passwordSchema } from "./validation";

describe("passwordSchema", () => {
	test.each([
		["Abcdef1!", true],
		["Abcde1!", false],
		[`Ab1!${"x".repeat(28)}`, true],
		[`Ab1!${"x".repeat(29)}`, false],
		["abcdef1!", false],
		["ABCDEF1!", false],
		["Abcdefg!", false],
		["Abcdefg1", false],
	])("%s -> %p", (value, ok) => {
		expect(passwordSchema.safeParse(value).success).toBe(ok);
	});
});

describe("nicknameSchema", () => {
	test.each([
		["abc", false],
		["abcd", true],
		["a".repeat(16), true],
		["a".repeat(17), false],
		["bad name", false],
		["good_name1", false],
		["Goodname", false],
		["goodname", true],
		["name1234", false],
	])("%s -> %p", (value, ok) => {
		expect(nicknameSchema.safeParse(value).success).toBe(ok);
	});
});
