import { type ReactElement, useEffect, useState } from "react";

export interface ProfileNavItem {
	id: string;
	label: string;
}

/**
 * The profile's section menu: a sticky column from lg up, a scrolling row above the content on small screens. The
 * section in view is marked with aria-current; the links are plain anchors, so they work without scripting.
 */
export default function ProfileNav({
	items,
	label,
}: {
	items: ProfileNavItem[];
	label: string;
}): ReactElement {
	const [active, setActive] = useState(items[0]?.id ?? "");

	useEffect(() => {
		const hash = window.location.hash.slice(1);
		if (items.some((i) => i.id === hash)) setActive(hash);
		if (typeof IntersectionObserver === "undefined") return;
		const seen = new Set<string>();
		const observer = new IntersectionObserver(
			(entries) => {
				for (const e of entries) {
					if (e.isIntersecting) seen.add(e.target.id);
					else seen.delete(e.target.id);
				}
				const first = items.find((i) => seen.has(i.id));
				if (first) setActive(first.id);
			},
			{ rootMargin: "-15% 0px -65% 0px" },
		);
		for (const i of items) {
			const el = document.getElementById(i.id);
			if (el) observer.observe(el);
		}
		return () => observer.disconnect();
	}, [items]);

	return (
		<nav aria-label={label} className="min-w-0 lg:sticky lg:top-24">
			<ul className="flex gap-1 overflow-x-auto pb-1 lg:flex-col lg:overflow-visible lg:pb-0">
				{items.map((i) => (
					<li key={i.id} className="shrink-0">
						<a
							href={`#${i.id}`}
							aria-current={active === i.id ? "true" : undefined}
							onClick={() => setActive(i.id)}
							className="block rounded-md px-3 py-2 text-sm whitespace-nowrap text-muted hover:bg-hover hover:text-fg aria-[current=true]:bg-surface aria-[current=true]:font-semibold aria-[current=true]:text-fg"
						>
							{i.label}
						</a>
					</li>
				))}
			</ul>
		</nav>
	);
}
