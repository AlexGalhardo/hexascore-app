import type { Client } from "../../../../packages/app-client/src/api";
import type { Session } from "./session";

export async function logout(deps: {
	client: Pick<Client, "tokensRevokeCurrent">;
	session: Pick<Session, "signOut">;
}): Promise<void> {
	try {
		await deps.client.tokensRevokeCurrent();
	} catch {
		// Offline or already revoked: the local wipe below is what protects the user (SDD 10 §6.2).
	}
	await deps.session.signOut();
}
