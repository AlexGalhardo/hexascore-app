import { type Faker, fakerEN, fakerES, fakerPT_BR } from "@faker-js/faker";
import type { ExternalIds, ItemMeta, ItemType } from "./schema/catalog";

// Deterministic fake catalog for dev/preview. Every record is derived from (SEED, type, index) alone, so generating
// item #n never depends on how many items come before it: `--items` can grow without reshuffling existing rows.

export const SEED = 20261002;
const TYPE_INDEX: Record<ItemType, number> = { game: 1, movie: 2, series: 3, book: 4 };

const GENRES: Record<ItemType, string[]> = {
	game: [
		"Action",
		"Adventure",
		"RPG",
		"Strategy",
		"Shooter",
		"Puzzle",
		"Platformer",
		"Racing",
		"Simulation",
		"Horror",
	],
	movie: ["Drama", "Comedy", "Thriller", "Science Fiction", "Horror", "Animation", "Documentary", "Romance", "Crime"],
	series: ["Drama", "Comedy", "Crime", "Fantasy", "Science Fiction", "Mystery", "Animation", "Reality", "Thriller"],
	book: ["Fiction", "Fantasy", "Mystery", "Romance", "Biography", "History", "Science Fiction", "Poetry", "Horror"],
};
const PLATFORMS = ["PC", "PlayStation 5", "PlayStation 4", "Xbox Series X|S", "Nintendo Switch", "macOS", "iOS"];
const COUNTRIES = ["US", "GB", "BR", "ES", "JP", "FR", "KR", "MX", "AR", "DE"];

export type SeedItem = {
	wikidataQid: string;
	type: ItemType;
	slugBase: string;
	releaseDate: string;
	popularity: number;
	externalIds: ExternalIds;
	meta: ItemMeta;
	translations: {
		locale: "en" | "pt-BR" | "es";
		title: string;
		synopsis: string | null;
		synopsisLicense: string | null;
		synopsisSourceUrl: string | null;
	}[];
	externalScores: { source: string; value: string }[];
	affiliateLinks: { store: "amazon" | "mercadolivre" | "steam" | "gog"; url: string }[];
	/** Deterministic rating plan: how many users rate it and the item's "true" quality. */
	ratingCount: number;
	quality: number;
};

export function slugify(text: string): string {
	return text
		.normalize("NFKD")
		.replace(/[̀-ͯ]/g, "")
		.toLowerCase()
		.replace(/[^a-z0-9]+/g, "-")
		.replace(/^-+|-+$/g, "")
		.slice(0, 80);
}

const cap = (s: string) => s.charAt(0).toUpperCase() + s.slice(1);

function englishTitle(f: Faker, type: ItemType): string {
	const adj = cap(f.word.adjective());
	const noun = cap(f.word.noun());
	const other = cap(f.word.noun());
	const patterns: Record<ItemType, string[]> = {
		game: [`${adj} ${noun}`, `${noun} Tactics`, `${adj} ${noun} II`, `${noun}: ${adj} ${other}`],
		movie: [`The ${adj} ${noun}`, `${noun} of ${other}`, `A ${adj} ${noun}`, `${noun}`],
		series: [`${noun} Street`, `The ${noun}s`, `${adj} ${noun}`, `House of ${noun}`],
		book: [`The ${noun} of ${other}`, `${adj} ${noun}`, `Letters from ${f.location.city()}`, `The ${adj} ${noun}`],
	};
	return f.helpers.arrayElement(patterns[type]);
}

function fakeId(f: Faker, length: number): string {
	return f.string.numeric({ length, allowLeadingZeros: false });
}

