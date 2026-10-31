import { ApiError } from "../../../../packages/app-client/src/api";
import { googleSignIn, signInWithGoogle } from "./google";

const issued = { token: "hx_app_t", tokenId: "id1", expiresAt: "2027-01-01T00:00:00Z" };
const fail = (status: number, code: string) => new ApiError(status, { status, code } as never);

test("cancel leaves silently without calling the API", async () => {
	const client = { tokensGoogle: jest.fn() };
	expect(await googleSignIn(client, async () => null, "Pixel")).toEqual({ kind: "cancelled" });
	expect(client.tokensGoogle).not.toHaveBeenCalled();
});

test("posts only the ID token and the device name", async () => {
	const client = { tokensGoogle: jest.fn(async () => issued) };
	expect(await googleSignIn(client, async () => "idtok", "Pixel")).toEqual({
		kind: "ok",
		token: issued,
	});
	expect(client.tokensGoogle).toHaveBeenCalledWith({
		body: { idToken: "idtok", deviceName: "Pixel" },
	});
});

test("409 email-unverified-at-provider shows its message", async () => {
	const client = {
		tokensGoogle: jest.fn(async () => {
			throw fail(409, "email-unverified-at-provider");
		}),
	};
	expect(await googleSignIn(client, async () => "idtok", "Pixel")).toMatchObject({
		kind: "error",
		key: "auth.errors.emailUnverifiedAtProvider",
	});
});

test("the 2FA branch retries with the same ID token and a code", async () => {
	const tokensGoogle = jest
		.fn()
		.mockRejectedValueOnce(fail(422, "two-factor-required"))
		.mockResolvedValueOnce(issued);
	const client = { tokensGoogle };
	expect(await googleSignIn(client, async () => "idtok", "Pixel")).toEqual({ kind: "twoFactor" });
	expect(
		await signInWithGoogle(client, { idToken: "idtok", deviceName: "Pixel", code: "123456" }),
	).toEqual({
		kind: "ok",
		token: issued,
	});
	expect(tokensGoogle).toHaveBeenLastCalledWith({
		body: { idToken: "idtok", deviceName: "Pixel", code: "123456" },
	});
});
