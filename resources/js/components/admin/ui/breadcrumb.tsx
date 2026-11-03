import { Link } from "@inertiajs/react";
import { ChevronRight } from "lucide-react";
import type { ReactElement } from "react";

import { cn } from "../../../lib/cn";

export interface Crumb {
	label: string;
	href?: string;
}

/** The current page is the last crumb and carries aria-current. */
export function Breadcrumb({
	items,
	className,
}: {
	items: Crumb[];
	className?: string;
}): ReactElement {
	return (
		<nav aria-label="Breadcrumb" className={className}>
			<ol className="flex items-center gap-1.5 text-sm text-muted">
				{items.map((item, i) => {
					const last = i === items.length - 1;
					return (
						<li key={item.href ?? item.label} className="flex items-center gap-1.5">
							{item.href && !last ? (
								<Link href={item.href} className="hover:text-fg">
									{item.label}
								</Link>
							) : (
								<span
									aria-current={last ? "page" : undefined}
									className={cn(last && "font-medium text-fg")}
								>
									{item.label}
								</span>
							)}
							{!last && <ChevronRight aria-hidden="true" className="size-3.5" />}
						</li>
					);
				})}
			</ol>
		</nav>
	);
}
