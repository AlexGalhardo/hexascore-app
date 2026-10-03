import { type Client, createClient, createConfig } from "./gen/client";

export * from "./gen";
export type { Client } from "./gen/client";

export type HexascoreOptions = {
	/** Secret key from https://hexascore.app/profile/api (`hx_live_…`). Keep it on the server. */
	apiKey: string;
	/** Defaults to https://api.hexascore.app/v1. */
	baseUrl?: string;
};

/** The only hand-written piece: a client with the base URL and Bearer auth set. Pass it as `{ client }`. */
export function createHexascoreClient({ apiKey, baseUrl = "https://api.hexascore.app/v1" }: HexascoreOptions): Client {
	return createClient(createConfig({ baseUrl, headers: { Authorization: `Bearer ${apiKey}` } }));
}
