import { describe, expect, test } from "bun:test";
import nextConfig from "../next.config";
import { API_CSP, createNonce, isScalarPage, pageCsp } from "./security-headers";

function directives(csp: string): Map<string, string[]> {
	return new Map(
		csp.split(";").map((part) => {
			const [name, ...values] = part.trim().split(/\s+/);
			return [name as string, values];
		}),
	);
}

describe("page CSP", () => {
	const nonce = createNonce();
	const csp = directives(pageCsp(nonce, "/games", false));

	test("scripts only by nonce + strict-dynamic, never unsafe-inline/eval in production", () => {
		const script = csp.get("script-src") ?? [];
		expect(script).toContain(`'nonce-${nonce}'`);
		expect(script).toContain("'strict-dynamic'");
		expect(script).not.toContain("'unsafe-inline'");
		expect(script).not.toContain("'unsafe-eval'");
	});

	test("no framing, no plugins, trailers only from youtube-nocookie", () => {
		expect(csp.get("frame-ancestors")).toEqual(["'none'"]);
		expect(csp.get("object-src")).toEqual(["'none'"]);
		expect(csp.get("frame-src")).toEqual(["'self'", "https://www.youtube-nocookie.com"]);
		expect(csp.has("upgrade-insecure-requests")).toBe(true);
	});

	test("nonce is fresh and base64", () => {
		expect(createNonce()).not.toBe(nonce);
		expect(nonce).toMatch(/^[A-Za-z0-9+/]{22}==$/);
	});

	test("Scalar allowances only on its two pages, in every locale", () => {
		for (const path of ["/docs/reference", "/pt-br/docs/reference", "/es/profile/api", "/profile/api"]) {
			expect(isScalarPage(path)).toBe(true);
			expect(directives(pageCsp(nonce, path, false)).get("connect-src")).toContain("https://api.hexascore.app");
		}
		for (const path of ["/docs", "/profile", "/profile/api/checkout", "/games/docs/reference"]) {
			expect(isScalarPage(path)).toBe(false);
		}
		expect(csp.get("connect-src")).toEqual(["'self'"]);
	});

	test("dev adds unsafe-eval (React debugging) and websockets (HMR)", () => {
		const dev = directives(pageCsp(nonce, "/", true));
		expect(dev.get("script-src")).toContain("'unsafe-eval'");
		expect(dev.get("connect-src")).toContain("ws:");
	});

	test("only previews allow the Vercel Toolbar frame", () => {
		expect(directives(pageCsp(nonce, "/", false, true)).get("frame-src")).toContain("https://vercel.live");
		expect(directives(pageCsp(nonce, "/", false)).get("frame-src")).not.toContain("https://vercel.live");
	});
});

describe("next.config headers", async () => {
	const rules = (await nextConfig.headers?.()) ?? [];
	const forSource = (source: string) =>
		new Map((rules.find((r) => r.source === source)?.headers ?? []).map((h) => [h.key.toLowerCase(), h.value]));

	test("every response gets HSTS, nosniff, referrer, permissions and COOP", () => {
		const all = forSource("/:path*");
		expect(all.get("strict-transport-security")).toContain("max-age=63072000");
		expect(all.get("x-content-type-options")).toBe("nosniff");
		expect(all.get("referrer-policy")).toBe("strict-origin-when-cross-origin");
		expect(all.get("permissions-policy")).toContain("camera=()");
		expect(all.get("cross-origin-opener-policy")).toBe("same-origin");
		// A global CSP would be merged with BotID's `frame-ancestors 'self'` on its challenge path and break it.
		expect(all.has("content-security-policy")).toBe(false);
	});

	test("API responses are locked down to JSON", () => {
		expect(forSource("/api/:path*").get("content-security-policy")).toBe(API_CSP);
	});
});
