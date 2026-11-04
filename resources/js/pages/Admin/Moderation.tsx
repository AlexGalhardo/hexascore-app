import { Head, Link, router } from "@inertiajs/react";
import type { FormEvent, ReactElement } from "react";

import { ActionButton } from "../../components/admin/ActionForm";
import { Field } from "../../components/admin/Field";
import { PageHeader } from "../../components/admin/PageHeader";
import { Badge } from "../../components/admin/ui/badge";
import { Button, buttonVariants } from "../../components/admin/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "../../components/admin/ui/card";
import { Checkbox } from "../../components/admin/ui/checkbox";
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

interface Row {
	id: string;
	kind: string;
	targetType: string;
	targetID: string;
	domain: string;
	locale: string;
	status: string;
	fields: string[] | null;
	authorID: number;
	assignedTo: number;
	authorTier: string;
	authorAccuracy: number;
	version: number;
	age: string;
	competing: boolean;
	stale: boolean;
	unreachable: boolean;
	recApprove: number;
	recReject: number;
}

interface Props {
	queue: {
		rows: Row[] | null;
		filter: {
			status: string;
			kind: string;
			domain: string;
			age: string;
			field: string;
			competing: boolean;
			stale: boolean;
		};
		nextUrl: string;
		statuses: string[] | null;
		kinds: string[] | null;
		domains: string[] | null;
		ageBuckets: string[] | null;
	};
	canApply: boolean;
	rejectReasons: string[] | null;
}

function Options({ values, any }: { values: string[] | null; any: string }): ReactElement {
	return (
		<>
			<option value="">{any}</option>
			{(values ?? []).map((v) => (
				<option key={v} value={v}>
					{v}
				</option>
			))}
		</>
	);
}

