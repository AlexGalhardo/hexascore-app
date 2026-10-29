import { describe, expect, it } from "vitest";

import { localized } from "./route";

describe("localized", () => {
	it("leaves the default locale unprefixed", () => {
		expect(localized("/items?q=a", "en")).toBe("/items?q=a");
	});

	it("prefixes pt-BR and es", () => {
		expect(localized("/", "pt-BR")).toBe("/pt-br/");
		expect(localized("/items/1#top", "es")).toBe("/es/items/1#top");
	});

	it("never prefixes external or protocol-relative URLs or unknown locales", () => {
		expect(localized("https://example.com/x", "pt-BR")).toBe("https://example.com/x");
		expect(localized("//evil.test/x", "pt-BR")).toBe("//evil.test/x");
		expect(localized("/items", "xx")).toBe("/items");
	});
});
