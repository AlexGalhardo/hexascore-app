import { ITEM_TYPES, type ItemType } from "./db/schema/catalog";

export const SECTION_PATHS = {
	game: "games",
	movie: "movies",
	series: "series",
	book: "books",
} as const satisfies Record<ItemType, string>;

export type SectionPath = (typeof SECTION_PATHS)[ItemType];

// Literal class names so Tailwind picks them up.
export const SECTION_ACCENT = {
	game: { text: "text-game", bg: "bg-game", border: "border-game", decoration: "decoration-game" },
	movie: { text: "text-movie", bg: "bg-movie", border: "border-movie", decoration: "decoration-movie" },
	series: { text: "text-series", bg: "bg-series", border: "border-series", decoration: "decoration-series" },
	book: { text: "text-book", bg: "bg-book", border: "border-book", decoration: "decoration-book" },
} as const satisfies Record<ItemType, Record<"text" | "bg" | "border" | "decoration", string>>;

export function typeFromSection(section: string): ItemType | null {
	return ITEM_TYPES.find((t) => SECTION_PATHS[t] === section) ?? null;
}

export function itemHref(type: ItemType, slug: string): `/${SectionPath}/${string}` {
	return `/${SECTION_PATHS[type]}/${slug}`;
}
