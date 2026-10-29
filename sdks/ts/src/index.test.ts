import { expect, test } from "bun:test";
import { createHexascoreClient, getItem, HEXASCORE_BASE_URL, listItems } from "./index.ts";

function recorder(body: unknown, status = 200): { calls: Request[]; fetch: typeof globalThis.fetch } {
	const calls: Request[] = [];
	const fetch = (async (input: RequestInfo | URL, init?: RequestInit) => {
		calls.push(input instanceof Request ? input : new Request(input, init));
		return new Response(JSON.stringify(body), { status, headers: { "content-type": "application/json" } });
	}) as typeof globalThis.fetch;
	return { calls, fetch };
}

test("sends the Bearer key to the production base URL with the query serialized", async () => {
	const { calls, fetch } = recorder({ data: [], next_cursor: null });
	const client = createHexascoreClient({ apiKey: "hx_live_test", fetch });

	const { data, error } = await listItems({ client, path: { type: "games" }, query: { sort: "-score", limit: "5", min_score: "8" } });

	expect(error).toBeUndefined();
	expect(data).toEqual({ data: [], next_cursor: null });
	const url = new URL(calls[0]?.url ?? "");
	expect(`${url.origin}${url.pathname}`).toBe(`${HEXASCORE_BASE_URL}/api/v1/games`);
	expect(Object.fromEntries(url.searchParams)).toEqual({ sort: "-score", limit: "5", min_score: "8" });
	expect(calls[0]?.headers.get("authorization")).toBe("Bearer hx_live_test");
});

test("returns the API problem instead of throwing, and honours a custom base URL", async () => {
	const problem = { type: "about:blank", title: "Not Found", status: 404, code: "not_found", detail: "No such game." };
	const { calls, fetch } = recorder(problem, 404);
	const client = createHexascoreClient({ apiKey: "k", baseUrl: "https://example.test", fetch });

	const { data, error, response } = await getItem({ client, path: { type: "games", id: "019a0000-0000-7000-8000-000000000000" } });

	expect(data).toBeUndefined();
	expect(error).toMatchObject({ code: "not_found", status: 404 });
	expect(response?.status).toBe(404);
	expect(calls[0]?.url).toBe("https://example.test/api/v1/games/019a0000-0000-7000-8000-000000000000");
});

test("refuses to build a client without a key", () => {
	expect(() => createHexascoreClient({ apiKey: "" })).toThrow(TypeError);
});
