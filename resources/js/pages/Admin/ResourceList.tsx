import { Head, Link, router, usePage } from "@inertiajs/react";
import { ArrowDown, ArrowUp } from "lucide-react";
import type { FormEvent, ReactElement } from "react";

import { ActionButton } from "../../components/admin/ActionForm";
import { Field } from "../../components/admin/Field";
import { PageHeader } from "../../components/admin/PageHeader";
import { CellValue, colLabel, type ResColumn } from "../../components/admin/resources";
import { Button, buttonVariants } from "../../components/admin/ui/button";
import { Card, CardContent } from "../../components/admin/ui/card";
import { Input } from "../../components/admin/ui/input";
import { Select } from "../../components/admin/ui/select";
import {
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
} from "../../components/admin/ui/table";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";

interface Props {
	resource: { slug: string; navKey: string; hasDetail: boolean; searchable: boolean };
	columns: ResColumn[];
	rows: { id: string; cells: unknown[] }[];
	filters: { key: string; kind: "enum" | "bool"; options: string[] | null; value: string }[];
	page: number;
	hasNext: boolean;
	q: string;
	sort: string;
	dir: string;
}

/** The cell of the column called key. */
function cellOf(r: Props["rows"][number], columns: ResColumn[], key: string): unknown {
	return r.cells[columns.findIndex((c) => c.key === key)];
}

/** The record id, or the first two cell values for the lists without a detail page (composite keys). */
function rowKey(r: Props["rows"][number]): string {
	return r.id || r.cells.slice(0, 2).map(String).join("|");
}

export default function ResourceList({
	resource,
	columns,
	rows,
	filters,
	page,
	hasNext,
	q,
	sort,
	dir,
}: Props): ReactElement {
	const t = useT();
	const { locale } = usePage<SharedProps>().props;
	const base = `/admin/r/${resource.slug}`;
	const title = t(`admin.nav.${resource.navKey}`);
	const current = () => {
		const params: Record<string, string> = {};
		if (q) params.q = q;
		if (sort) {
			params.sort = sort;
			params.dir = dir;
		}
		for (const f of filters) if (f.value) params[`f_${f.key}`] = f.value;
		return params;
	};
	const go = (extra: Record<string, string | number>) =>
		router.get(base, { ...current(), ...extra });
	const apply = (e: FormEvent<HTMLFormElement>) => {
		e.preventDefault();
		const data = Object.fromEntries(new FormData(e.currentTarget)) as Record<string, string>;
		router.get(base, Object.fromEntries(Object.entries(data).filter(([, v]) => v !== "")));
	};
	const link = buttonVariants({ variant: "outline", size: "sm" });
	const query = (extra: Record<string, string | number>) =>
		new URLSearchParams({ ...current(), ...extra } as Record<string, string>).toString();
	const first = columns[0]?.key;
	return (
		<>
			<Head title={title} />
			<PageHeader
				title={title}
				crumbs={[{ label: t("admin.title"), href: "/admin/overview" }, { label: title }]}
			/>
			<div className="grid gap-4">
				{(resource.searchable || filters.length > 0) && (
					<Card>
						<CardContent>
							<form
								onSubmit={apply}
								aria-label={t("admin.res.filters")}
								className="flex flex-wrap items-end gap-3"
							>
								{resource.searchable && (
									<Field
										label={t("admin.res.search")}
										className="min-w-56 flex-1"
									>
										{(id) => (
											<Input
												id={id}
												name="q"
												defaultValue={q}
												placeholder={t("admin.res.searchPlaceholder")}
											/>
										)}
									</Field>
								)}
								{filters.map((f) => (
									<Field key={f.key} label={colLabel(t, f.key)}>
										{(id) => (
											<Select
												id={id}
												name={`f_${f.key}`}
												defaultValue={f.value}
											>
												<option value="">{t("admin.common.any")}</option>
												{f.kind === "bool" ? (
													<>
														<option value="true">
															{t("admin.common.yes")}
														</option>
														<option value="false">
															{t("admin.common.no")}
														</option>
													</>
												) : (
													(f.options ?? []).map((o) => (
														<option key={o} value={o}>
															{o}
														</option>
													))
												)}
											</Select>
										)}
									</Field>
								))}
								<Button type="submit">{t("admin.common.filter")}</Button>
								{(q || filters.some((f) => f.value)) && (
									<Link href={base} className={link}>
										{t("admin.res.clear")}
									</Link>
								)}
							</form>
						</CardContent>
					</Card>
				)}
				<Card>
					<CardContent>
						{rows.length === 0 ? (
							<p className="py-3 text-sm text-muted">{t("admin.res.noRows")}</p>
						) : (
							<Table>
								<TableHeader>
									<TableRow>
										{columns.map((c) => (
											<TableHead
												key={c.key}
												aria-sort={
													sort === c.key
														? dir === "asc"
															? "ascending"
															: "descending"
														: undefined
												}
											>
												{c.sortable ? (
													<button
														type="button"
														className="inline-flex items-center gap-1 hover:text-fg"
														onClick={() =>
															go({
																sort: c.key,
																dir:
																	sort === c.key && dir === "desc"
																		? "asc"
																		: "desc",
																page: 1,
															})
														}
													>
														{colLabel(t, c.key)}
														{sort === c.key &&
															(dir === "asc" ? (
																<ArrowUp
																	aria-hidden="true"
																	className="size-3"
																/>
															) : (
																<ArrowDown
																	aria-hidden="true"
																	className="size-3"
																/>
															))}
													</button>
												) : (
													colLabel(t, c.key)
												)}
											</TableHead>
										))}
										{resource.slug === "rating-flags" && (
											<TableHead className="text-right">
												{t("admin.common.actions")}
											</TableHead>
										)}
									</TableRow>
								</TableHeader>
								<TableBody>
									{rows.map((r) => (
										<TableRow key={rowKey(r)}>
											{columns.map((c, ci) => {
												const cell = (
													<CellValue
														kind={c.kind}
														value={r.cells[ci]}
														locale={locale}
														t={t}
													/>
												);
												return (
													<TableCell
														key={c.key}
														className="max-w-72 truncate"
													>
														{resource.hasDetail && c.key === first ? (
															<Link
																href={`${base}/${r.id}`}
																className="font-medium text-primary hover:underline"
															>
																{cell}
															</Link>
														) : (
															cell
														)}
													</TableCell>
												);
											})}
											{resource.slug === "rating-flags" && (
												<TableCell className="text-right">
													<ActionButton
														action={`/admin/scoring/flags/${String(cellOf(r, columns, "user_id"))}/${String(cellOf(r, columns, "item_id"))}/restore`}
													>
														{t("admin.res.actions.restore")}
													</ActionButton>
												</TableCell>
											)}
										</TableRow>
									))}
								</TableBody>
							</Table>
						)}
						{(page > 1 || hasNext) && (
							<nav
								aria-label={t("admin.res.page", { n: page })}
								className="flex items-center justify-between pt-4 text-sm"
							>
								{page > 1 ? (
									<Link
										href={`${base}?${query({ page: page - 1 })}`}
										rel="prev"
										className={link}
									>
										{t("admin.res.previous")}
									</Link>
								) : (
									<span />
								)}
								<span className="text-muted">
									{t("admin.res.page", { n: page })}
								</span>
								{hasNext ? (
									<Link
										href={`${base}?${query({ page: page + 1 })}`}
										rel="next"
										className={link}
									>
										{t("admin.res.next")}
									</Link>
								) : (
									<span />
								)}
							</nav>
						)}
					</CardContent>
				</Card>
			</div>
		</>
	);
}
