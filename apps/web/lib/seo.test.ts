import { describe, expect, test } from "bun:test";
import { disallowedPaths, itemEntries, localizedUrl, SITEMAP_CHUNK, sitemapEntry, sitemapIds } from "./seo";

const BASE = "https://hexascore.app";

describe("seo", () => {
	test("localized urls follow the as-needed prefixes", () => {
		expect(localizedUrl(BASE, "en", "/")).toBe("https://hexascore.app/");
		expect(localizedUrl(BASE, "pt-BR", "/")).toBe("https://hexascore.app/pt-br");
		expect(localizedUrl(BASE, "es", "/games")).toBe("https://hexascore.app/es/games");
	});

	test("entries carry every locale plus x-default as alternates", () => {
		expect(sitemapEntry(BASE, "/pricing").alternates?.languages).toEqual({
			en: "https://hexascore.app/pricing",
			"pt-BR": "https://hexascore.app/pt-br/pricing",
			es: "https://hexascore.app/es/pricing",
			"x-default": "https://hexascore.app/pricing",
		});
		const [item] = itemEntries(BASE, [{ type: "game", slug: "zelda", updatedAt: "2026-10-01T00:00:00.000Z" }]);
		expect(item).toMatchObject({
			url: "https://hexascore.app/games/zelda",
			lastModified: "2026-10-01T00:00:00.000Z",
		});
	});

	test("sitemap ids: static file plus one per item chunk", () => {
		expect(sitemapIds(0)).toEqual([0]);
		expect(sitemapIds(1)).toEqual([0, 1]);
		expect(sitemapIds(SITEMAP_CHUNK)).toEqual([0, 1]);
		expect(sitemapIds(SITEMAP_CHUNK + 1)).toEqual([0, 1, 2]);
	});

	test("private pages are disallowed in every locale", () => {
		const paths = disallowedPaths();
		for (const p of ["/api/", "/admin", "/pt-br/profile", "/es/feed", "/search"]) expect(paths).toContain(p);
		expect(paths).not.toContain("/games");
	});
});
