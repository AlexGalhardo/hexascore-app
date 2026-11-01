import { Link, router, usePage } from "@inertiajs/react";
import type { FormEvent } from "react";

import EditionPicker from "../../components/library/EditionPicker";
import type { ItemRef } from "../../components/library/ItemLink";
import ShelfCard from "../../components/library/ShelfCard";
import Seo from "../../components/Seo";
import { btnClass } from "../../components/ui/Btn";
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

const field =
	"min-h-11 w-full rounded-xs border border-field bg-field-bg px-3 text-[15px] text-fg focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-primary";
const label = "grid gap-1.5 text-[13px] font-semibold sm:w-[220px]";

export default function Index({ counts, groups, filters }: Props) {
	const t = useT();
	const { locale, auth } = usePage<SharedProps>().props;
	const owner = auth?.user?.nickname ?? auth?.user?.name;
	const all = sum(counts, "", "");
	const path = localized("/library", locale);

	const pageHref = (type: string, cursor?: number) =>
		`${path}?${new URLSearchParams({ ...(filters.status && { status: filters.status }), type, ...(filters.sort && { sort: filters.sort }), ...(cursor !== undefined && { cursor: String(cursor) }) })}`;

	const clearEdition = (id: string) =>
		router.delete(localized(`/items/${id}/edition`, locale), { preserveScroll: true });

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
			<header className="mt-6 flex flex-wrap items-end justify-between gap-x-6 gap-y-2 md:mt-10">
				<div>
					{owner && (
						<p className="text-xs font-bold tracking-[0.1em] text-primary uppercase">
							{t("library.shelvesOf", { name: owner })}
						</p>
					)}
					<h1 className="mt-2 font-display text-5xl leading-none font-bold tracking-[-0.02em] md:text-[64px]">
						{t("library.title")}
					</h1>
				</div>
				<p className="text-sm text-muted">{t("library.titleCount", { count: all })}</p>
			</header>
			<form onSubmit={apply} className="mt-6 flex flex-wrap items-end gap-3 md:mt-8">
				<label className={label}>
					{t("library.status")}
					<select name="status" defaultValue={filters.status} className={field}>
						<option value="">
							{t("library.allStatuses")} ({sum(counts, "", filters.type)})
						</option>
						{STATUSES.map((s) => (
							<option key={s} value={s}>
								{t(`social.status.shelves.${s}`)} ({sum(counts, s, filters.type)})
							</option>
						))}
					</select>
				</label>
				<label className={label}>
					{t("library.type")}
					<select name="type" defaultValue={filters.type} className={field}>
						<option value="">{t("library.allTypes")}</option>
						{TYPES.map((s) => (
							<option key={s} value={s}>
								{t(`itemTypes.${s}`)} ({sum(counts, filters.status, s)})
							</option>
						))}
					</select>
				</label>
				<label className={label}>
					{t("library.sort")}
					<select name="sort" defaultValue={filters.sort || "updated"} className={field}>
						{SORTS.map((s) => (
							<option key={s} value={s}>
								{t(`library.sorts.${s}`)}
							</option>
						))}
					</select>
				</label>
				<button type="submit" className={btnClass("primary")}>
					{t("library.apply")}
				</button>
			</form>
			{groups.data.length === 0 ? (
				<p className="mt-8 border-t-2 border-fg pt-6 font-display text-2xl italic">
					{t("library.empty")}
				</p>
			) : (
				<>
					{STATUSES.map((status) => {
						const cards = groups.data.flatMap((g) =>
							g.items.filter((i) => i.status === status),
						);
						if (cards.length === 0) return null;
						return (
							<section
								key={status}
								className="mt-8 border-t-2 border-fg pt-4"
								aria-labelledby={`status-${status}`}
							>
								<h2
									id={`status-${status}`}
									className="flex flex-wrap items-baseline gap-x-3 font-display text-[28px] leading-tight font-bold"
								>
									{t(`social.status.shelves.${status}`)}
									<span className="font-sans text-sm font-normal text-muted">
										{t("library.titleCount", {
											count: sum(counts, status, filters.type),
										})}
									</span>
								</h2>
								<ul className="mt-4 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
									{cards.map((item) => (
										<li key={item.id}>
											<ShelfCard
												item={item}
												stars={item.stars}
												statusLabel={t(
													`social.status.${item.type}.${item.status}`,
												)}
											>
												{item.type === "book" && !item.edition && (
													<EditionPicker itemId={item.id} />
												)}
												{item.edition && (
													<span className="text-sm text-muted">
														{t("library.edition", {
															title: item.edition.title,
														})}{" "}
														<button
															type="button"
															className="font-semibold text-primary underline"
															onClick={() => clearEdition(item.id)}
														>
															{t("library.clearEdition")}
														</button>
													</span>
												)}
											</ShelfCard>
										</li>
									))}
								</ul>
							</section>
						);
					})}
					{groups.data
						.filter((g) => g.next >= 0)
						.map((g) => (
							<p key={g.type} className="mt-6">
								{t(`itemTypes.${g.type}`)}{" "}
								<span className="text-muted">
									{t("library.showing", {
										count: g.items.length,
										total: g.total,
									})}
								</span>{" "}
								<Link
									className="inline-flex min-h-11 items-center font-semibold text-primary hover:underline"
									href={
										filters.type
											? pageHref(filters.type, g.next)
											: pageHref(g.type)
									}
								>
									{filters.type ? t("lists.next") : t("library.seeAll")}
								</Link>
							</p>
						))}
				</>
			)}
		</>
	);
}
