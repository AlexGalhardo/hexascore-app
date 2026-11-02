import config from "../app.config";
import { APP_LINK_PATTERNS, resolveDeepLink } from "./linking";

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
	// Entity pages live under their category on the web (owner round 2026-11-01 item 19); the app has no slug route.
	["https://hexascore.app/games/platforms/playstation-5", "/", null],
	["https://hexascore.app/pt-br/movies/people/ridley-scott", "/", "pt-BR"],
	["https://hexascore.app/books/dune/editions/1", "/", null],
])("%s -> %s (%s)", (url, path, locale) => {
	expect(resolveDeepLink(url)).toEqual({ path, locale });
});

// Android pathAdvancedPattern matches the whole path; the regex subset used here is identical in JS.
const opensApp = (path: string): boolean =>
	APP_LINK_PATTERNS.some((pattern) => new RegExp(`^${pattern}$`).test(path));

test.each([
	["/games/elden-ring", true],
	["/games/elden-ring/", true],
	["/pt-br/movies/arrival", true],
	["/es/series/dark", true],
	["/entities/42", true],
	["/pt-br/lists/7", true],
	["/pt-br", true],
	["/es", true],
	["/games/platforms/playstation-5", false],
	["/movies/people/ridley-scott", false],
	["/pt-br/movies/people/ridley-scott", false],
	["/es/games/studios/fromsoftware", false],
	["/books/dune/editions/1", false],
	["/pricing", false],
	["/pt-br/admin", false],
])("App Links %s opens the app: %s", (path, expected) => {
	expect(opensApp(path)).toBe(expected);
});

test("app.config declares verified App Links for exactly APP_LINK_PATTERNS on hexascore.app", () => {
	const filters = config.android?.intentFilters ?? [];
	expect(filters).toHaveLength(1);
	const [filter] = filters;
	expect(filter?.autoVerify).toBe(true);
	expect(filter?.action).toBe("VIEW");
	expect(filter?.category).toEqual(["BROWSABLE", "DEFAULT"]);
	const data = Array.isArray(filter?.data) ? filter.data : [];
	expect(data.map((d) => [d.scheme, d.host, d.pathAdvancedPattern])).toEqual(
		APP_LINK_PATTERNS.map((pattern) => ["https", "hexascore.app", pattern]),
	);
});
