import { expect, test } from "../helpers/fixtures";

test.describe("security headers", () => {
	test("pages get a per-request nonce CSP that every script carries", async ({ request }) => {
		const first = await request.get("/");
		const second = await request.get("/");
		expect(first.status()).toBe(200);
		const csp = first.headers()["content-security-policy"] ?? "";
		const nonce = /'nonce-([^']+)'/.exec(csp)?.[1];
		expect(nonce, csp).toBeTruthy();
		expect(csp).toContain("'strict-dynamic'");
		expect(csp).toContain("frame-ancestors 'none'");
		expect(csp).not.toContain("'unsafe-eval'");
		expect(second.headers()["content-security-policy"]).not.toContain(`'nonce-${nonce}'`);

		const html = await first.text();
		const scripts = html.match(/<script\b[^>]*>/g) ?? [];
		expect(scripts.length).toBeGreaterThan(0);
		for (const tag of scripts) expect(tag).toContain(`nonce="${nonce}"`);

		const h = first.headers();
		expect(h["strict-transport-security"]).toContain("max-age=63072000");
		expect(h["x-content-type-options"]).toBe("nosniff");
		expect(h["referrer-policy"]).toBe("strict-origin-when-cross-origin");
		expect(h["permissions-policy"]).toContain("camera=()");
		expect(h["cross-origin-opener-policy"]).toBe("same-origin");
		expect(h["x-powered-by"]).toBeUndefined();
	});

	test("API responses are JSON-only and cannot be framed", async ({ request }) => {
		const response = await request.get("/api/v1/games");
		expect(response.status()).toBe(401);
		const h = response.headers();
		expect(h["content-type"]).toContain("application/problem+json");
		expect(h["content-security-policy"]).toBe(
			"default-src 'none'; frame-ancestors 'none'; base-uri 'none'; form-action 'none'",
		);
		expect(h["x-frame-options"]).toBe("DENY");
		expect(h["x-content-type-options"]).toBe("nosniff");
		expect(h["strict-transport-security"]).toBeTruthy();
	});

	test("Scalar allowances are scoped to the reference pages", async ({ request }) => {
		const reference = (await request.get("/docs/reference")).headers()["content-security-policy"] ?? "";
		const docs = (await request.get("/docs")).headers()["content-security-policy"] ?? "";
		expect(reference).toContain("https://api.hexascore.app");
		expect(docs).not.toContain("https://api.hexascore.app");
	});
});
