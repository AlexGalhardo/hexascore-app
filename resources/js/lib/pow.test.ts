import { expect, test } from "vitest";

import { leadingZeroBits, sha256, solve } from "./pow-worker";

const hex = (d: Uint32Array) => Array.from(d, (w) => w.toString(16).padStart(8, "0")).join("");

test("sha256 matches the known vectors", () => {
	expect(hex(sha256(new TextEncoder().encode("abc")))).toBe(
		"ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad",
	);
	expect(hex(sha256(new Uint8Array(0)))).toBe(
		"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
	);
});

test("solve returns a nonce with the requested leading zero bits", () => {
	const n = solve("challenge.sig", 10);
	const digest = sha256(new TextEncoder().encode(`challenge.sig${n}`));
	expect(leadingZeroBits(digest)).toBeGreaterThanOrEqual(10);
});

test("solve stays correct as the nonce grows past several digit counts", () => {
	const n = solve("boundary.sig", 16);
	expect(Number(n)).toBeGreaterThan(9999);
	const digest = sha256(new TextEncoder().encode(`boundary.sig${n}`));
	expect(leadingZeroBits(digest)).toBeGreaterThanOrEqual(16);
});
