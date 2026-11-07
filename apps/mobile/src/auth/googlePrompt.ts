import {
	AuthRequest,
	CodeChallengeMethod,
	exchangeCodeAsync,
	makeRedirectUri,
	ResponseType,
} from "expo-auth-session";
import { getRandomBytes } from "expo-crypto";

const GOOGLE_AUTHORIZATION_ENDPOINT = "https://accounts.google.com/o/oauth2/v2/auth";
const GOOGLE_TOKEN_ENDPOINT = "https://oauth2.googleapis.com/token";

/** Set EXPO_PUBLIC_GOOGLE_CLIENT_ID (the installed-app OAuth client) to enable the button. */
export const googleClientId = process.env.EXPO_PUBLIC_GOOGLE_CLIENT_ID ?? "";

const hex = (bytes: Uint8Array): string =>
	Array.from(bytes, (b) => b.toString(16).padStart(2, "0")).join("");
const _base64Url = (b64: string): string =>
	b64.replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "");

/** Google's installed-app redirect: the reversed client id as a custom scheme (registered in app.config.ts). */
export function googleRedirectUri(clientId: string): string {
	const prefix = clientId.replace(/\.apps\.googleusercontent\.com$/, "");
	return makeRedirectUri({ scheme: `com.googleusercontent.apps.${prefix}`, path: "oauthredirect" });
}

/**
 * Authorization code flow with PKCE (S256) in the system browser, bound to a server-issued nonce (SDD 03 §6.2).
 * state and nonce come from the platform CSPRNG (expo-crypto) and the PKCE verifier from AuthRequest (also expo-crypto); there is no weaker fallback. Resolves to
 * the ID token, or null when the user cancels, the state does not match or any step fails: the caller signs in
 * only with a token.
 */
export async function promptGoogleIdToken(
	clientId: string,
	fetchNonce: () => Promise<string | null>,
): Promise<string | null> {
	try {
		const nonce = await fetchNonce();
		if (!nonce) return null;

		const state = hex(getRandomBytes(16));
		const redirectUri = googleRedirectUri(clientId);

		const request = new AuthRequest({
			clientId,
			redirectUri,
			responseType: ResponseType.Code,
			scopes: ["openid", "profile", "email"],
			// PKCE S256: AuthRequest draws the verifier from expo-crypto and sends the S256 challenge.
			usePKCE: true,
			codeChallengeMethod: CodeChallengeMethod.S256,
			state,
			extraParams: { nonce },
		});
		const result = await request.promptAsync({
			authorizationEndpoint: GOOGLE_AUTHORIZATION_ENDPOINT,
		});
		const verifier = request.codeVerifier;
		if (
			result.type !== "success" ||
			result.params.state !== state ||
			!result.params.code ||
			!verifier
		)
			return null;

		const tokens = await exchangeCodeAsync(
			{ clientId, code: result.params.code, redirectUri, extraParams: { code_verifier: verifier } },
			{ tokenEndpoint: GOOGLE_TOKEN_ENDPOINT },
		);
		return tokens.idToken ?? null;
	} catch {
		return null;
	}
}
