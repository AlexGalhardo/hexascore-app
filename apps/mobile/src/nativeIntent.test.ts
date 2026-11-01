import { redirectSystemPath } from "../app/+native-intent";

const mockSetSession = jest.fn();
// The factory runs before `mockSetSession` is assigned, so it forwards instead of capturing it.
jest.mock("./auth/runtime", () => ({
	appLocale: { setSession: (locale: string) => mockSetSession(locale) },
}));

beforeEach(() => mockSetSession.mockClear());

test("a /pt-br App Link applies pt-BR to the session and drops the prefix", () => {
	expect(
		redirectSystemPath({ path: "https://hexascore.app/pt-br/games/elden-ring", initial: true }),
	).toBe("/games/elden-ring");
	expect(mockSetSession).toHaveBeenCalledWith("pt-BR");
});

test("a /es App Link applies es", () => {
	redirectSystemPath({ path: "https://hexascore.app/es/movies/arrival", initial: false });
	expect(mockSetSession).toHaveBeenCalledWith("es");
});

test("a link without a locale prefix keeps the current locale", () => {
	expect(redirectSystemPath({ path: "https://hexascore.app/books/dune", initial: false })).toBe(
		"/books/dune",
	);
	expect(mockSetSession).not.toHaveBeenCalled();
});
