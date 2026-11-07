import { AuthRequest, makeRedirectUri, ResponseType } from "expo-auth-session";

const GOOGLE_AUTHORIZATION_ENDPOINT = "https://accounts.google.com/o/oauth2/v2/auth";

/** Set EXPO_PUBLIC_GOOGLE_CLIENT_ID (an OAuth client allowed to return ID tokens) to enable the button. */
export const googleClientId = process.env.EXPO_PUBLIC_GOOGLE_CLIENT_ID ?? "";

// Only the ID token is requested (SDD 03 §6.2). Resolves to null when the user cancels or it fails.
// shortcut: implicit ID-token flow through the system browser; unverified against a real Google client
// (no client id in this environment), switch to the Credential Manager if Google refuses it for Android clients.
export async function promptGoogleIdToken(clientId: string): Promise<string | null> {
	const nonce = `${Date.now().toString(36)}${Math.random().toString(36).slice(2)}`;
	const request = new AuthRequest({
		clientId,
		redirectUri: makeRedirectUri({ scheme: "hexascore", path: "auth/google" }),
		responseType: ResponseType.IdToken,
		scopes: ["openid", "profile", "email"],
		usePKCE: false,
		extraParams: { nonce },
	});
	try {
		const result = await request.promptAsync({
			authorizationEndpoint: GOOGLE_AUTHORIZATION_ENDPOINT,
		});
		return result.type === "success" ? (result.params.id_token ?? null) : null;
	} catch {
		return null;
	}
}
