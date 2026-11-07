import { googleRedirectUri, promptGoogleIdToken } from "./googlePrompt";

const mockPrompt = jest.fn();
const mockExchange = jest.fn();
let lastConfig: Record<string, unknown> = {};

jest.mock("expo-auth-session", () => ({
	CodeChallengeMethod: { S256: "S256" },
	ResponseType: { Code: "code" },
	makeRedirectUri: ({ scheme, path }: { scheme: string; path: string }) => `${scheme}:/${path}`,
	exchangeCodeAsync: (...args: unknown[]) => mockExchange(...args),
	AuthRequest: class {
		codeVerifier = "verifier-xyz";
		constructor(config: Record<string, unknown>) {
			lastConfig = config;
		}
		promptAsync = (...args: unknown[]) => mockPrompt(...args);
	},
}));
jest.mock("expo-crypto", () => ({
	getRandomBytes: (n: number) => new Uint8Array(n).fill(0xab),
}));

const nonce = async () => "server-nonce";
const ok = (state: string) => ({ type: "success", params: { code: "auth-code", state } });

beforeEach(() => {
	mockPrompt.mockReset();
	mockExchange.mockReset();
});

test("the redirect is the reversed client id as a custom scheme", () => {
	expect(googleRedirectUri("123-abc.apps.googleusercontent.com")).toBe(
		"com.googleusercontent.apps.123-abc:/oauthredirect",
	);
});

test("code flow: PKCE S256, server nonce and random state go out, the code is exchanged with the verifier", async () => {
	mockPrompt.mockImplementation(async () => ok(String(lastConfig.state)));
	mockExchange.mockResolvedValue({ idToken: "the-id-token" });

	expect(await promptGoogleIdToken("123-abc.apps.googleusercontent.com", nonce)).toBe(
		"the-id-token",
	);

	expect(lastConfig).toMatchObject({
		responseType: "code",
		usePKCE: true,
		codeChallengeMethod: "S256",
		extraParams: { nonce: "server-nonce" },
		state: "abababababababababababababababab",
	});
	expect(mockExchange).toHaveBeenCalledWith(
		expect.objectContaining({ code: "auth-code", extraParams: { code_verifier: "verifier-xyz" } }),
		{ tokenEndpoint: "https://oauth2.googleapis.com/token" },
	);
});

test("fails closed without a server nonce", async () => {
	expect(await promptGoogleIdToken("c", async () => null)).toBeNull();
	expect(mockPrompt).not.toHaveBeenCalled();
});

test("a returned state that differs is refused before the exchange", async () => {
	mockPrompt.mockResolvedValue(ok("attacker-state"));
	expect(await promptGoogleIdToken("c", nonce)).toBeNull();
	expect(mockExchange).not.toHaveBeenCalled();
});

test("cancel, a missing code or an exchange without an ID token give null", async () => {
	mockPrompt.mockResolvedValue({ type: "cancel" });
	expect(await promptGoogleIdToken("c", nonce)).toBeNull();

	mockPrompt.mockImplementation(async () => ({
		type: "success",
		params: { state: lastConfig.state },
	}));
	expect(await promptGoogleIdToken("c", nonce)).toBeNull();

	mockPrompt.mockImplementation(async () => ok(String(lastConfig.state)));
	mockExchange.mockResolvedValue({});
	expect(await promptGoogleIdToken("c", nonce)).toBeNull();

	mockExchange.mockRejectedValue(new Error("network"));
	expect(await promptGoogleIdToken("c", nonce)).toBeNull();
});
