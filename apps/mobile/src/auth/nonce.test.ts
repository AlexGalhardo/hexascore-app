import { secureNonce } from "./nonce";

test("nonce is 32 hex characters from the injected CSPRNG", () => {
	const source = {
		getRandomValues: (bytes: Uint8Array<ArrayBuffer>) => {
			bytes.fill(0xab);
			return bytes;
		},
	};
	expect(secureNonce(source)).toBe("ab".repeat(16));
});

test("no CSPRNG gives no nonce instead of falling back to Math.random", () => {
	const spy = jest.spyOn(Math, "random");
	expect(secureNonce({})).toBeNull();
	expect(spy).not.toHaveBeenCalled();
});
