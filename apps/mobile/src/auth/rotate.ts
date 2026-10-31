import { AppState } from "react-native";
import type { Client } from "../../../../packages/app-client/src/api";
import type { Session } from "./session";
import { createTokenStore, isStoredToken, type KeyValueVault } from "./tokenStore";

const DAY = 86_400_000;
const ROTATE_WITHIN = 15 * DAY;
const MIN_INTERVAL = DAY;
const LAST_ROTATE_KEY = "hexascore.lastRotate";

export type RotateDeps = {
	vault: KeyValueVault;
	client: Pick<Client, "tokensRotate">;
	session: Pick<Session, "signIn" | "handleError">;
	now?: () => number;
};

// Called on every foreground. A lost response ends in a 401 and a new login (SDD 10 §8); the old token is never retried after that.
export async function rotateIfDue(deps: RotateDeps): Promise<"skipped" | "rotated" | "failed"> {
	const now = (deps.now ?? Date.now)();
	const current = await createTokenStore(deps.vault).load();
	if (!current) return "skipped";
	const remaining = Date.parse(current.expiresAt) - now;
	if (remaining >= ROTATE_WITHIN) return "skipped";
	const last = Number(await deps.vault.get(LAST_ROTATE_KEY));
	if (Number.isFinite(last) && now - last < MIN_INTERVAL) return "skipped";

	try {
		const body: unknown = await deps.client.tokensRotate();
		if (!isStoredToken(body)) return "failed";
		await deps.session.signIn({
			token: body.token,
			tokenId: body.tokenId,
			expiresAt: body.expiresAt,
		});
		// Stamped on success only, so a failed attempt is retried at the next foreground.
		await deps.vault.set(LAST_ROTATE_KEY, String(now));
		return "rotated";
	} catch (error) {
		await deps.session.handleError(error);
		return "failed";
	}
}

/** Runs `rotateIfDue` whenever the app returns to the foreground; returns the unsubscribe. */
export function startForegroundRotation(deps: RotateDeps): () => void {
	const subscription = AppState.addEventListener("change", (state) => {
		if (state === "active") void rotateIfDue(deps);
	});
	return () => subscription.remove();
}
