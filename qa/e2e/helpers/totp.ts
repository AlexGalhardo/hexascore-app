import { createHmac } from "node:crypto";

const BASE32 = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567";

function base32Decode(input: string): Buffer {
	const clean = input.replace(/=+$/, "").replace(/\s/g, "").toUpperCase();
	let bits = 0;
	let value = 0;
	const out: number[] = [];
	for (const char of clean) {
		const idx = BASE32.indexOf(char);
		if (idx < 0) throw new Error(`invalid base32 character: ${char}`);
		value = (value << 5) | idx;
		bits += 5;
		if (bits >= 8) {
			out.push((value >>> (bits - 8)) & 0xff);
			bits -= 8;
		}
	}
	return Buffer.from(out);
}

/** RFC 6238 TOTP (SHA-1, 30 s, 6 digits) — what authenticator apps compute from the secret shown on /profile. */
export function totp(secret: string, at = Date.now()): string {
	const counter = Buffer.alloc(8);
	counter.writeBigUInt64BE(BigInt(Math.floor(at / 30_000)));
	const hmac = createHmac("sha1", base32Decode(secret)).update(counter).digest();
	const offset = (hmac[hmac.length - 1] as number) & 0x0f;
	const code = (hmac.readUInt32BE(offset) & 0x7fffffff) % 1_000_000;
	return code.toString().padStart(6, "0");
}

/** Waits out the last seconds of a window so a code is not rejected for expiring mid-request. */
export async function freshTotp(secret: string): Promise<string> {
	const left = 30_000 - (Date.now() % 30_000);
	if (left < 4_000) await new Promise((resolve) => setTimeout(resolve, left + 250));
	return totp(secret);
}
