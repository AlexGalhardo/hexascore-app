import { logout } from "./logout";
import { createSession } from "./session";
import { createTokenStore, memoryVault } from "./tokenStore";

const saved = { token: "t", tokenId: "i", expiresAt: "2027-01-01T00:00:00Z" };

test("logout calls DELETE /tokens/current and wipes locally", async () => {
	const vault = memoryVault();
	const session = createSession({ vault, wipeUserData: async () => {} });
	await session.signIn(saved);
	const client = { tokensRevokeCurrent: jest.fn(async () => undefined) };
	await logout({ client, session });
	expect(client.tokensRevokeCurrent).toHaveBeenCalledTimes(1);
	expect(await createTokenStore(vault).load()).toBeNull();
});

test("logout with the network down still wipes the token", async () => {
	const vault = memoryVault();
	const wipeUserData = jest.fn(async () => {});
	const session = createSession({ vault, wipeUserData });
	await session.signIn(saved);
	const client = {
		tokensRevokeCurrent: jest.fn(async () => {
			throw new TypeError("Network request failed");
		}),
	};
	await logout({ client, session });
	expect(session.token()).toBeNull();
	expect(session.state()).toEqual({ kind: "signedOut" });
	expect(wipeUserData).toHaveBeenCalledTimes(1);
});
