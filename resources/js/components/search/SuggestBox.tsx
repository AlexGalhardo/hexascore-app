import { router, usePage } from "@inertiajs/react";
import { Fragment, useEffect, useId, useRef, useState } from "react";

import { localized } from "../../lib/route";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";

export interface Suggestion {
	id: string;
	type: string;
	kind: string;
	title: string;
	year: number | null;
	path: string;
}

const MIN_CHARS = 2;
const DEBOUNCE_MS = 150;

// Header suggest combobox (ARIA 1.2 combobox with a listbox popup, focus stays in the input).
export default function SuggestBox() {
	const t = useT();
	const { locale } = usePage<SharedProps>().props;
	const base = useId();
	const listId = `${base}-list`;
	const [text, setText] = useState("");
	const [items, setItems] = useState<Suggestion[]>([]);
	const [open, setOpen] = useState(false);
	const [active, setActive] = useState(-1);
	const [pending, setPending] = useState(false);
	const seq = useRef(0);

	useEffect(() => {
		const q = text.trim();
		if (q.length < MIN_CHARS) {
			setItems([]);
			setOpen(false);
			setPending(false);
			return;
		}
		const ctrl = new AbortController();
		const mine = ++seq.current;
		setPending(true);
		const timer = setTimeout(async () => {
			try {
				const res = await fetch(
					`${localized("/search/suggest", locale)}?q=${encodeURIComponent(q)}`,
					{ signal: ctrl.signal, headers: { Accept: "application/json" } },
				);
				if (!res.ok) throw new Error(String(res.status));
				const body = (await res.json()) as { data: Suggestion[] };
				if (mine !== seq.current) return;
				setItems(body.data);
				setActive(-1);
				setOpen(true);
			} catch (err) {
				if ((err as Error).name === "AbortError") return;
				setItems([]);
			} finally {
				if (mine === seq.current) setPending(false);
			}
		}, DEBOUNCE_MS);
		return () => {
			clearTimeout(timer);
			ctrl.abort();
		};
	}, [text, locale]);

	const expanded = open && items.length > 0;
	const optionId = (i: number) => `${base}-opt-${i}`;

	function go(path: string) {
		setOpen(false);
		router.visit(path);
	}

	function onKeyDown(e: React.KeyboardEvent<HTMLInputElement>) {
		switch (e.key) {
			case "ArrowDown":
				e.preventDefault();
				if (!open && items.length > 0) setOpen(true);
				if (items.length > 0) setActive((a) => (a + 1) % items.length);
				break;
			case "ArrowUp":
				e.preventDefault();
				if (items.length > 0) setActive((a) => (a <= 0 ? items.length - 1 : a - 1));
				break;
			case "Enter":
				if (expanded && active >= 0 && items[active]) {
					e.preventDefault();
					go(items[active].path);
				}
				break;
			case "Escape":
				if (open) {
					e.preventDefault();
					setOpen(false);
					setActive(-1);
				}
				break;
		}
	}

	return (
		<search className="relative">
			<form
				method="get"
				action={localized("/search", locale)}
				onSubmit={(e) => {
					e.preventDefault();
					const q = text.trim();
					if (q) go(`${localized("/search", locale)}?q=${encodeURIComponent(q)}`);
				}}
			>
				<input
					type="search"
					name="q"
					role="combobox"
					aria-label={t("nav.search")}
					aria-autocomplete="list"
					aria-expanded={expanded}
					aria-controls={listId}
					aria-activedescendant={expanded && active >= 0 ? optionId(active) : undefined}
					autoComplete="off"
					placeholder={t("nav.search")}
					value={text}
					onChange={(e) => setText(e.target.value)}
					onKeyDown={onKeyDown}
					onBlur={() => setOpen(false)}
					onFocus={() => items.length > 0 && setOpen(true)}
					className="min-h-11 w-full rounded-md border border-border bg-bg px-3 text-fg sm:w-64"
				/>
				<div
					id={listId}
					role="listbox"
					aria-label={t("search.suggestions")}
					hidden={!expanded}
					className="absolute right-0 z-50 mt-1 w-72 rounded-md border border-border bg-surface shadow-lg"
				>
					{items.map((s, i) => (
						<Fragment key={s.id}>
							{/* biome-ignore lint/a11y/useFocusableInteractive: ARIA 1.2 combobox, focus stays on the input (aria-activedescendant) */}
							<div
								id={optionId(i)}
								role="option"
								aria-selected={i === active}
								// mousedown fires before the input blur closes the popup.
								onMouseDown={(e) => {
									e.preventDefault();
									go(s.path);
								}}
								className={`flex min-h-11 cursor-pointer items-center gap-2 px-3 ${i === active ? "bg-bg" : ""}`}
							>
								<span className="truncate">{s.title}</span>
								{s.year != null && (
									<span className="text-sm text-muted">{s.year}</span>
								)}
								<span className="ml-auto text-xs text-muted">
									{t(`itemTypes.${s.type}`)}
								</span>
							</div>
						</Fragment>
					))}
				</div>
				{open && items.length === 0 && (
					<p className="absolute right-0 z-50 mt-1 w-72 rounded-md border border-border bg-surface px-3 py-2 text-sm text-muted">
						{t("search.noSuggestions")}
					</p>
				)}
				<span role="status" className="sr-only">
					{pending ? t("search.searching") : ""}
				</span>
			</form>
		</search>
	);
}
