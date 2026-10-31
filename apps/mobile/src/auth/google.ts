import type { Client } from "../../../../packages/app-client/src/api";
import { issueToken, type LoginOutcome } from "./login";

export type GoogleOutcome = LoginOutcome | { kind: "cancelled" };

export type GoogleSignInInput = {
	idToken: string;
	deviceName: string;
	code?: string;
	recoveryCode?: string;
};

// Only the ID token goes to the server; no access token is requested or stored (SDD 03 §6.2).
export function signInWithGoogle(
	client: Pick<Client, "tokensGoogle">,
	input: GoogleSignInInput,
): Promise<LoginOutcome> {
	return issueToken(() => client.tokensGoogle({ body: input }));
}

/** `prompt` opens the Google account picker and resolves to null when the user cancels. */
export async function googleSignIn(
	client: Pick<Client, "tokensGoogle">,
	prompt: () => Promise<string | null>,
	deviceName: string,
): Promise<GoogleOutcome> {
	const idToken = await prompt();
	if (idToken === null) return { kind: "cancelled" };
	return signInWithGoogle(client, { idToken, deviceName });
}
