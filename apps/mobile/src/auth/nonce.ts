type RandomSource = { getRandomValues?: (bytes: Uint8Array<ArrayBuffer>) => unknown };

/**
 * A 128-bit hex nonce from the platform CSPRNG. Returns null when none exists: the caller must not sign in
 * without one (Math.random is predictable, so there is no fallback).
 */
export function secureNonce(source: RandomSource | undefined = globalThis.crypto): string | null {
	if (typeof source?.getRandomValues !== "function") return null;
	const bytes = new Uint8Array(new ArrayBuffer(16));
	source.getRandomValues(bytes);
	return Array.from(bytes, (b) => b.toString(16).padStart(2, "0")).join("");
}
