import { type UseQueryResult, useQuery } from "@tanstack/react-query";
import { useEffect, useState } from "react";
import type { SuggestJSON } from "../../../../packages/app-client/src/api";
import type { ReadClient } from "../api/client";
import type { AppLocale } from "../i18n";

// Same floor as the server's tooShort answer (SDD 07): shorter queries are never sent.
export const MIN_QUERY = 2;
// The server answers 400 above 100 runes (SDD 07 query normalisation), so the client cuts first.
export const MAX_QUERY = 100;

/** Trims and keeps the first MAX_QUERY characters (code points, so an emoji is never split). */
export function cutQuery(text: string): string {
	return [...text.trim()].slice(0, MAX_QUERY).join("");
}

export function useDebounced(value: string, delay: number): string {
	const [settled, setSettled] = useState(value);
	useEffect(() => {
		const timer = setTimeout(() => setSettled(value), delay);
		return () => clearTimeout(timer);
	}, [value, delay]);
	return settled;
}

/** GET /search/suggest for the settled, cut query: one call per debounce window. */
export function useSuggest(
	read: Pick<ReadClient, "searchSuggest">,
	input: { query: string; type?: string; locale: AppLocale; delay: number; enabled: boolean },
): { q: string; suggest: UseQueryResult<SuggestJSON> } {
	const q = useDebounced(cutQuery(input.query), input.delay);
	const { type, locale } = input;
	const suggest = useQuery({
		queryKey: ["searchSuggest", { q, type, locale }] as const,
		enabled: input.enabled && q.length >= MIN_QUERY,
		queryFn: () => read.searchSuggest({ query: { q, ...(type ? { type } : {}), locale } }),
	});
	return { q, suggest };
}
