import { expect, test } from "../helpers/fixtures";

// HSTS is only sent to HTTPS clients and the container is reached over plain HTTP, so it is not asserted here.
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

		// The Inertia page object is a non-executable JSON block and needs no nonce.
		const scripts = ((await first.text()).match(/<script\b[^>]*>/g) ?? []).filter(
			(tag) => !tag.includes('type="application/json"'),
		);
		expect(scripts.length).toBeGreaterThan(0);
		for (const tag of scripts) expect(tag).toContain(`nonce="${nonce}"`);

		const h = first.headers();
		expect(h["x-content-type-options"]).toBe("nosniff");
		expect(h["x-frame-options"]).toBe("DENY");
		expect(h["referrer-policy"]).toBe("strict-origin-when-cross-origin");
		expect(h["permissions-policy"]).toContain("camera=()");
		expect(h["cross-origin-opener-policy"]).toBe("same-origin");
		expect(h["x-powered-by"]).toBeUndefined();
	});

	test("API errors are RFC 9457 problems and cannot be framed or sniffed", async ({
		request,
	}) => {
		const response = await request.get("/api/v1/games");
		expect(response.status()).toBe(401);
		const h = response.headers();
		expect(h["content-type"]).toContain("application/problem+json");
		expect(h["x-frame-options"]).toBe("DENY");
		expect(h["x-content-type-options"]).toBe("nosniff");
	});
});
