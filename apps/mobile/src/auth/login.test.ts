import { ApiError } from "../../../../packages/app-client/src/api";
import { defaultDeviceName, isValidDeviceName } from "./deviceName";
import { challengeUrl, login } from "./login";

const input = { identifier: "ana", password: "pw", deviceName: "HexaScore · Pixel" };
const issued = { token: "hx_app_t", tokenId: "id1", expiresAt: "2027-01-01T00:00:00Z" };
const fail = (status: number, code: string) => new ApiError(status, { status, code } as never);

const clientWith = (impl: () => Promise<unknown>) => ({ tokensCreate: jest.fn(impl) });

test("returns the token on success", async () => {
	const client = clientWith(async () => issued);
	expect(await login(client, input)).toEqual({ kind: "ok", token: issued });
	expect(client.tokensCreate).toHaveBeenCalledWith({ body: input });
});

test("422 two-factor-required asks for the code", async () => {
	const client = clientWith(async () => {
		throw fail(422, "two-factor-required");
	});
	expect(await login(client, input)).toEqual({ kind: "twoFactor" });
});

test("a recovery code is sent as recoveryCode and accepted", async () => {
	const client = clientWith(async () => issued);
	const outcome = await login(client, { ...input, recoveryCode: "abcd-efgh" });
	expect(outcome.kind).toBe("ok");
	expect(client.tokensCreate).toHaveBeenCalledWith({
		body: { ...input, recoveryCode: "abcd-efgh" },
	});
});

test("428 challenge-required opens the web login", async () => {
	const client = clientWith(async () => {
		throw fail(428, "challenge-required");
	});
	expect(await login(client, input)).toEqual({ kind: "challenge" });
	expect(challengeUrl("https://hexascore.app/app/v1")).toBe("https://hexascore.app/login");
});

test("invalid credentials and network failures map to auth.errors keys", async () => {
	const bad = clientWith(async () => {
		throw fail(422, "invalid-credentials");
	});
	expect(await login(bad, input)).toMatchObject({
		kind: "error",
		key: "auth.errors.invalidCredentials",
	});
	const down = clientWith(async () => {
		throw new TypeError("Network request failed");
	});
	expect(await login(down, input)).toMatchObject({ kind: "error", key: "auth.errors.network" });
});

test("a malformed 201 body is an error, never a stored token", async () => {
	const client = clientWith(async () => ({ token: "x" }));
	expect(await login(client, input)).toMatchObject({ kind: "error" });
});

test("device name defaults to HexaScore · model and stays within 1-64", () => {
	expect(defaultDeviceName("Pixel 9")).toBe("HexaScore · Pixel 9");
	expect(defaultDeviceName(null)).toBe("HexaScore · Android");
	expect(defaultDeviceName("x".repeat(100))).toHaveLength(64);
	expect(isValidDeviceName("  ")).toBe(false);
	expect(isValidDeviceName("a".repeat(65))).toBe(false);
	expect(isValidDeviceName("a")).toBe(true);
});
