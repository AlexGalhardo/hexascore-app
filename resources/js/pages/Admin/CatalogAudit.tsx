import { Head, Link, router } from "@inertiajs/react";
import { type FormEvent, type ReactElement, useState } from "react";

import { ActionForm } from "../../components/admin/ActionForm";
import { Field } from "../../components/admin/Field";
import { PageHeader } from "../../components/admin/PageHeader";
import { Badge } from "../../components/admin/ui/badge";
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
import Dialog from "../../components/Dialog";
import { useT } from "../../lib/useT";

interface AuditRow {
	id: number;
	occurredAt: string;
	actorId: number;
	actorRole: string;
	source: string;
	suggestionId: string;
	targetType: string;
	targetId: string;
	field: string;
	before: string;
	after: string;
	reason: string;
	revertsAuditId: number;
	reverted: boolean;
	revertible: boolean;
}

interface Props {
	rows: AuditRow[] | null;
	filter: { targetType: string; targetID: string; field: string; source: string };
	nextUrl: string;
	sources: string[] | null;
	targets: string[] | null;
	canRevert: boolean;
	revertReasons: string[] | null;
}

const MODES = ["single", "cascade", "force"] as const;

function RevertDialog({
	row,
	reasons,
	onClose,
}: {
	row: AuditRow | null;
	reasons: string[];
	onClose: () => void;
}): ReactElement {
	const t = useT();
	return (
		<Dialog
			open={row !== null}
			onOpenChange={(open) => !open && onClose()}
			title={t("admin.audit.revertTitle", { id: row?.id ?? "" })}
		>
			{row && (
				<ActionForm
					action={`/admin/catalog-audit/${row.id}/revert`}
					submit={t("admin.audit.revert")}
					className="mt-4"
				>
					<Field label={t("admin.audit.mode")}>
						{(id) => (
							<Select id={id} name="mode">
								{MODES.map((m) => (
									<option key={m} value={m}>
										{t(`admin.audit.modes.${m}`)}
									</option>
								))}
							</Select>
						)}
					</Field>
					<Field label={t("admin.audit.reason")}>
						{(id) => (
							<Select id={id} name="reason" required defaultValue="">
								<option value="">{t("admin.common.choose")}</option>
								{reasons.map((r) => (
									<option key={r} value={r}>
										{r}
									</option>
								))}
							</Select>
						)}
					</Field>
				</ActionForm>
			)}
		</Dialog>
	);
}