export function generateItem(type: ItemType, index: number, users: number): SeedItem {
	const seed = [SEED, TYPE_INDEX[type], index];
	const f = fakerEN;
	f.seed(seed);
	fakerPT_BR.seed(seed);
	fakerES.seed(seed);

	const title = englishTitle(f, type);
	const slugBase = slugify(title) || `${type}-${index}`;
	const releaseDate = f.date
		.between({ from: "1975-01-01T00:00:00Z", to: "2026-09-30T00:00:00Z" })
		.toISOString()
		.slice(0, 10);
	// Zipf-like popularity: a few hits, a long tail (drives how many seed users rate the item).
	const popularity = Math.round((1000 / (index + 1) ** 0.8) * f.number.float({ min: 0.8, max: 1.2 }) * 100) / 100;
	const ratingCount = Math.round(Math.min(0.9, popularity / 300) * users);

	const meta: ItemMeta = {
		genres: f.helpers.arrayElements(GENRES[type], { min: 1, max: 3 }),
		creators: type === "game" ? [f.company.name()] : f.helpers.multiple(() => f.person.fullName(), { count: 2 }),
		countries: [f.helpers.arrayElement(COUNTRIES)],
	};
	if (type === "game") meta.platforms = f.helpers.arrayElements(PLATFORMS, { min: 1, max: 4 });
	if (type === "movie") meta.runtimeMinutes = f.number.int({ min: 80, max: 190 });
	if (type === "series") {
		meta.seasons = f.number.int({ min: 1, max: 9 });
		meta.episodes = meta.seasons * f.number.int({ min: 6, max: 22 });
	}
	if (type === "book") meta.pages = f.number.int({ min: 90, max: 1100 });
	if (type !== "book") {
		// FAKE ids: never resolve on YouTube; the 11-char shape only exercises the embed UI.
		meta.trailers = f.helpers.multiple(
			() => ({ youtubeId: `FAKE${f.string.alphanumeric(7)}`, title: "Seed trailer (fake id)" }),
			{ count: f.number.int({ min: 0, max: 3 }) },
		);
	}

	const externalIds: ExternalIds = {};
	if (type === "movie" || type === "series") {
		externalIds.imdb = `tt9${fakeId(f, 7)}`;
		externalIds.rottentomatoes = slugBase.replace(/-/g, "_");
	}
	if (type !== "book") externalIds.metacritic = slugBase;
	if (type === "game") {
		externalIds.hltb = fakeId(f, 6);
		if (f.datatype.boolean()) externalIds.steam = fakeId(f, 7);
	}
	if (type === "book") {
		externalIds.isbn13 = `978${fakeId(f, 10)}`;
		externalIds.goodreads = fakeId(f, 8);
		externalIds.openlibrary = `OL${fakeId(f, 7)}W`;
	}

	const sourceUrl = `https://example.test/wiki/${slugBase}`;
	const translations: SeedItem["translations"] = [
		{
			locale: "en",
			title,
			synopsis: `${f.company.catchPhrase()}. ${f.commerce.productDescription()}`,
			synopsisLicense: "CC BY-SA 4.0",
			synopsisSourceUrl: sourceUrl,
		},
	];
	if (f.number.int(9) < 8) {
		translations.push({
			locale: "pt-BR",
			title: f.datatype.boolean() ? `${cap(fakerPT_BR.word.noun())} ${fakerPT_BR.word.adjective()}` : title,
			synopsis: fakerPT_BR.commerce.productDescription(),
			synopsisLicense: "CC BY-SA 4.0",
			synopsisSourceUrl: sourceUrl,
		});
	}
	// ~30% without Spanish: exercises the English fallback notice.
	if (f.number.int(9) < 7) {
		translations.push({
			locale: "es",
			title: f.datatype.boolean() ? fakerES.commerce.productName() : title,
			synopsis: f.datatype.boolean() ? fakerES.commerce.productDescription() : null,
			synopsisLicense: "CC BY-SA 4.0",
			synopsisSourceUrl: sourceUrl,
		});
	}

	const externalScores: SeedItem["externalScores"] = [];
	if (f.number.int(9) < 3) {
		if (type !== "book")
			externalScores.push({ source: "metacritic", value: String(f.number.int({ min: 40, max: 95 })) });
		if (type === "movie" || type === "series") {
			externalScores.push({
				source: "imdb",
				value: f.number.float({ min: 4, max: 9.4, fractionDigits: 1 }).toFixed(1),
			});
			externalScores.push({ source: "rottentomatoes", value: `${f.number.int({ min: 20, max: 100 })}%` });
		}
		if (type === "game")
			externalScores.push({ source: "hltb_main", value: `${f.number.int({ min: 2, max: 80 })} h` });
		if (type === "book") {
			externalScores.push({
				source: "goodreads",
				value: f.number.float({ min: 2.5, max: 4.7, fractionDigits: 2 }).toFixed(2),
			});
		}
	}

	const affiliateLinks: SeedItem["affiliateLinks"] = [];
	if ((type === "book" || type === "game") && f.datatype.boolean()) {
		const stores = type === "book" ? (["amazon", "mercadolivre"] as const) : (["steam", "gog"] as const);
		for (const store of stores)
			affiliateLinks.push({ store, url: `https://store.example.test/${store}/${slugBase}` });
	}

	return {
		wikidataQid: `seed-${type}-${String(index).padStart(6, "0")}`,
		type,
		slugBase,
		releaseDate,
		popularity,
		externalIds,
		meta,
		translations,
		externalScores,
		affiliateLinks,
		ratingCount,
		quality: f.number.int({ min: 25, max: 95 }),
	};
}

export type SeedUser = {
	id: string;
	name: string;
	email: string;
	username: string;
	profilePrivate: boolean;
};

export function seedUserId(index: number): string {
	return `seed-user-${String(index).padStart(6, "0")}`;
}

export function generateUser(index: number): SeedUser {
	fakerEN.seed([SEED, 99, index]);
	const first = fakerEN.person.firstName();
	const id = seedUserId(index);
	return {
		id,
		name: `${first} ${fakerEN.person.lastName()}`,
		email: `${id}@example.test`,
		// nickname rule: 4–16 chars of [A-Za-z0-9_]; the index suffix keeps it unique.
		username: `${slugify(first).replace(/-/g, "").slice(0, 9) || "user"}_${index}`.slice(0, 16),
		profilePrivate: index % 7 === 0,
	};
}

/** c1..c6 for user `u` rating an item of the given quality — deterministic, clamped to 0..100. */
export function generateRating(item: { wikidataQid: string; quality: number }, userIndex: number): number[] {
	fakerEN.seed([SEED, 77, userIndex, ...[...item.wikidataQid].map((c) => c.charCodeAt(0))]);
	const mood = fakerEN.number.int({ min: -12, max: 12 });
	return Array.from({ length: 6 }, () =>
		Math.max(0, Math.min(100, item.quality + mood + fakerEN.number.int({ min: -15, max: 15 }))),
	);
}

/** Which user indexes (0-based) rate the item: a deterministic stride through the user list. */
export function raterIndexes(itemOrdinal: number, count: number, users: number): number[] {
	if (users === 0) return [];
	const start = (itemOrdinal * 7919) % users;
	return Array.from({ length: Math.min(count, users) }, (_, k) => (start + k) % users);
}
