import type { CatalogType } from "../navigation/catalog";

// Mirror of app/scoring CategoryKeys (A-06), in column order c1..c6. The /app/v1 OpenAPI types the
// rating body as `unknown`, so the generated client carries no enum to read them from.
// shortcut: hand-kept copy, replace with the generated enum once the rating body gets a schema.
export const CATEGORY_KEYS = {
	games: ["gameplay", "fun", "sound", "technical", "visual", "storytelling"],
	movies: ["screenplay", "acting", "direction", "cinematography", "soundtrack", "entertainment"],
	series: ["screenplay", "acting", "characters", "pacing", "production", "entertainment"],
	books: ["plot", "characters", "writing", "pacing", "originality", "impact"],
} as const satisfies Record<CatalogType, readonly string[]>;