export default function CatalogAudit({
	rows,
	filter,
	nextUrl,
	sources,
	targets,
	canRevert,
	revertReasons,
}: Props): ReactElement {
	const t = useT();
	const list = rows ?? [];
	const [reverting, setReverting] = useState<AuditRow | null>(null);
	const query = new URLSearchParams({
		target: filter.targetType,
		target_id: filter.targetID,
		field: filter.field,
		source: filter.source,
	}).toString();
	const link = buttonVariants({ variant: "outline", size: "sm" });
	const apply = (e: FormEvent<HTMLFormElement>) => {
		e.preventDefault();
		const data = Object.fromEntries(new FormData(e.currentTarget)) as Record<string, string>;
		const params = Object.fromEntries(Object.entries(data).filter(([, v]) => v !== ""));
		router.get("/admin/catalog-audit", params, { preserveScroll: true });
	};
	return (
		<>
			<Head title={t("admin.audit.title")} />
			<PageHeader
				title={t("admin.audit.title")}
				description={t("admin.audit.description")}
				crumbs={[
					{ label: t("admin.title"), href: "/admin/overview" },
					{ label: t("admin.audit.title") },
				]}
				actions={
					<>
						<a href={`/admin/catalog-audit/export.csv?${query}`} className={link}>
							{t("admin.audit.exportCsv")}
						</a>
						<a href={`/admin/catalog-audit/export.ndjson?${query}`} className={link}>
							{t("admin.audit.exportNdjson")}
						</a>
						<Link href="/admin/locks" className={link}>
							{t("admin.audit.locks")}
						</Link>
					</>
				}
			/>
			<div className="grid gap-4">
				<Card>
					<CardContent>
						<form
							onSubmit={apply}
							className="grid gap-3 sm:grid-cols-2 lg:grid-cols-5 lg:items-end"
							aria-label={t("admin.common.filter")}
						>
							<Field label={t("admin.audit.target")}>
								{(id) => (
									<Select id={id} name="target" defaultValue={filter.targetType}>
										<option value="">{t("admin.common.any")}</option>
										{(targets ?? []).map((x) => (
											<option key={x}>{x}</option>
										))}
									</Select>
								)}
							</Field>
							<Field label={t("admin.audit.targetId")}>
								{(id) => (
									<Input
										id={id}
										name="target_id"
										defaultValue={filter.targetID}
									/>
								)}
							</Field>
							<Field label={t("admin.audit.field")}>
								{(id) => <Input id={id} name="field" defaultValue={filter.field} />}
							</Field>
							<Field label={t("admin.audit.source")}>
								{(id) => (
									<Select id={id} name="source" defaultValue={filter.source}>
										<option value="">{t("admin.common.any")}</option>
										{(sources ?? []).map((x) => (
											<option key={x}>{x}</option>
										))}
									</Select>
								)}
							</Field>
							<div>
								<Button type="submit">{t("admin.common.filter")}</Button>
							</div>
						</form>
					</CardContent>
				</Card>
				<Card>
					<CardContent>
						{list.length === 0 ? (
							<p className="py-3 text-sm text-muted">{t("admin.audit.empty")}</p>
						) : (
							<Table>
								<TableHeader>
									<TableRow>
										<TableHead>{t("admin.audit.id")}</TableHead>
										<TableHead>{t("admin.audit.when")}</TableHead>
										<TableHead>{t("admin.audit.actor")}</TableHead>
										<TableHead>{t("admin.audit.source")}</TableHead>
										<TableHead>{t("admin.audit.target_col")}</TableHead>
										<TableHead>{t("admin.audit.field")}</TableHead>
										<TableHead>{t("admin.audit.before")}</TableHead>
										<TableHead>{t("admin.audit.after")}</TableHead>
										<TableHead className="text-right">
											{t("admin.audit.revert")}
										</TableHead>
									</TableRow>
								</TableHeader>
								<TableBody>
									{list.map((r) => (
										<TableRow key={r.id}>
											<TableCell className="tabular-nums">{r.id}</TableCell>
											<TableCell className="whitespace-nowrap text-muted">
												{r.occurredAt.slice(0, 16).replace("T", " ")}
											</TableCell>
											<TableCell>
												{r.actorRole}
												{r.actorId ? ` #${r.actorId}` : ""}
											</TableCell>
											<TableCell>
												{r.source}
												{r.revertsAuditId
													? ` ${t("admin.audit.ofRevert", { id: r.revertsAuditId })}`
													: ""}
												{r.reason ? ` (${r.reason})` : ""}
											</TableCell>
											<TableCell>
												{r.targetType}{" "}
												<span className="text-muted">{r.targetId}</span>
												{r.suggestionId && (
													<>
														{" · "}
														<Link
															href={`/admin/moderation/${r.suggestionId}`}
															className="text-primary hover:underline"
														>
															{t("admin.audit.suggestion")}
														</Link>
													</>
												)}
											</TableCell>
											<TableCell>{r.field}</TableCell>
											<TableCell
												className="max-w-48 truncate"
												title={r.before}
											>
												<code>{r.before}</code>
											</TableCell>
											<TableCell
												className="max-w-48 truncate"
												title={r.after}
											>
												<code>{r.after}</code>
											</TableCell>
											<TableCell className="text-right">
												{r.reverted ? (
													<Badge variant="secondary">
														{t("admin.audit.reverted")}
													</Badge>
												) : canRevert && r.revertible ? (
													<Button
														size="sm"
														variant="outline"
														onClick={() => setReverting(r)}
													>
														{t("admin.audit.revert")}
													</Button>
												) : (
													"–"
												)}
											</TableCell>
										</TableRow>
									))}
								</TableBody>
							</Table>
						)}
						{nextUrl && (
							<p className="pt-3 text-sm">
								<Link
									href={nextUrl}
									rel="next"
									className="text-primary hover:underline"
								>
									{t("admin.common.olderEntries")}
								</Link>
							</p>
						)}
					</CardContent>
				</Card>
			</div>
			<RevertDialog
				row={reverting}
				reasons={revertReasons ?? []}
				onClose={() => setReverting(null)}
			/>
		</>
	);
}
