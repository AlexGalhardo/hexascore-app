import type { ExternalIds, ItemType } from "@/lib/db/schema/catalog";
import type { ExternalScore } from "./types";

export type ExternalLink = { provider: string; source: string; value: string | null; url: string };

type Provider = "metacritic" | "rottentomatoes" | "imdb" | "steam" | "goodreads" | "hltb" | "openlibrary";

const SOURCE_PROVIDER: Record<string, Provider> = {
	metacritic: "metacritic",
	metacritic_user: "metacritic",
	rottentomatoes: "rottentomatoes",
	rottentomatoes_audience: "rottentomatoes",
	imdb: "imdb",
	steam: "steam",
	goodreads: "goodreads",
	hltb_main: "hltb",
	hltb_complete: "hltb",
};

// Ids are already format-checked by parseExternalIds, so interpolating them cannot change the host.
function providerUrl(provider: Provider, type: ItemType, ids: ExternalIds): string | null {
	switch (provider) {
		case "metacritic": {
			const kind = { game: "game", movie: "movie", series: "tv", book: null }[type];
			return ids.metacritic && kind ? `https://www.metacritic.com/${kind}/${ids.metacritic}/` : null;
		}
		case "rottentomatoes": {
			const kind = { movie: "m", series: "tv", game: null, book: null }[type];
			return ids.rottentomatoes && kind ? `https://www.rottentomatoes.com/${kind}/${ids.rottentomatoes}` : null;
		}
		case "imdb":
			return ids.imdb ? `https://www.imdb.com/title/${ids.imdb}/` : null;
		case "steam":
			return ids.steam ? `https://store.steampowered.com/app/${ids.steam}/` : null;
		case "goodreads":
			return ids.goodreads ? `https://www.goodreads.com/book/show/${ids.goodreads}` : null;
		case "hltb":
			return ids.hltb ? `https://howlongtobeat.com/game/${ids.hltb}` : null;
		case "openlibrary":
			return ids.openlibrary ? `https://openlibrary.org/works/${ids.openlibrary}` : null;
	}
}

function safeHttps(url: string | null): string | null {
	if (!url) return null;
	try {
		return new URL(url).protocol === "https:" ? url : null;
	} catch {
		return null;
	}
}

const ID_ONLY: Provider[] = ["imdb", "metacritic", "rottentomatoes", "hltb", "steam", "goodreads", "openlibrary"];

/** Third-party scores first (attributed, linked), then plain links for known ids without a score. */
export function buildExternalLinks(type: ItemType, ids: ExternalIds, scores: ExternalScore[]): ExternalLink[] {
	const links: ExternalLink[] = [];
	const seen = new Set<Provider>();
	for (const score of scores) {
		const provider = SOURCE_PROVIDER[score.source];
		if (!provider) continue;
		const url = safeHttps(score.url) ?? providerUrl(provider, type, ids);
		if (!url) continue;
		seen.add(provider);
		links.push({ provider, source: score.source, value: score.value, url });
	}
	for (const provider of ID_ONLY) {
		if (seen.has(provider)) continue;
		const url = providerUrl(provider, type, ids);
		if (url) links.push({ provider, source: provider, value: null, url });
	}
	return links;
}
