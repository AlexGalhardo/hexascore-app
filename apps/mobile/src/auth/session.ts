import { ApiError } from "../../../../packages/app-client/src/api";
import { createTokenStore, type KeyValueVault, type StoredToken } from "./tokenStore";

export type SessionState =
	| { kind: "loading" }
	| { kind: "signedOut" }
	| { kind: "active"; emailNotVerified: boolean }
	| { kind: "banned"; reason: string | null; expiresAt: string | null }
	| { kind: "deletionScheduled"; scheduledAt: string };

export type SessionDeps = {
	vault: KeyValueVault;
	/** Wipes the query cache and the offline queue for the user (SDD 10 §6.2). */
	wipeUserData: () => Promise<void>;
};

export type Session = {
	state(): SessionState;
	subscribe(listener: (state: SessionState) => void): () => void;
	token(): string | null;
	/** Reads the vault on launch. */
	restore(): Promise<void>;
	signIn(token: StoredToken): Promise<void>;
	/** Feed every failed API call here; returns true when the session changed. */
	handleError(error: unknown): Promise<boolean>;
	/** Feed the `GET /me` result; routes to the cancel screen when deletion is scheduled. */
	observeMe(me: { deletionScheduledAt?: string | null }): void;
	signOut(): Promise<void>;
};

// The generated ApiProblem omits RFC 9457 extension members (VF-NEW-P5.A.3.1); read them as unknown.
const extension = (error: ApiError, name: string): string | null => {
	const value: unknown = (error.problem as Record<string, unknown> | undefined)?.[name];
	return typeof value === "string" ? value : null;
};

export function createSession(deps: SessionDeps): Session {
	const store = createTokenStore(deps.vault);
	const listeners = new Set<(state: SessionState) => void>();
	let current: SessionState = { kind: "loading" };
	let token: string | null = null;

	const set = (next: SessionState): void => {
		current = next;
		for (const listener of listeners) listener(current);
	};

	const wipe = async (next: SessionState): Promise<void> => {
		token = null;
		await store.clear();
		await deps.wipeUserData();
		set(next);
	};

	return {
		state: () => current,
		subscribe(listener) {
			listeners.add(listener);
			return () => {
				listeners.delete(listener);
			};
		},
		token: () => token,
		async restore() {
			const saved = await store.load();
			token = saved?.token ?? null;
			set(saved ? { kind: "active", emailNotVerified: false } : { kind: "signedOut" });
		},
		async signIn(saved) {
			await store.save(saved);
			token = saved.token;
			set({ kind: "active", emailNotVerified: false });
		},
		async handleError(error) {
			if (!(error instanceof ApiError)) return false;
			const code = extension(error, "code");
			if (error.status === 401) {
				await wipe({ kind: "signedOut" });
				return true;
			}
			if (error.status === 403 && code === "account-banned") {
				await wipe({
					kind: "banned",
					reason: extension(error, "reason"),
					expiresAt: extension(error, "expiresAt"),
				});
				return true;
			}
			if (error.status === 403 && code === "email-not-verified" && current.kind === "active") {
				set({ kind: "active", emailNotVerified: true });
				return true;
			}
			return false;
		},
		observeMe(me) {
			if (current.kind === "active" && me.deletionScheduledAt) {
				set({ kind: "deletionScheduled", scheduledAt: me.deletionScheduledAt });
			}
		},
		signOut: () => wipe({ kind: "signedOut" }),
	};
}
