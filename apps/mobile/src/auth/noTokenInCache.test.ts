import { createSession } from "./session";
import { memoryVault } from "./tokenStore";

// The persisted query cache (P5.A.4.1) is a JSON dump of query state. Everything
// the session exposes to the UI must stay free of the token (FR-10-2).
test("serialised session state never carries the token", async () => {
	const session = createSession({ vault: memoryVault(), wipeUserData: async () => {} });
	await session.signIn({ token: "hx_app_abc", tokenId: "t1", expiresAt: "2027-01-01T00:00:00Z" });
	expect(JSON.stringify(session.state())).not.toContain("hx_app_");
});
