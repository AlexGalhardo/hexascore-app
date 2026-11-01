import { useEffect, useRef, useState } from "react";

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

/** Debounced /search/suggest lookup; stale answers are dropped. */
export function useSuggest(text: string, locale: string, type?: string): SuggestState {
	const [state, setState] = useState<SuggestState>({ items: [], pending: false, answered: "" });
	const seq = useRef(0);

	useEffect(() => {
		const q = text.trim();
		if (q.length < MIN_CHARS) {
			seq.current++;
			setState({ items: [], pending: false, answered: "" });
			return;
		}
		const ctrl = new AbortController();
		const mine = ++seq.current;
		setState((s) => ({ ...s, pending: true }));
		const timer = setTimeout(async () => {
			try {
				const res = await fetch(
					`${localized("/search/suggest", locale)}?q=${encodeURIComponent(q)}${type ? `&type=${encodeURIComponent(type)}` : ""}`,
					{ signal: ctrl.signal, headers: { Accept: "application/json" } },
				);
				if (!res.ok) throw new Error(String(res.status));
				const body = (await res.json()) as { data: Suggestion[] };
				if (mine === seq.current)
					setState({ items: body.data, pending: false, answered: q });
			} catch (err) {
				if ((err as Error).name === "AbortError") return;
				if (mine === seq.current) setState({ items: [], pending: false, answered: q });
			}
		}, DEBOUNCE_MS);
		return () => {
			clearTimeout(timer);
			ctrl.abort();
		};
	}, [text, locale, type]);

	return state;
}
