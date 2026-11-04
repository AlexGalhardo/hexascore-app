import type { ReactElement, ReactNode } from "react";

import { Breadcrumb, type Crumb } from "./ui/breadcrumb";

/** The top of an admin screen: breadcrumb, title, one line of context and the page actions. */
export function PageHeader({
	title,
	description,
	crumbs,
	actions,
}: {
	title: string;
	description?: string;
	crumbs?: Crumb[];
	actions?: ReactNode;
}): ReactElement {
	return (
		<header className="mb-6 grid gap-3">
			{crumbs && <Breadcrumb items={crumbs} />}
			<div className="flex flex-wrap items-end justify-between gap-3">
				<div className="grid gap-1">
					<h1 className="text-2xl font-semibold tracking-tight">{title}</h1>
					{description && <p className="text-sm text-muted">{description}</p>}
				</div>
				{actions && <div className="flex items-center gap-2">{actions}</div>}
			</div>
		</header>
	);
}
