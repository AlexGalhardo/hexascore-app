type Props = { label: string; items: string[] | undefined; max?: number; moreLabel: (count: number) => string };

const chip = "rounded-full border border-border bg-surface px-2.5 py-0.5 text-xs text-fg";

/** Labelled chips; past `max`, a native <details> reveals the rest (no JS, the full list stays in the HTML). */
export function ChipList({ label, items, max = 5, moreLabel }: Props) {
	if (!items?.length) return null;
	const shown = items.slice(0, max);
	const rest = items.slice(max);
	return (
		<div className="flex flex-wrap items-center gap-1.5">
			<span className="mr-1 text-xs font-semibold text-muted">{label}</span>
			{shown.map((item) => (
				<span key={item} className={chip}>
					{item}
				</span>
			))}
			{rest.length > 0 ? (
				<details className="contents [&[open]>summary]:hidden">
					<summary className={`${chip} cursor-pointer list-none text-muted hover:text-fg`}>
						{moreLabel(rest.length)}
					</summary>
					{rest.map((item) => (
						<span key={item} className={chip}>
							{item}
						</span>
					))}
				</details>
			) : null}
		</div>
	);
}
