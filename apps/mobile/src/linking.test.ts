import config from "../app.config";
import { APP_LINK_PREFIXES, resolveDeepLink } from "./linking";

test.each([
	["https://hexascore.app/games/elden-ring", "/games/elden-ring", null],
	["https://hexascore.app/pt-br/games/elden-ring", "/games/elden-ring", "pt-BR"],
	["https://hexascore.app/es/movies/arrival?x=1", "/movies/arrival?x=1", "es"],
	["https://hexascore.app/entities/42", "/entities/42", null],
	["https://hexascore.app/lists/7", "/lists/7", null],
	["https://hexascore.app/pt-br", "/", "pt-BR"],
	["/books/dune", "/books/dune", null],
	["hexascore://series/dark", "/series/dark", null],
	// Anything outside the declared prefixes lands on home instead of an unknown-route screen.
	["https://hexascore.app/pricing", "/", null],
	["https://hexascore.app/es/admin", "/", "es"],
])("%s -> %s (%s)", (url, path, locale) => {
	expect(resolveDeepLink(url)).toEqual({ path, locale });
});

test("app.config declares verified App Links for every prefix on hexascore.app", () => {
	const filters = config.android?.intentFilters ?? [];
	expect(filters).toHaveLength(1);
	const [filter] = filters;
	expect(filter?.autoVerify).toBe(true);
	expect(filter?.action).toBe("VIEW");
	expect(filter?.category).toEqual(["BROWSABLE", "DEFAULT"]);
	const data = Array.isArray(filter?.data) ? filter.data : [];
	expect(data.map((d) => [d.scheme, d.host, d.pathPrefix])).toEqual(
		APP_LINK_PREFIXES.map((prefix) => ["https", "hexascore.app", prefix]),
	);
	expect(APP_LINK_PREFIXES).toEqual([
		"/games",
		"/movies",
		"/series",
		"/books",
		"/entities",
		"/lists",
		"/pt-br",
		"/es",
	]);
});
