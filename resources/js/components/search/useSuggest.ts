import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { useEffect, useState } from "react";

import { localized } from "../../lib/route";

export interface Suggestion {
	id: string;
	type: string;
	kind: string;
	title: string;
	year: number | null;
	coverThumbUrl?: string;
	path: string;
}

export const MIN_CHARS = 2;
const DEBOUNCE_MS = 150;

export interface SuggestState {
	items: Suggestion[];
	pending: boolean;
	/** The query the current items answer (empty until the first answer). */
	answered: string;
}

interface Answer {
	q: string;
	items: Suggestion[];
}

/** Debounced, cached /search/suggest lookup; a superseded request is aborted. */
export function useSuggest(text: string, locale: string, type?: string): SuggestState {
	const q = text.trim();
	const [debounced, setDebounced] = useState("");
	useEffect(() => {
		const timer = setTimeout(() => setDebounced(q), DEBOUNCE_MS);
		return () => clearTimeout(timer);
	}, [q]);

	const enabled = debounced.length >= MIN_CHARS;
	const query = useQuery({
		queryKey: ["suggest", locale, type ?? "", debounced],
		queryFn: async ({ signal }): Promise<Answer> => {
			const res = await fetch(
				`${localized("/search/suggest", locale)}?q=${encodeURIComponent(debounced)}${type ? `&type=${encodeURIComponent(type)}` : ""}`,
				{ signal, headers: { Accept: "application/json" } },
			);
			if (!res.ok) throw new Error(String(res.status));
			const body = (await res.json()) as { data: Suggestion[] };
			return { q: debounced, items: body.data };
		},
		enabled,
		// The next keystroke is the retry.
		retry: false,
		placeholderData: keepPreviousData,
	});

	if (q.length < MIN_CHARS) return { items: [], pending: false, answered: "" };
	const pending = debounced !== q || query.isFetching;
	if (query.isError && !query.isFetching) return { items: [], pending, answered: debounced };
	return { items: query.data?.items ?? [], pending, answered: query.data?.q ?? "" };
}
