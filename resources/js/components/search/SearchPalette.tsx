import { Dialog as BaseDialog } from "@base-ui/react/dialog";
import { router, usePage } from "@inertiajs/react";
import { type RefObject, useEffect, useId, useRef, useState } from "react";

import { localized } from "../../lib/route";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";
import { MIN_CHARS, type Suggestion, useSuggest } from "./useSuggest";

const MAX_ROWS = 8;
const RECENT_KEY = "recentSearches";
const MAX_RECENT = 5;

function readRecent(): string[] {
	try {
		const v: unknown = JSON.parse(localStorage.getItem(RECENT_KEY) ?? "[]");
		return Array.isArray(v) ? v.filter((x): x is string => typeof x === "string") : [];
	} catch {
		return [];
	}
}

function remember(q: string): void {
	try {
		const next = [q, ...readRecent().filter((x) => x !== q)].slice(0, MAX_RECENT);
		localStorage.setItem(RECENT_KEY, JSON.stringify(next));
	} catch {
		// Storage can be blocked (private mode): recent searches are a convenience only.
	}
}

type Row =
	| { kind: "item"; s: Suggestion }
	| { kind: "recent"; q: string }
	| { kind: "all"; q: string };

interface Props {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	/** Focus returns here on close (the navbar trigger). */
	returnFocus: RefObject<HTMLElement | null>;
}

