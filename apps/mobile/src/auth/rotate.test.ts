import { ApiError } from "../../../../packages/app-client/src/api";
import { rotateIfDue } from "./rotate";
import { createSession } from "./session";
import { createTokenStore, memoryVault } from "./tokenStore";

const DAY = 86_400_000;
const NOW = Date.parse("2026-11-01T00:00:00Z");
const fresh = { token: "new", tokenId: "n1", expiresAt: new Date(NOW + 60 * DAY).toISOString() };

async function setup(daysLeft: number, rotate: () => Promise<unknown> = async () => fresh) {
	const vault = memoryVault();
	const session = createSession({ vault, wipeUserData: async () => {} });
	await session.signIn({
		token: "old",
		tokenId: "o1",
		expiresAt: new Date(NOW + daysLeft * DAY).toISOString(),
	});
	const client = { tokensRotate: jest.fn(rotate) };
	const run = (now = NOW) => rotateIfDue({ vault, client, session, now: () => now });
	return { vault, session, client, run };
}

test("no rotation at 16 days left", async () => {
	const { run, client } = await setup(16);
	expect(await run()).toBe("skipped");
	expect(client.tokensRotate).not.toHaveBeenCalled();
});

test("one rotation at 14 days left swaps the stored token; a second foreground within 24 h does nothing", async () => {
	const { run, client, session, vault } = await setup(14);
	expect(await run()).toBe("rotated");
	expect(session.token()).toBe("new");
	expect((await createTokenStore(vault).load())?.tokenId).toBe("n1");
	// still "due" by the clock the second time: force it by rewinding the stored expiry
	await session.signIn({
		token: "new",
		tokenId: "n1",
		expiresAt: new Date(NOW + 10 * DAY).toISOString(),
	});
	expect(await run(NOW + DAY - 1)).toBe("skipped");
	expect(client.tokensRotate).toHaveBeenCalledTimes(1);
	expect(await run(NOW + DAY)).toBe("rotated");
});

test("a failed rotation keeps the old token and is retried at the next foreground", async () => {
	let fail = true;
	const { run, session, client } = await setup(5, async () => {
		if (fail) throw new TypeError("offline");
		return fresh;
	});
	expect(await run()).toBe("failed");
	expect(session.token()).toBe("old");
	fail = false;
	expect(await run()).toBe("rotated");
	expect(client.tokensRotate).toHaveBeenCalledTimes(2);
});

test("a 401 on rotate signs the session out", async () => {
	const { run, session } = await setup(5, async () => {
		throw new ApiError(401, { status: 401, code: "unauthorized" } as never);
	});
	expect(await run()).toBe("failed");
	expect(session.state()).toEqual({ kind: "signedOut" });
});

test("a malformed rotate body never replaces the token", async () => {
	const { run, session } = await setup(5, async () => ({ nope: true }));
	expect(await run()).toBe("failed");
	expect(session.token()).toBe("old");
});
