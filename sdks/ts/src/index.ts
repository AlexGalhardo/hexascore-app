import { type Client, createClient, createConfig } from "./gen/client/index.ts";

export type { Client } from "./gen/client/index.ts";
export * from "./gen/index.ts";

/** Production base URL of the Hexascore API v1. */
export const HEXASCORE_BASE_URL = "https://api.hexascore.app";

export type HexascoreOptions = {
	/** Secret key from https://hexascore.app/profile/api (`hx_live_…`). Keep it on the server. */
	apiKey: string;
	/** Defaults to {@link HEXASCORE_BASE_URL}. Another deployment: `https://<host>`. */
	baseUrl?: string;
	/** Custom `fetch` (tests, proxies, runtimes without a global one). */
	fetch?: typeof globalThis.fetch;
};

/**
 * The only hand-written piece: a client with the base URL and Bearer auth set. Pass it to every operation as
 * `{ client }`.
 *
 * @example
 * ```ts
 * const client = createHexascoreClient({ apiKey: process.env.HEXASCORE_API_KEY! });
 * const { data, error } = await listItems({ client, path: { type: "games" }, query: { sort: "-score", limit: 10 } });
 * ```
 */
export function createHexascoreClient({ apiKey, baseUrl = HEXASCORE_BASE_URL, fetch }: HexascoreOptions): Client {
	if (!apiKey) throw new TypeError("createHexascoreClient: apiKey is required");
	return createClient(createConfig({ baseUrl, headers: { Authorization: `Bearer ${apiKey}` }, ...(fetch && { fetch }) }));
}
