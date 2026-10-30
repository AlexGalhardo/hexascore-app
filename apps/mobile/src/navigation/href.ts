import type { Href } from "expo-router";

declare module "expo-router" {
	export namespace ExpoRouter {
		interface __routes {
			href:
				| "/"
				| "/login"
				| "/search"
				| "/lists"
				| "/profile"
				| "/update-required"
				| "/games"
				| "/movies"
				| "/series"
				| "/books"
				| "/games/trending"
				| `/games/${string}`
				| `/movies/${string}`
				| `/series/${string}`
				| `/books/${string}`
				| `/entities/${string}`
				| `/lists/${string}`;
		}
	}
}

export function navigate(href: Href): Href {
	return href;
}

export const home: Href = "/";

// @ts-expect-error unknown route
export const unknownRoute: Href = "/not-a-real-route";
