import { router, usePage } from "@inertiajs/react";
import { Fragment, useEffect, useId, useState } from "react";
import { withQueryClient } from "../../lib/queryClient";

import { localized } from "../../lib/route";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";
import { type Suggestion, useSuggest } from "./useSuggest";

export type { Suggestion } from "./useSuggest";

interface Props {
	// Replaces the default navigation to the suggestion (and to /search on submit): the box becomes a picker.
	onPick?: (s: Suggestion) => void;
	// Restricts suggestions to one item type.
	type?: string;
	label?: string;
}

// Suggest combobox (ARIA 1.2 combobox with a listbox popup, focus stays in the input). Header search by default.
function SuggestBox({ onPick, type, label }: Props) {
	const t = useT();
	const { locale } = usePage<SharedProps>().props;
	const base = useId();
	const listId = `${base}-list`;
	const [text, setText] = useState("");
	const { items, pending, answered } = useSuggest(text, locale, type);
	const [open, setOpen] = useState(false);
	const [active, setActive] = useState(-1);

	// A fresh answer opens the popup with nothing highlighted; a short query closes it.
	// biome-ignore lint/correctness/useExhaustiveDependencies: a new answer for the same query must reopen the popup
	useEffect(() => {
		setActive(-1);
		setOpen(answered !== "");
	}, [answered, items]);

	const expanded = open && items.length > 0;
	const optionId = (i: number) => `${base}-opt-${i}`;

	function go(path: string) {
		setOpen(false);
		router.visit(path);
	}

	function pick(s: Suggestion) {
		if (!onPick) return go(s.path);
		setOpen(false);
		setText("");
		onPick(s);
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
					pick(items[active]);
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
					if (q && !onPick)
						go(`${localized("/search", locale)}?q=${encodeURIComponent(q)}`);
				}}
			>
				<input
					type="search"
					name="q"
					role="combobox"
					aria-label={label ?? t("nav.search")}
					aria-autocomplete="list"
					aria-expanded={expanded}
					aria-controls={listId}
					aria-activedescendant={expanded && active >= 0 ? optionId(active) : undefined}
					autoComplete="off"
					placeholder={label ?? t("nav.search")}
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
									pick(s);
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

export default withQueryClient(SuggestBox);
