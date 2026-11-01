import { ApiError, type Client, createClient } from "../../../../packages/app-client/src/api";
import type { AppLocale } from "../i18n/locale";
import { isBelowMinimum } from "./version";

const READ_METHODS = [
	"getEntity",
	"getItem",
	"library",
	"listBookGenres",
	"listBookTrending",
	"listEditions",
	"listEntityItems",
	"listGameGenres",
	"listGameTrending",
	"listItems",
	"listMovieGenres",
	"listMovieTrending",
	"listSeriesGenres",
	"listSeriesTrending",
	"listsIndex",
	"listsShow",
	"me",
	"meShow",
	"notificationsIndex",
	"searchIndex",
	"searchSuggest",
	"tokensIndex",
	"usersShow",
] as const;

export type ReadClient = Pick<Client, (typeof READ_METHODS)[number]>;

declare const idempotencyBrand: unique symbol;
/** A UUID v7 chosen when the write is created (SDD 10 §7.3); only `newIdempotencyKey` mints one. */
export type IdempotencyKey = string & { readonly [idempotencyBrand]: true };

type RandomBytes = () => Uint8Array;

const defaultRandom: RandomBytes = () => {
	const bytes = new Uint8Array(16);
	const webCrypto = (globalThis as { crypto?: { getRandomValues?: (a: Uint8Array) => Uint8Array } })
		.crypto;
	// shortcut: Math.random when Hermes exposes no crypto; keys only need uniqueness per user, not secrecy. Replace with expo-crypto if it is ever approved.
	if (webCrypto?.getRandomValues) webCrypto.getRandomValues(bytes);
	else for (let i = 0; i < 16; i++) bytes[i] = Math.floor(Math.random() * 256);
	return bytes;
};

export function uuidv7(now: number = Date.now(), random: RandomBytes = defaultRandom): string {
	const bytes = random().slice();
	for (let i = 0; i < 6; i++) bytes[i] = Math.floor(now / 2 ** (8 * (5 - i))) & 0xff;
	bytes[6] = ((bytes[6] ?? 0) & 0x0f) | 0x70;
	bytes[8] = ((bytes[8] ?? 0) & 0x3f) | 0x80;
	const hex = Array.from(bytes, (b) => b.toString(16).padStart(2, "0")).join("");
	return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20, 32)}`;
}

export type AppClientOptions = {
	/** From EXPO_PUBLIC_API_BASE_URL (SDD 10 §4). */
	baseUrl: string;
	/** Running app version (SemVer), sent as X-App-Version. */
	version: string;
	/** Read per request, so a locale change applies without a restart. */
	locale: () => AppLocale;
	token?: () => string | null;
	fetch?: typeof fetch;
	onUpgradeRequired?: () => void;
};

export type AppClient = {
	read: ReadClient;
	/** Writes cannot be issued without a key: the key is a required, branded argument. */
	write<T>(key: IdempotencyKey, call: (client: Client) => Promise<T>): Promise<T>;
	newIdempotencyKey(): IdempotencyKey;
	upgradeRequired(): boolean;
};

export function createAppClient(options: AppClientOptions): AppClient {
	let upgrade = false;
	const markUpgrade = (): void => {
		if (upgrade) return;
		upgrade = true;
		options.onUpgradeRequired?.();
	};
	const baseFetch: typeof fetch = options.fetch ?? ((input, init) => globalThis.fetch(input, init));

	// 426 is the guard's answer; the header also arrives on 2xx so the screen shows before the next 426 (FR-10-5).
	const watching: typeof fetch = async (input, init) => {
		const response = await baseFetch(input, init);
		const minimum = response.headers.get("X-App-Min-Version");
		if (response.status === 426 || isBelowMinimum(options.version, minimum)) markUpgrade();
		return response;
	};

	const headers = (): Record<string, string> => {
		const token = options.token?.();
		return {
			"X-App-Version": options.version,
			"Accept-Language": options.locale(),
			...(token ? { Authorization: `Bearer ${token}` } : {}),
		};
	};

	const reader = createClient({ baseUrl: options.baseUrl, fetch: watching, headers });
	return {
		read: reader,
		async write(key, call) {
			let retryAfter: number | null = null;
			// ApiError keeps no headers; the offline queue needs Retry-After on 429 (SDD 10 §7.3).
			const recording: typeof fetch = async (input, init) => {
				const response = await watching(input, init);
				const seconds = Number(response.headers.get("Retry-After") ?? Number.NaN);
				retryAfter = Number.isFinite(seconds) ? seconds : null;
				return response;
			};
			try {
				return await call(
					createClient({
						baseUrl: options.baseUrl,
						fetch: recording,
						headers: () => ({ ...headers(), "Idempotency-Key": key }),
					}),
				);
			} catch (error) {
				if (error instanceof ApiError) throw Object.assign(error, { retryAfter });
				throw error;
			}
		},
		newIdempotencyKey: () => uuidv7() as IdempotencyKey,
		upgradeRequired: () => upgrade,
	};
}