// Command palette (design v1, round 2 item 16): Base UI Dialog for the focus trap, Escape and focus
// return; inside, an ARIA 1.2 combobox whose listbox holds the first 8 suggestions plus "See all".
export default function SearchPalette({ open, onOpenChange, returnFocus }: Props) {
	const t = useT();
	const { locale } = usePage<SharedProps>().props;
	const [text, setText] = useState("");
	const [active, setActive] = useState(0);
	const [recent, setRecent] = useState<string[]>([]);
	const input = useRef<HTMLInputElement>(null);
	const base = useId();
	const { items, pending, answered } = useSuggest(open ? text : "", locale);
	const q = text.trim();
	const searchHref = (s: string) => `${localized("/search", locale)}?q=${encodeURIComponent(s)}`;

	useEffect(() => {
		if (!open) return;
		setRecent(readRecent());
		setText("");
	}, [open]);

	const short = q.length < MIN_CHARS;
	const fresh = answered === q;
	const rows: Row[] = short
		? recent.map((r) => ({ kind: "recent", q: r }))
		: fresh
			? [
					...items.slice(0, MAX_ROWS).map((s): Row => ({ kind: "item", s })),
					{ kind: "all", q },
				]
			: [];

	// biome-ignore lint/correctness/useExhaustiveDependencies: reset the highlight whenever the list changes
	useEffect(() => setActive(0), [rows.length, answered, short]);

	function choose(row: Row) {
		onOpenChange(false);
		if (row.kind === "item") {
			if (q) remember(q);
			router.visit(row.s.path);
		} else {
			remember(row.q);
			router.visit(searchHref(row.q));
		}
	}

	function onKeyDown(e: React.KeyboardEvent<HTMLInputElement>) {
		if (e.key === "ArrowDown" && rows.length > 0) {
			e.preventDefault();
			setActive((a) => (a + 1) % rows.length);
		} else if (e.key === "ArrowUp" && rows.length > 0) {
			e.preventDefault();
			setActive((a) => (a <= 0 ? rows.length - 1 : a - 1));
		} else if (e.key === "Enter") {
			e.preventDefault();
			if (rows[active]) choose(rows[active]);
			else if (!short) choose({ kind: "all", q });
		}
	}

	const listId = `${base}-list`;
	const optionId = (i: number) => `${base}-opt-${i}`;
	const loading = !short && (pending || !fresh);
	const noResults = !short && fresh && items.length === 0;
	const option = (i: number) =>
		`flex min-h-11 w-full cursor-pointer items-center gap-4 px-4 py-2 text-left ${i === active ? "bg-hover outline-2 -outline-offset-2 outline-primary" : ""}`;

	return (
		<BaseDialog.Root open={open} onOpenChange={onOpenChange}>
			<BaseDialog.Portal>
				<BaseDialog.Backdrop className="fixed inset-0 z-[55] bg-backdrop" />
				<BaseDialog.Popup
					initialFocus={input}
					finalFocus={returnFocus}
					className="fixed inset-0 z-[60] flex flex-col bg-bg text-fg outline-none sm:inset-x-0 sm:top-[12vh] sm:bottom-auto sm:mx-auto sm:max-h-[76vh] sm:w-[calc(100%-2rem)] sm:max-w-[640px] sm:border sm:border-fg sm:shadow-[0_4px_0_var(--color-border)]"
				>
					<BaseDialog.Title className="sr-only">{t("search.title")}</BaseDialog.Title>
					<search className="flex items-center gap-2 border-b border-fg px-4">
						<svg
							aria-hidden="true"
							viewBox="0 0 24 24"
							className="size-5 shrink-0"
							fill="none"
						>
							<path
								d="M11 4a7 7 0 1 0 0 14a7 7 0 1 0 0-14 M16 16 L21 21"
								stroke="currentColor"
								strokeWidth="1.75"
								strokeLinecap="round"
							/>
						</svg>
						<input
							ref={input}
							type="search"
							role="combobox"
							aria-label={t("nav.search")}
							aria-autocomplete="list"
							aria-expanded={rows.length > 0}
							aria-controls={listId}
							aria-activedescendant={rows[active] ? optionId(active) : undefined}
							autoComplete="off"
							enterKeyHint="search"
							placeholder={t("nav.search")}
							value={text}
							onChange={(e) => setText(e.target.value)}
							onKeyDown={onKeyDown}
							className="h-14 min-w-0 flex-1 bg-transparent text-[17px] outline-none placeholder:text-muted [&::-webkit-search-cancel-button]:hidden"
						/>
						<BaseDialog.Close className="inline-flex min-h-11 shrink-0 items-center rounded-xs px-2 text-sm font-semibold text-primary hover:bg-hover focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fg">
							<span className="sm:hidden">{t("search.palette.close")}</span>
							<kbd className="hidden h-[22px] items-center rounded-sm border border-border bg-surface px-1.5 font-sans text-xs font-semibold text-muted sm:inline-flex">
								Esc
							</kbd>
						</BaseDialog.Close>
					</search>
					<div className="min-h-0 flex-1 overflow-y-auto py-2">
						{short && recent.length > 0 && (
							<p className="px-4 pt-1 pb-2 text-xs font-bold tracking-[0.1em] text-muted uppercase">
								{t("search.palette.recent")}
							</p>
						)}
						{short && recent.length === 0 && (
							<p className="px-4 py-3 text-[15px] text-muted">{t("search.prompt")}</p>
						)}
						{loading && (
							<ul
								aria-hidden="true"
								className="animate-pulse motion-reduce:animate-none"
							>
								{[0, 1, 2].map((i) => (
									<li key={i} className="flex items-center gap-4 px-4 py-2">
										<span className="h-12 w-8 bg-surface" />
										<span className="h-4 w-1/2 bg-surface" />
									</li>
								))}
							</ul>
						)}
						{noResults && (
							<p className="px-4 pt-3 pb-1 text-[15px]">
								{t("search.noResults", { q })}
							</p>
						)}
						<div id={listId} role="listbox" aria-label={t("search.suggestions")}>
							{rows.map((row, i) => (
								// biome-ignore lint/a11y/useFocusableInteractive: ARIA 1.2 combobox, focus stays on the input (aria-activedescendant)
								// biome-ignore lint/a11y/useKeyWithClickEvents: the input handles the keys for every option
								<div
									key={row.kind === "item" ? row.s.id : `${row.kind}-${row.q}`}
									id={optionId(i)}
									role="option"
									aria-selected={i === active}
									onMouseMove={() => setActive(i)}
									onClick={() => choose(row)}
									className={`${option(i)} ${row.kind === "all" ? "mt-1 border-t border-border font-semibold text-primary" : ""}`}
								>
									{row.kind === "item" && (
										<>
											{row.s.coverThumbUrl ? (
												<img
													src={row.s.coverThumbUrl}
													alt=""
													width={32}
													height={48}
													loading="lazy"
													className="aspect-[2/3] w-8 shrink-0 bg-surface object-cover"
												/>
											) : (
												<span className="aspect-[2/3] w-8 shrink-0 border border-border bg-surface" />
											)}
											<span className="min-w-0 flex-1">
												<span className="block truncate font-display text-base font-bold">
													{row.s.title}
													{row.s.year != null && (
														<span className="ml-2 font-sans text-sm font-normal text-muted tabular-nums">
															{row.s.year}
														</span>
													)}
												</span>
												<span className="block text-[13px] text-muted">
													{t(`itemTypes.${row.s.type}`)}
												</span>
											</span>
										</>
									)}
									{row.kind === "recent" && (
										<span className="truncate">{row.q}</span>
									)}
									{row.kind === "all" && (
										<span>{t("search.palette.seeAll", { q: row.q })}</span>
									)}
								</div>
							))}
						</div>
					</div>
					<span role="status" className="sr-only">
						{loading
							? t("search.searching")
							: noResults
								? t("search.noSuggestions")
								: ""}
					</span>
				</BaseDialog.Popup>
			</BaseDialog.Portal>
		</BaseDialog.Root>
	);
}
