import { expect, test } from "bun:test";
import { isCronAuthorized } from "./cron-auth";

const secret = "a".repeat(64);

test.each([
	[`Bearer ${secret}`, secret, true],
	[`Bearer ${secret}x`, secret, false],
	[`bearer ${secret}`, secret, false],
	[secret, secret, false],
	[null, secret, false],
	["Bearer ", "", false],
	[`Bearer ${secret}`, undefined, false],
] as const)("%p -> %p", (header, configured, ok) => {
	expect(isCronAuthorized(header, configured)).toBe(ok);
});
