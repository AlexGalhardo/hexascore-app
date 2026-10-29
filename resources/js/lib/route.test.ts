import { describe, expect, it } from "vitest";

import { localized, unlocalized } from "./route";

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

describe("unlocalized", () => {
	it("strips a known prefix only at a segment boundary", () => {
		expect(unlocalized("/pt-br/games/x?a=1")).toBe("/games/x?a=1");
		expect(unlocalized("/es")).toBe("/");
		expect(unlocalized("/es/")).toBe("/");
		expect(unlocalized("/estrella")).toBe("/estrella");
		expect(unlocalized("/games")).toBe("/games");
	});
});
