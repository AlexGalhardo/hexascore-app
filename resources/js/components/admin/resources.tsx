import type { ReactElement } from "react";

import { Badge, type BadgeProps } from "./ui/badge";

export interface ResColumn {
	key: string;
	kind: "text" | "id" | "number" | "bool" | "time" | "date" | "badge" | "json";
	sortable: boolean;
}

type Translate = (key: string, params?: Record<string, string | number>) => string;

/** The translated column label, or the key turned into words when the catalogue has none. */
export function colLabel(t: Translate, key: string): string {
	const k = `admin.res.col.${key}`;
	const v = t(k);
	if (v !== k) return v;
	const words = key.replace(/_/g, " ");
	return words.charAt(0).toUpperCase() + words.slice(1);
}

const STATUS: Record<string, BadgeProps["variant"]> = {
	succeeded: "success",
	resolved: "success",
	partial: "warning",
	failed: "destructive",
	open: "warning",
	banned: "destructive",
	admin: "default",
};

/** One cell value drawn by its kind. Empty values are a dash; numbers and dates use the active locale. */
export function CellValue({
	kind,
	value,
	locale,
	t,
}: {
	kind: ResColumn["kind"];
	value: unknown;
	locale: string;
	t: Translate;
}): ReactElement {
	const empty = <span className="text-muted">–</span>;
	if (value === null || value === undefined || value === "") return empty;
	switch (kind) {
		case "bool":
			return <>{value ? t("admin.common.yes") : t("admin.common.no")}</>;
		case "number": {
			const n = Number(value);
			return (
				<span className="tabular-nums">
					{Number.isNaN(n) ? String(value) : n.toLocaleString(locale)}
				</span>
			);
		}
		case "time": {
			const d = new Date(String(value));
			if (Number.isNaN(d.getTime())) return <>{String(value)}</>;
			return (
				<span className="whitespace-nowrap">
					{new Intl.DateTimeFormat(locale, {
						dateStyle: "short",
						timeStyle: "short",
					}).format(d)}
				</span>
			);
		}
		case "badge":
			return <Badge variant={STATUS[String(value)] ?? "secondary"}>{String(value)}</Badge>;
		case "id":
			return <code className="text-xs text-muted">{String(value)}</code>;
		case "json":
			return (
				<code className="block max-h-64 overflow-auto text-xs break-words whitespace-pre-wrap">
					{String(value)}
				</code>
			);
		default:
			return <>{String(value)}</>;
	}
}
