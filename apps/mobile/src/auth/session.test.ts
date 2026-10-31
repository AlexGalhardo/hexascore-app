import { ApiError } from "../../../../packages/app-client/src/api";
import { createSession } from "./session";
import { createTokenStore, memoryVault } from "./tokenStore";

const saved = { token: "hx_app_secret", tokenId: "t1", expiresAt: "2027-01-01T00:00:00Z" };

function setup() {
	const vault = memoryVault();
	const wipeUserData = jest.fn(async () => {});
	const session = createSession({ vault, wipeUserData });
	return { vault, wipeUserData, session, store: createTokenStore(vault) };
}

const problem = (status: number, extra: Record<string, unknown>): ApiError =>
	new ApiError(status, { status, ...extra });

test("signIn persists only token, tokenId and expiresAt", async () => {
	const { session, vault } = setup();
	await session.signIn({ ...saved, extra: "x" } as typeof saved);
	expect(JSON.parse((await vault.get("hexascore.token")) as string)).toEqual(saved);
	expect(session.token()).toBe("hx_app_secret");
	expect(session.state()).toEqual({ kind: "active", emailNotVerified: false });
});

test("restore reads the vault and tolerates corrupt data", async () => {
	const a = setup();
	await a.session.signIn(saved);
	const b = createSession({ vault: a.vault, wipeUserData: async () => {} });
	await b.restore();
	expect(b.token()).toBe("hx_app_secret");
	await a.vault.set("hexascore.token", "{not json");
	const c = createSession({ vault: a.vault, wipeUserData: async () => {} });
	await c.restore();
	expect(c.state()).toEqual({ kind: "signedOut" });
});

test("401 wipes token, cache and queue", async () => {
	const { session, store, wipeUserData } = setup();
	await session.signIn(saved);
	expect(await session.handleError(problem(401, { code: "unauthorized" }))).toBe(true);
	expect(await store.load()).toBeNull();
	expect(session.token()).toBeNull();
	expect(wipeUserData).toHaveBeenCalledTimes(1);
	expect(session.state()).toEqual({ kind: "signedOut" });
});

test("403 account-banned wipes and exposes reason and end date", async () => {
	const { session, store } = setup();
	await session.signIn(saved);
	await session.handleError(
		problem(403, { code: "account-banned", reason: "spam", expiresAt: "2026-12-01T00:00:00Z" }),
	);
	expect(await store.load()).toBeNull();
	expect(session.state()).toEqual({
		kind: "banned",
		reason: "spam",
		expiresAt: "2026-12-01T00:00:00Z",
	});
});

test("403 email-not-verified keeps the token and raises the banner", async () => {
	const { session, store, wipeUserData } = setup();
	await session.signIn(saved);
	await session.handleError(problem(403, { code: "email-not-verified" }));
	expect(session.state()).toEqual({ kind: "active", emailNotVerified: true });
	expect(await store.load()).toEqual(saved);
	expect(wipeUserData).not.toHaveBeenCalled();
});

test("other errors and non-API errors leave the session alone", async () => {
	const { session } = setup();
	await session.signIn(saved);
	expect(await session.handleError(problem(500, {}))).toBe(false);
	expect(await session.handleError(new Error("offline"))).toBe(false);
	expect(session.state().kind).toBe("active");
});

test("deletionScheduledAt routes to the cancel screen", async () => {
	const { session } = setup();
	await session.signIn(saved);
	session.observeMe({ deletionScheduledAt: null });
	expect(session.state().kind).toBe("active");
	session.observeMe({ deletionScheduledAt: "2026-11-30T00:00:00Z" });
	expect(session.state()).toEqual({
		kind: "deletionScheduled",
		scheduledAt: "2026-11-30T00:00:00Z",
	});
});

test("subscribers see changes until they unsubscribe", async () => {
	const { session } = setup();
	const seen: string[] = [];
	const off = session.subscribe((s) => seen.push(s.kind));
	await session.signIn(saved);
	off();
	await session.signOut();
	expect(seen).toEqual(["active"]);
});
