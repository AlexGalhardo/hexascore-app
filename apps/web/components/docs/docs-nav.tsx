"use client";

import { Link, usePathname } from "@/i18n/navigation";
import { DOC_PAGES, type DocKey } from "./pages";

export function DocsNav({ label, labels }: { label: string; labels: Record<DocKey, string> }) {
	const pathname = usePathname();
	return (
		<nav aria-label={label}>
			<ul className="flex gap-1 overflow-x-auto text-sm lg:flex-col">
				{DOC_PAGES.map((page) => {
					const active = pathname === page.href;
					return (
						<li key={page.href} className="shrink-0">
							<Link
								href={page.href}
								aria-current={active ? "page" : undefined}
								className={`block rounded-md px-3 py-2 hover:bg-surface ${active ? "bg-surface font-semibold" : "text-muted"}`}
							>
								{labels[page.key]}
							</Link>
						</li>
					);
				})}
			</ul>
		</nav>
	);
}
