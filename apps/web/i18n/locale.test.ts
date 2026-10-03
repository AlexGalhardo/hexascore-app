import { describe, expect, test } from "bun:test";
import { localeFromRequest } from "./locale";

const req = (headers: Record<string, string>): Request => new Request("http://localhost/api/auth/x", { headers });

describe("localeFromRequest", () => {
	test("prefers the locale cookie", () => {
		expect(localeFromRequest(req({ cookie: "a=1; locale=es", referer: "http://localhost/pt-br/login" }))).toBe(
			"es",
		);
	});

	test("falls back to the referer prefix", () => {
		expect(localeFromRequest(req({ referer: "http://localhost/pt-br/signup" }))).toBe("pt-BR");
		expect(localeFromRequest(req({ referer: "http://localhost/es" }))).toBe("es");
	});

	test("ignores unknown cookies and look-alike paths, then uses the default", () => {
		expect(localeFromRequest(req({ cookie: "locale=fr", referer: "http://localhost/espanol" }))).toBe("en");
		expect(localeFromRequest(req({ referer: "not a url" }))).toBe("en");
		expect(localeFromRequest()).toBe("en");
	});
});
