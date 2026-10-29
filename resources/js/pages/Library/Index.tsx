import { Link, router, usePage } from "@inertiajs/react";
import type { FormEvent } from "react";

import ItemLink, { type ItemRef } from "../../components/library/ItemLink";
import Seo from "../../components/Seo";
import { localized } from "../../lib/route";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";

const STATUSES = ["want", "in_progress", "completed", "dropped"] as const;
const TYPES = ["game", "movie", "series", "book"] as const;
const SORTS = ["updated", "title", "score"] as const;

interface LibraryItem extends ItemRef {
	status: string;
	stars: number | null;
	edition: { id: string; title: string } | null;
}

interface Group {
	type: ItemRef["type"];
	total: number;
	items: LibraryItem[];
	next: number;
}

interface Filters {
	status: string;
	type: string;
	sort: string;
}

interface Props {
	counts: Record<string, Record<string, number>>;
	groups: { data: Group[] };
	filters: Filters;
}

const sum = (counts: Props["counts"], status: string, type: string): number =>
	Object.entries(counts ?? {})
		.filter(([s]) => !status || s === status)
		.reduce(
			(n, [, byType]) =>
				n + (type ? (byType[type] ?? 0) : Object.values(byType).reduce((a, b) => a + b, 0)),
			0,
		);

export default function Index({ counts, groups, filters }: Props) {
	const t = useT();
	const { locale } = usePage<SharedProps>().props;
	const path = localized("/library", locale);

	const pageHref = (type: string, cursor?: number) =>
		`${path}?${new URLSearchParams({ ...(filters.status && { status: filters.status }), type, ...(filters.sort && { sort: filters.sort }), ...(cursor !== undefined && { cursor: String(cursor) }) })}`;

	const clearEdition = (id: string) =>
		router.delete(`/items/${id}/edition`, { preserveScroll: true });

	const apply = (e: FormEvent<HTMLFormElement>) => {
		e.preventDefault();
		const data = new FormData(e.currentTarget);
		const query: Record<string, string> = {};
		for (const key of ["status", "type", "sort"]) {
			const value = String(data.get(key) ?? "");
			if (value) query[key] = value;
		}
		router.get(path, query);
	};

	return (
		<>
			<Seo title={t("library.title")} noindex />
			<h1 className="text-2xl font-semibold">{t("library.title")}</h1>
			<form onSubmit={apply} className="mt-4 flex flex-wrap items-end gap-3">
				<label className="flex flex-col text-sm">
					{t("library.status")}
					<select
						name="status"
						defaultValue={filters.status}
						className="rounded-md border border-border bg-surface px-2 py-1"
					>
						<option value="">{t("library.allStatuses")}</option>
						{STATUSES.map((s) => (
							<option key={s} value={s}>
								{t(`social.status.shelves.${s}`)} ({sum(counts, s, filters.type)})
							</option>
						))}
					</select>
				</label>
				<label className="flex flex-col text-sm">
					{t("library.type")}
					<select
						name="type"
						defaultValue={filters.type}
						className="rounded-md border border-border bg-surface px-2 py-1"
					>
						<option value="">{t("library.allTypes")}</option>
						{TYPES.map((s) => (
							<option key={s} value={s}>
								{t(`itemTypes.${s}`)} ({sum(counts, filters.status, s)})
							</option>
						))}
					</select>
				</label>
				<label className="flex flex-col text-sm">
					{t("library.sort")}
					<select
						name="sort"
						defaultValue={filters.sort || "updated"}
						className="rounded-md border border-border bg-surface px-2 py-1"
					>
						{SORTS.map((s) => (
							<option key={s} value={s}>
								{t(`library.sorts.${s}`)}
							</option>
						))}
					</select>
				</label>
				<button
					type="submit"
					className="rounded-md border border-border px-3 py-1 font-medium"
				>
					{t("library.apply")}
				</button>
			</form>
			{groups.data.length === 0 ? (
				<p className="mt-6">{t("library.empty")}</p>
			) : (
				groups.data.map((g) => (
					<section key={g.type} className="mt-6" aria-labelledby={`group-${g.type}`}>
						<h2 id={`group-${g.type}`} className="text-lg font-semibold">
							{t(`itemTypes.${g.type}`)}{" "}
							<span className="text-sm font-normal text-muted">
								{t("library.showing", { count: g.items.length, total: g.total })}
							</span>
						</h2>
						<ul className="mt-2 divide-y divide-border">
							{g.items.map((item) => (
								<li
									key={item.id}
									className="flex flex-wrap items-center gap-x-4 gap-y-1 py-2"
								>
									<ItemLink item={item} />
									<span className="text-sm text-muted">
										{t(`social.status.${item.type}.${item.status}`)}
									</span>
									{item.stars !== null && (
										<span className="text-sm">
											{t("library.score", { stars: item.stars })}
										</span>
									)}
									{item.edition && (
										<span className="text-sm text-muted">
											{t("library.edition", { title: item.edition.title })}{" "}
											<button
												type="button"
												className="underline"
												onClick={() => clearEdition(item.id)}
											>
												{t("library.clearEdition")}
											</button>
										</span>
									)}
								</li>
							))}
						</ul>
						{g.next >= 0 && (
							<Link
								className="mt-2 inline-block underline"
								href={
									filters.type ? pageHref(filters.type, g.next) : pageHref(g.type)
								}
							>
								{filters.type ? t("lists.next") : t("library.seeAll")}
							</Link>
						)}
					</section>
				))
			)}
		</>
	);
}
