import { expect, test } from "../helpers/fixtures";

const UNKNOWN_UUID = "00000000-0000-4000-8000-000000000000";

test.describe("api contract", () => {
	test("the OpenAPI document is served as JSON with paths", async ({ request }) => {
		const response = await request.get("/api/v1/openapi.json");
		expect(response.status()).toBe(200);
		expect(response.headers()["content-type"]).toContain("application/json");
		const doc = (await response.json()) as { openapi?: unknown; paths?: unknown };
		expect(String(doc.openapi)).toMatch(/^3\./);
		expect(Object.keys(doc.paths ?? {}).length).toBeGreaterThan(0);
	});

	test("protected app API routes require a bearer token (401 problem)", async ({ request }) => {
		for (const path of ["/app/v1/me", "/app/v1/me/for-you", "/app/v1/tokens"]) {
			const response = await request.get(path);
			expect(response.status(), path).toBe(401);
			expect(response.headers()["content-type"], path).toContain("application/problem+json");
		}
	});

	test("an invalid bearer token is rejected with 401, not served", async ({ request }) => {
		const response = await request.get("/app/v1/me", {
			headers: { authorization: "Bearer hx_live_invalid" },
		});
		expect(response.status()).toBe(401);
	});

	test("public item ratings: unknown id is 404, malformed id is 404, bad page is 400", async ({
		request,
	}) => {
		const unknown = await request.get(`/api/items/${UNKNOWN_UUID}/ratings`);
		expect(unknown.status()).toBe(404);

		const malformed = await request.get("/api/items/not-a-uuid/ratings");
		expect(malformed.status()).toBe(404);

		for (const page of ["0", "-1", "abc", "101"]) {
			const response = await request.get(`/api/items/${UNKNOWN_UUID}/ratings?page=${page}`);
			expect(response.status(), `page=${page}`).toBe(400);
		}
	});

	test("/go/{id} rejects unknown offers without redirecting to a third party", async ({
		request,
	}) => {
		const response = await request.get(`/go/${UNKNOWN_UUID}`, { maxRedirects: 0 });
		expect(response.status()).toBe(404);
		expect(response.headers().location ?? "").toBe("");
	});
});
