/** Minimal secure key-value port; the SecureStore adapter implements it (Q-B1). */
export type KeyValueVault = {
	get(key: string): Promise<string | null>;
	set(key: string, value: string): Promise<void>;
	delete(key: string): Promise<void>;
};

export type StoredToken = { token: string; tokenId: string; expiresAt: string };

const KEY = "hexascore.token";

export function memoryVault(): KeyValueVault {
	const data = new Map<string, string>();
	return {
		get: async (key) => data.get(key) ?? null,
		set: async (key, value) => {
			data.set(key, value);
		},
		delete: async (key) => {
			data.delete(key);
		},
	};
}

export function isStoredToken(value: unknown): value is StoredToken {
	if (typeof value !== "object" || value === null) return false;
	const v = value as Record<string, unknown>;
	return (
		typeof v.token === "string" && typeof v.tokenId === "string" && typeof v.expiresAt === "string"
	);
}

export type TokenStore = {
	load(): Promise<StoredToken | null>;
	save(token: StoredToken): Promise<void>;
	clear(): Promise<void>;
};

// Only these three fields are persisted (SDD 10 §5, about 100 B, under the SecureStore warning size).
export function createTokenStore(vault: KeyValueVault): TokenStore {
	return {
		async load() {
			const raw = await vault.get(KEY);
			if (raw === null) return null;
			try {
				const parsed: unknown = JSON.parse(raw);
				return isStoredToken(parsed) ? parsed : null;
			} catch {
				return null;
			}
		},
		async save({ token, tokenId, expiresAt }) {
			await vault.set(KEY, JSON.stringify({ token, tokenId, expiresAt }));
		},
		clear: () => vault.delete(KEY),
	};
}
