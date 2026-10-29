import { createHmac } from "node:crypto";

const ALPHABET = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567";
const STEP_SECONDS = 30;

function base32Decode(input: string): Uint8Array {
	let bits = "";
	for (const char of input.replace(/[\s=]/g, "").toUpperCase()) {
		const index = ALPHABET.indexOf(char);
		if (index < 0) throw new Error(`invalid base32 character ${char}`);
		bits += index.toString(2).padStart(5, "0");
	}
	const bytes: number[] = [];
	for (let i = 0; i + 8 <= bits.length; i += 8)
		bytes.push(Number.parseInt(bits.slice(i, i + 8), 2));
	return Uint8Array.from(bytes);
}

export function currentStep(now = Date.now()): number {
	return Math.floor(now / 1000 / STEP_SECONDS);
}

// RFC 6238 (HMAC-SHA1, 30 s step, 6 digits), the defaults every authenticator app uses.
export function totp(secret: string, step = currentStep()): string {
	const counter = new DataView(new ArrayBuffer(8));
	counter.setBigUint64(0, BigInt(step));
	const digest = createHmac("sha1", base32Decode(secret))
		.update(new Uint8Array(counter.buffer))
		.digest();
	const offset = (digest[digest.length - 1] ?? 0) & 0x0f;
	const binary = new DataView(digest.buffer, digest.byteOffset).getUint32(offset) & 0x7fffffff;
	return String(binary % 1_000_000).padStart(6, "0");
}

// The server refuses a code from a step it already accepted, so a second login waits for the next step.
export async function codeAfter(secret: string, usedStep: number): Promise<string> {
	while (currentStep() <= usedStep) await new Promise((resolve) => setTimeout(resolve, 500));
	return totp(secret);
}