export default function Moderation({ queue, canApply, rejectReasons }: Props): ReactElement {
	const t = useT();
	const rows = queue.rows ?? [];
	const f = queue.filter;
	const link = buttonVariants({ variant: "outline", size: "sm" });
	const apply = (e: FormEvent<HTMLFormElement>) => {
		e.preventDefault();
		const data = Object.fromEntries(new FormData(e.currentTarget)) as Record<string, string>;
		router.get(
			"/admin/moderation",
			Object.fromEntries(Object.entries(data).filter(([, v]) => v !== "")),
		);
	};
	return (
		<>
			<Head title={t("admin.moderation.title")} />
			<PageHeader
				title={t("admin.moderation.title")}
				crumbs={[
					{ label: t("admin.title"), href: "/admin/overview" },
					{ label: t("admin.moderation.title") },
				]}
				actions={
					<Link href="/admin/moderation/metrics" className={link}>
						{t("admin.moderation.metrics")}
					</Link>
				}
			/>
			<div className="grid gap-4">
				<Card>
					<CardContent>
						<form
							onSubmit={apply}
							aria-label={t("admin.common.filter")}
							className="grid gap-3 sm:grid-cols-2 lg:grid-cols-6 lg:items-end"
						>
							<Field label={t("admin.moderation.statusField")}>
								{(id) => (
									<Select id={id} name="status" defaultValue={f.status}>
										<Options
											values={queue.statuses}
											any={t("admin.moderation.open")}
										/>
									</Select>
								)}
							</Field>
							<Field label={t("admin.moderation.kind")}>
								{(id) => (
									<Select id={id} name="kind" defaultValue={f.kind}>
										<Options values={queue.kinds} any={t("admin.common.any")} />
									</Select>
								)}
							</Field>
							<Field label={t("admin.moderation.domain")}>
								{(id) => (
									<Select id={id} name="domain" defaultValue={f.domain}>
										<Options
											values={queue.domains}
											any={t("admin.common.any")}
										/>
									</Select>
								)}
							</Field>
							<Field label={t("admin.moderation.age")}>
								{(id) => (
									<Select id={id} name="age" defaultValue={f.age}>
										<Options
											values={queue.ageBuckets}
											any={t("admin.common.any")}
										/>
									</Select>
								)}
							</Field>
							<Field label={t("admin.moderation.field")}>
								{(id) => <Input id={id} name="field" defaultValue={f.field} />}
							</Field>
							<div className="flex flex-wrap items-center gap-3 text-sm lg:col-span-6">
								<span className="flex items-center gap-2">
									<Checkbox
										id="q-competing"
										name="competing"
										value="1"
										defaultChecked={f.competing}
									/>
									<label htmlFor="q-competing">
										{t("admin.moderation.competing")}
									</label>
								</span>
								<span className="flex items-center gap-2">
									<Checkbox
										id="q-stale"
										name="stale"
										value="1"
										defaultChecked={f.stale}
									/>
									<label htmlFor="q-stale">{t("admin.moderation.stale")}</label>
								</span>
								<Button type="submit">{t("admin.common.filter")}</Button>
							</div>
						</form>
					</CardContent>
				</Card>
				<Card>
					<CardContent>
						{rows.length === 0 ? (
							<p className="py-3 text-sm text-muted">{t("admin.moderation.empty")}</p>
						) : (
							<Table>
								<TableHeader>
									<TableRow>
										<TableHead>
											<span className="sr-only">
												{t("admin.common.actions")}
											</span>
										</TableHead>
										<TableHead>{t("admin.moderation.age")}</TableHead>
										<TableHead>{t("admin.moderation.kind")}</TableHead>
										<TableHead>{t("admin.moderation.target")}</TableHead>
										<TableHead>{t("admin.moderation.fields")}</TableHead>
										<TableHead>{t("admin.moderation.author")}</TableHead>
										<TableHead>{t("admin.moderation.badges")}</TableHead>
										<TableHead>
											{t("admin.moderation.recommendations")}
										</TableHead>
										<TableHead>{t("admin.moderation.assignee")}</TableHead>
										<TableHead className="text-right">
											{t("admin.common.actions")}
										</TableHead>
									</TableRow>
								</TableHeader>
								<TableBody>
									{rows.map((r) => {
										const inReview = r.status === "in_review";
										return (
											<TableRow key={r.id}>
												<TableCell>
													<Checkbox
														name="ids"
														value={r.id}
														form="bulk"
														aria-label={t("admin.moderation.select", {
															id: r.id,
														})}
													/>
												</TableCell>
												<TableCell>
													<Badge variant="secondary">{r.age}</Badge>
												</TableCell>
												<TableCell>{r.kind}</TableCell>
												<TableCell>
													<Link
														href={`/admin/moderation/${r.id}`}
														className="text-primary hover:underline"
													>
														{r.targetType} · {r.domain} · {r.locale}
													</Link>
												</TableCell>
												<TableCell className="text-muted">
													{(r.fields ?? []).join(", ")}
												</TableCell>
												<TableCell>
													{r.authorID
														? `#${r.authorID} ${r.authorTier} ${r.authorAccuracy.toFixed(2)}`
														: "–"}
												</TableCell>
												<TableCell>
													<span className="flex flex-wrap gap-1">
														{r.competing && (
															<Badge variant="warning">
																{t("admin.suggestion.competing")}
															</Badge>
														)}
														{r.stale && (
															<Badge variant="warning">
																{t("admin.suggestion.stale")}
															</Badge>
														)}
														{r.unreachable && (
															<Badge variant="destructive">
																{t("admin.moderation.unreachable")}
															</Badge>
														)}
													</span>
												</TableCell>
												<TableCell className="text-muted">
													{r.recApprove || r.recReject
														? t("admin.moderation.recs", {
																approve: r.recApprove,
																reject: r.recReject,
															})
														: "–"}
												</TableCell>
												<TableCell className="text-muted">
													{r.assignedTo ? `#${r.assignedTo}` : "–"}
												</TableCell>
												<TableCell className="text-right">
													<ActionButton
														action={`/admin/moderation/${r.id}/${inReview ? "release" : "claim"}`}
														fields={{ version: r.version }}
													>
														{inReview
															? t("admin.moderation.release")
															: t("admin.moderation.claim")}
													</ActionButton>
												</TableCell>
											</TableRow>
										);
									})}
								</TableBody>
							</Table>
						)}
						{queue.nextUrl && (
							<p className="pt-3 text-sm">
								<Link
									href={queue.nextUrl}
									rel="next"
									className="text-primary hover:underline"
								>
									{t("admin.common.nextPage")}
								</Link>
							</p>
						)}
					</CardContent>
				</Card>
				{rows.length > 0 && (
					<Card>
						<CardHeader>
							<CardTitle>{t("admin.moderation.bulkApply")}</CardTitle>
						</CardHeader>
						<CardContent>
							{/* A plain post: the page that answers lists the outcome of each selected suggestion. */}
							<form
								id="bulk"
								method="post"
								action="/admin/moderation/bulk"
								className="flex flex-wrap items-end gap-3"
							>
								<Field label={t("admin.moderation.bulkAction")}>
									{(id) => (
										<Select id={id} name="action">
											<option value="claim">
												{t("admin.moderation.claim")}
											</option>
											<option value="release">
												{t("admin.moderation.release")}
											</option>
											{canApply && (
												<option value="approve">
													{t("admin.moderation.approve")}
												</option>
											)}
											{canApply && (
												<option value="reject">
													{t("admin.moderation.reject")}
												</option>
											)}
										</Select>
									)}
								</Field>
								{canApply && (
									<Field label={t("admin.moderation.bulkReason")}>
										{(id) => (
											<Select id={id} name="reason">
												<option value="">–</option>
												{(rejectReasons ?? []).map((r) => (
													<option key={r} value={r}>
														{r}
													</option>
												))}
											</Select>
										)}
									</Field>
								)}
								<Button type="submit">{t("admin.moderation.bulkApply")}</Button>
							</form>
						</CardContent>
					</Card>
				)}
			</div>
		</>
	);
}
