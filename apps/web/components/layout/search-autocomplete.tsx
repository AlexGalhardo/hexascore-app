"use client";

import { Autocomplete } from "@base-ui/react/autocomplete";
import { useLocale, useTranslations } from "next-intl";
import { useEffect, useRef, useState } from "react";
import { z } from "zod";
import { useRouter } from "@/i18n/navigation";

const suggestionSchema = z.object({
	data: z.array(
		z.object({
			id: z.string(),
			type: z.enum(["game", "movie", "series", "book"]),
			title: z.string(),
			year: z.number().nullable(),
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
	const locale = useLocale();
	const router = useRouter();
	const [value, setValue] = useState("");
	const [items, setItems] = useState<Suggestion[]>([]);
	const [status, setStatus] = useState<"idle" | "loading" | "done">("idle");
	const controller = useRef<AbortController | null>(null);
	const highlighted = useRef<Suggestion | undefined>(undefined);

	useEffect(() => {
		const q = value.trim();
		if (q.length < 2) {
			setItems([]);
			setStatus("idle");
			return;
		}
		const timer = setTimeout(async () => {
			controller.current?.abort();
			const current = new AbortController();
			controller.current = current;
			setStatus("loading");
			try {
				const params = new URLSearchParams({ q, locale });
				const response = await fetch(`/api/search/suggest?${params}`, {
					signal: AbortSignal.any([current.signal, AbortSignal.timeout(5000)]),
				});
				if (!response.ok) throw new Error(`HTTP ${response.status}`);
				setItems(suggestionSchema.parse(await response.json()).data);
				setStatus("done");
			} catch (error) {
				if (current.signal.aborted) return;
				console.warn("[search] suggestions unavailable", error);
				setItems([]);
				setStatus("done");
			}
		}, 200);
		return () => clearTimeout(timer);
	}, [value, locale]);

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
							<Autocomplete.List className="max-h-80 overflow-y-auto">
								{(item: Suggestion) => (
									<Autocomplete.Item
										key={item.id}
										value={item}
										onClick={() => router.push(item.path)}
										className="flex min-h-11 cursor-default items-center justify-between gap-3 px-3 text-sm outline-none data-highlighted:bg-surface"
									>
										<span className="truncate">{item.title}</span>
										<span className="shrink-0 text-xs text-muted">
											{[tt(item.type), item.year].filter(Boolean).join(" · ")}
										</span>
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
