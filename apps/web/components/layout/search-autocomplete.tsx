"use client";

import { Autocomplete } from "@base-ui/react/autocomplete";
import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { useLocale, useTranslations } from "next-intl";
import { useEffect, useRef, useState } from "react";
import { z } from "zod";
import { CategoryLabel } from "@/components/catalog/category-label";
import { CoverThumb } from "@/components/catalog/cover-thumb";
import { ScoreBadge } from "@/components/score/score-badge";
import { useRouter } from "@/i18n/navigation";

const suggestionSchema = z.object({
	data: z.array(
		z.object({
			id: z.string(),
			type: z.enum(["game", "movie", "series", "book"]),
			title: z.string(),
			year: z.number().nullable(),
			score: z.number().int().min(0).max(100).nullable(),
			coverUrl: z.string().nullable(),
			coverThumbUrl: z.string().nullable(),
			path: z.string().startsWith("/"),
		}),
	),
});
type Suggestion = z.infer<typeof suggestionSchema>["data"][number];

type Props = { action: string; placeholder: string };

// Server-rendered as a plain <form action="/search"> with a named input, so search works before hydration and
// without JavaScript; once hydrated, Base UI adds the accessible combobox with debounced suggestions.
export function SearchAutocomplete({ action, placeholder }: Props) {
	const t = useTranslations("search");
	const tt = useTranslations("itemTypes");
	const ts = useTranslations("score");
	const locale = useLocale();
	const router = useRouter();
	const [value, setValue] = useState("");
	const [debounced, setDebounced] = useState("");
	const highlighted = useRef<Suggestion | undefined>(undefined);
	const q = value.trim();

	useEffect(() => {
		const timer = setTimeout(() => setDebounced(q), 200);
		return () => clearTimeout(timer);
	}, [q]);

	const enabled = q.length >= 2 && debounced.length >= 2;
	const suggestions = useQuery({
		queryKey: ["search-suggest", debounced, locale],
		queryFn: async ({ signal }): Promise<Suggestion[]> => {
			try {
				const params = new URLSearchParams({ q: debounced, locale });
				const response = await fetch(`/api/search/suggest?${params}`, {
					signal: AbortSignal.any([signal, AbortSignal.timeout(5000)]),
				});
				if (!response.ok) throw new Error(`HTTP ${response.status}`);
				return suggestionSchema.parse(await response.json()).data;
			} catch (error) {
				if (!signal.aborted) console.warn("[search] suggestions unavailable", error);
				throw error;
			}
		},
		enabled,
		// Keep the previous list visible while the next keystroke loads; a failure just shows "no suggestions".
		placeholderData: keepPreviousData,
		retry: false,
	});
	const items = enabled ? (suggestions.data ?? []) : [];
	const status = !enabled ? "idle" : suggestions.isFetching ? "loading" : "done";

	return (
		<form action={action}>
			<Autocomplete.Root
				items={items}
				value={value}
				onValueChange={(next) => setValue(next)}
				itemToStringValue={(item: Suggestion) => item.title}
				filter={null}
				onItemHighlighted={(item) => {
					highlighted.current = item;
				}}
			>
				<label className="sr-only" htmlFor="site-search">
					{t("label")}
				</label>
				<Autocomplete.Input
					id="site-search"
					name="q"
					type="search"
					maxLength={100}
					placeholder={placeholder}
					// Base UI swallows Enter while the popup is open; with no highlighted suggestion it means "search".
					onKeyDown={(event) => {
						if (event.key === "Enter" && !highlighted.current) event.currentTarget.form?.requestSubmit();
					}}
					className="min-h-11 w-full rounded-md border border-border bg-bg px-3"
				/>
				<Autocomplete.Portal hidden={status === "idle"}>
					<Autocomplete.Positioner sideOffset={4} align="start" className="z-50">
						<Autocomplete.Popup className="w-[var(--anchor-width)] min-w-64 max-w-[var(--available-width)] rounded-md border border-border bg-bg py-1 shadow-lg">
							<Autocomplete.Status className="px-3 py-1 text-xs text-muted empty:hidden">
								{status === "loading" ? t("searching") : null}
							</Autocomplete.Status>
							<Autocomplete.Empty className="px-3 py-2 text-sm text-muted empty:hidden">
								{status === "done" ? t("noSuggestions") : null}
							</Autocomplete.Empty>
							<Autocomplete.List className="max-h-[32rem] overflow-y-auto">
								{(item: Suggestion) => (
									<Autocomplete.Item
										key={item.id}
										value={item}
										onClick={() => router.push(item.path)}
										className="flex min-h-14 cursor-default items-center gap-3 px-3 py-1.5 text-sm outline-none data-highlighted:bg-surface"
									>
										<CoverThumb src={item.coverUrl} thumb={item.coverThumbUrl} />
										<span className="grid min-w-0 flex-1">
											<span className="truncate font-medium">{item.title}</span>
											<CategoryLabel type={item.type} className="text-xs text-muted">
												{[tt(item.type), item.year].filter(Boolean).join(" · ")}
											</CategoryLabel>
										</span>
										<ScoreBadge
											score={item.score}
											label={
												item.score === null ? ts("none") : ts("outOf", { score: item.score })
											}
											size="sm"
										/>
									</Autocomplete.Item>
								)}
							</Autocomplete.List>
						</Autocomplete.Popup>
					</Autocomplete.Positioner>
				</Autocomplete.Portal>
			</Autocomplete.Root>
		</form>
	);
}
