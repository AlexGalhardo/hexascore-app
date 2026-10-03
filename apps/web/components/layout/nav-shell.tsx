"use client";

import { type ReactNode, useEffect, useState } from "react";
import { usePathname } from "@/i18n/navigation";

type Props = {
	brand: ReactNode;
	search: ReactNode;
	actions: ReactNode;
	links: ReactNode;
	labels: { search: string; menu: string };
};

const iconButton = "grid size-11 place-items-center rounded-md border border-border hover:bg-surface sm:hidden";

// Desktop shows everything inline. On mobile the bar keeps only the brand, a search button and a hamburger;
// the same server-rendered nodes are revealed below it, so there is one search input and one set of links.
export function NavShell({ brand, search, actions, links, labels }: Props) {
	const [open, setOpen] = useState<"search" | "menu" | null>(null);
	const pathname = usePathname();

	// biome-ignore lint/correctness/useExhaustiveDependencies: close the panels whenever the route changes.
	useEffect(() => setOpen(null), [pathname]);

	useEffect(() => {
		if (open === "search") document.getElementById("site-search")?.focus();
	}, [open]);

	const toggle = (panel: "search" | "menu") => setOpen((current) => (current === panel ? null : panel));

	return (
		<div className="mx-auto flex max-w-6xl flex-wrap items-center gap-x-3 gap-y-2 px-4 py-3">
			{brand}
			<div className="ml-auto flex items-center gap-2 sm:hidden">
				<button
					type="button"
					aria-label={labels.search}
					aria-expanded={open === "search"}
					aria-controls="nav-search"
					onClick={() => toggle("search")}
					className={iconButton}
				>
					<svg
						aria-hidden="true"
						viewBox="0 0 24 24"
						className="size-5"
						fill="none"
						stroke="currentColor"
						strokeWidth="2"
					>
						<circle cx="11" cy="11" r="7" />
						<path d="m20 20-3.5-3.5" />
					</svg>
				</button>
				<button
					type="button"
					aria-label={labels.menu}
					aria-expanded={open === "menu"}
					aria-controls="nav-menu"
					onClick={() => toggle("menu")}
					className={iconButton}
				>
					<svg
						aria-hidden="true"
						viewBox="0 0 24 24"
						className="size-5"
						fill="none"
						stroke="currentColor"
						strokeWidth="2"
					>
						{open === "menu" ? <path d="M6 6l12 12M18 6 6 18" /> : <path d="M4 6h16M4 12h16M4 18h16" />}
					</svg>
				</button>
			</div>
			<search
				id="nav-search"
				className={`${open === "search" ? "block" : "hidden"} w-full sm:block sm:w-auto sm:flex-1`}
			>
				{search}
			</search>
			<div id="nav-menu" className={`${open === "menu" ? "grid" : "hidden"} w-full gap-3 sm:contents`}>
				<div className="flex items-center gap-2 sm:order-none">{actions}</div>
				{links}
			</div>
		</div>
	);
}
