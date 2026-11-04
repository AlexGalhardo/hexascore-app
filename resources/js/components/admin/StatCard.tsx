import type { LucideIcon } from "lucide-react";
import type { ReactElement } from "react";

import { cn } from "../../lib/cn";
import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";

/** One headline number. `tone` colours the number only, never the card, and the label always says what it is. */
export function StatCard({
	label,
	value,
	icon: Icon,
	hint,
	tone = "default",
	href,
}: {
	label: string;
	value: string;
	icon: LucideIcon;
	hint?: string;
	tone?: "default" | "danger";
	href?: string;
}): ReactElement {
	const body = (
		<Card className="h-full transition-colors hover:bg-hover/40">
			<CardHeader className="flex flex-row items-center justify-between">
				<CardTitle className="font-medium text-muted">{label}</CardTitle>
				<Icon aria-hidden="true" className="size-4 text-muted" />
			</CardHeader>
			<CardContent className="grid gap-1">
				<div
					className={cn(
						"text-3xl font-semibold tracking-tight tabular-nums",
						tone === "danger" && "text-danger",
					)}
				>
					{value}
				</div>
				{hint && <p className="text-xs text-muted">{hint}</p>}
			</CardContent>
		</Card>
	);
	return href ? (
		<a
			href={href}
			className="rounded-xl outline-none focus-visible:ring-2 focus-visible:ring-primary"
		>
			{body}
		</a>
	) : (
		body
	);
}
