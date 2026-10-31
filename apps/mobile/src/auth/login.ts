import type { Client } from "../../../../packages/app-client/src/api";
import { ApiError } from "../../../../packages/app-client/src/api";
import { describeError } from "../api/problems";
import { isStoredToken, type StoredToken } from "./tokenStore";

export type LoginInput = {
	identifier: string;
	password: string;
	deviceName: string;
	code?: string;
	recoveryCode?: string;
	challenge?: string;
};

export type LoginOutcome =
	| { kind: "ok"; token: StoredToken }
	| { kind: "twoFactor" }
	| { kind: "challenge" }
	| { kind: "error"; key: string; requestId: string | undefined };

const code = (error: ApiError): unknown => (error.problem as { code?: unknown } | undefined)?.code;

// The web login hosts the proof-of-work challenge (ADR 0016); the API base is `<origin>/app/v1`.
export function challengeUrl(baseUrl: string): string {
	return new URL("/login", baseUrl).toString();
}

/** Shared by the password and Google token routes: both answer with the same 201 and the same errors. */
export async function issueToken(call: () => Promise<unknown>): Promise<LoginOutcome> {
	try {
		const body: unknown = await call();
		if (!isStoredToken(body))
			return { kind: "error", key: "errors.unexpected", requestId: undefined };
		// abilities and any other member are dropped on purpose: only these three fields persist.
		return {
			kind: "ok",
			token: { token: body.token, tokenId: body.tokenId, expiresAt: body.expiresAt },
		};
	} catch (error) {
		if (!(error instanceof ApiError)) {
			return { kind: "error", key: "auth.errors.network", requestId: undefined };
		}
		const requestId = error.problem?.request_id;
		if (error.status === 422 && code(error) === "two-factor-required") return { kind: "twoFactor" };
		if (error.status === 428) return { kind: "challenge" };
		if (code(error) === "invalid-credentials") {
			return { kind: "error", key: "auth.errors.invalidCredentials", requestId };
		}
		if (error.status === 409 && code(error) === "email-unverified-at-provider") {
			return { kind: "error", key: "auth.errors.emailUnverifiedAtProvider", requestId };
		}
		return { kind: "error", ...describeError(error) };
	}
}

export function login(
	client: Pick<Client, "tokensCreate">,
	input: LoginInput,
): Promise<LoginOutcome> {
	return issueToken(() => client.tokensCreate({ body: input }));
}
