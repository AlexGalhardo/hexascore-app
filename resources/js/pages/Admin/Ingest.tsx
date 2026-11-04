import { Head, Link } from "@inertiajs/react";
import { type FormEvent, type ReactElement, type ReactNode, useId, useState } from "react";

import { ActionError } from "../../components/admin/ActionError";
import { Field } from "../../components/admin/Field";
import { runVariant } from "../../components/admin/overview";
import { PageHeader } from "../../components/admin/PageHeader";
import type { IngestStatus, RunRow } from "../../components/admin/types";
import { Badge } from "../../components/admin/ui/badge";
import { Button, buttonVariants } from "../../components/admin/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "../../components/admin/ui/card";
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
import { Textarea } from "../../components/admin/ui/textarea";
import Dialog from "../../components/Dialog";
import { useAction } from "../../lib/adminAction";
import { useT } from "../../lib/useT";

interface Props {
	runs: RunRow[] | null;
	ingest: IngestStatus;
}

function SectionCard({
	title,
	empty,
	isEmpty,
	children,
}: {
	title: string;
	empty: string;
	isEmpty: boolean;
	children: ReactNode;
}): ReactElement {
	return (
		<Card>
			<CardHeader>
				<CardTitle>{title}</CardTitle>
			</CardHeader>
			<CardContent>
				{isEmpty ? <p className="py-3 text-sm text-muted">{empty}</p> : children}
			</CardContent>
		</Card>
	);
}

function RollbackDialog({
	run,
	onClose,
}: {
	run: RunRow | null;
	onClose: () => void;
}): ReactElement {
	const t = useT();
	const action = useAction();
	const submit = async (e: FormEvent<HTMLFormElement>) => {
		e.preventDefault();
		if (!run) return;
		const reason = String(new FormData(e.currentTarget).get("reason") ?? "");
		if (await action.run(`/admin/ingest/runs/${run.id}/rollback`, { reason })) onClose();
	};
	return (
		<Dialog
			open={run !== null}
			onOpenChange={(open) => {
				if (!open) {
					action.clear();
					onClose();
				}
			}}
			title={t("admin.ingest.rollbackTitle", { id: run?.id ?? "" })}
			description={t("admin.ingest.rollbackHint")}
		>
			<form onSubmit={submit} className="mt-4 grid gap-4">
				<Field label={t("admin.ingest.reasonField")}>
					{(id) => <Input id={id} name="reason" maxLength={500} required />}
				</Field>
				<ActionError message={action.error} />
				<div className="flex justify-end gap-2">
					<Button variant="outline" onClick={onClose}>
						{t("admin.common.cancel")}
					</Button>
					<Button type="submit" variant="destructive" disabled={action.pending}>
						{t("admin.ingest.rollbackConfirm")}
					</Button>
				</div>
			</form>
		</Dialog>
	);
}

function RunsCard({ runs, canRun }: { runs: RunRow[]; canRun: boolean }): ReactElement {
	const t = useT();
	const cancel = useAction();
	const [rollback, setRollback] = useState<RunRow | null>(null);
	return (
		<SectionCard
			title={t("admin.ingest.lastRuns")}
			empty={t("admin.ingest.noRuns")}
			isEmpty={runs.length === 0}
		>
			<div className="grid gap-3">
				<ActionError message={cancel.error} />
				<Table>
					<TableHeader>
						<TableRow>
							<TableHead>{t("admin.ingest.run")}</TableHead>
							<TableHead>{t("admin.ingest.job")}</TableHead>
							<TableHead>{t("admin.ingest.source")}</TableHead>
							<TableHead>{t("admin.ingest.statusCol")}</TableHead>
							<TableHead className="text-right">{t("admin.ingest.seen")}</TableHead>
							<TableHead className="text-right">
								{t("admin.ingest.changed")}
							</TableHead>
							<TableHead className="text-right">{t("admin.ingest.failed")}</TableHead>
							<TableHead>{t("admin.ingest.started")}</TableHead>
							<TableHead>{t("admin.ingest.finished")}</TableHead>
							<TableHead>{t("admin.ingest.error")}</TableHead>
							<TableHead className="text-right">
								{t("admin.common.actions")}
							</TableHead>
						</TableRow>
					</TableHeader>
					<TableBody>
						{runs.map((r) => (
							<TableRow key={r.id}>
								<TableCell className="tabular-nums">{r.id}</TableCell>
								<TableCell className="font-medium">{r.job}</TableCell>
								<TableCell className="text-muted">{r.source || "–"}</TableCell>
								<TableCell>
									<Badge variant={runVariant(r.status)}>
										{t(`admin.dashboard.runs.statuses.${r.status}`)}
									</Badge>
								</TableCell>
								<TableCell className="text-right tabular-nums">{r.seen}</TableCell>
								<TableCell className="text-right tabular-nums">
									{r.changed}
								</TableCell>
								<TableCell className="text-right tabular-nums">
									{r.failed}
								</TableCell>
								<TableCell className="whitespace-nowrap text-muted">
									{r.started || "–"}
								</TableCell>
								<TableCell className="whitespace-nowrap text-muted">
									{r.finished || "–"}
								</TableCell>
								<TableCell className="max-w-64 truncate text-muted" title={r.error}>
									{r.error || "–"}
								</TableCell>
								<TableCell className="text-right">
									<div className="flex justify-end gap-2">
										{canRun &&
											(r.status === "queued" || r.status === "running") && (
												<Button
													size="sm"
													variant="outline"
													disabled={cancel.pending}
													onClick={() =>
														void cancel.run(
															`/admin/ingest/runs/${r.id}/cancel`,
															{},
														)
													}
												>
													{t("admin.ingest.cancelRun")}
												</Button>
											)}
										{r.rollbackable && (
											<Button
												size="sm"
												variant="outline"
												onClick={() => setRollback(r)}
											>
												{t("admin.ingest.rollback")}
											</Button>
										)}
									</div>
								</TableCell>
							</TableRow>
						))}
					</TableBody>
				</Table>
			</div>
			<RollbackDialog run={rollback} onClose={() => setRollback(null)} />
		</SectionCard>
	);
}

function StartRunCard({ ingest }: { ingest: IngestStatus }): ReactElement {
	const t = useT();
	const action = useAction();
	const [formKey, setFormKey] = useState(0);
	const dropId = useId();
	const stages = ingest.stages ?? [];
	const modes = ingest.modes ?? [];
	const submit = async (e: FormEvent<HTMLFormElement>) => {
		e.preventDefault();
		const data = Object.fromEntries(new FormData(e.currentTarget)) as Record<string, string>;
		if (await action.run("/admin/ingest/runs", data)) setFormKey((k) => k + 1);
	};
	return (
		<Card>
			<CardHeader>
				<CardTitle>{t("admin.ingest.startRun")}</CardTitle>
				<CardDescription>{t("admin.ingest.startHint")}</CardDescription>
			</CardHeader>
			<CardContent>
				<form key={formKey} onSubmit={submit} className="grid gap-4 md:grid-cols-2">
					<Field label={t("admin.ingest.jobField")}>
						{(id) => (
							<Select id={id} name="job" required>
								{stages.map((s) => (
									<option key={`${s.job}-${s.source}`} value={s.job}>
										{s.source ? `${s.job} (${s.source})` : s.job}
									</option>
								))}
							</Select>
						)}
					</Field>
					<Field
						label={t("admin.ingest.sourceField")}
						hint={t("admin.ingest.sourceHint")}
					>
						{(id) => <Input id={id} name="source" maxLength={60} />}
					</Field>
					<Field label={t("admin.ingest.mode")}>
						{(id) => (
							<Select id={id} name="mode">
								{modes.map((m) => (
									<option key={m} value={m}>
										{m}
									</option>
								))}
							</Select>
						)}
					</Field>
					<Field label={t("admin.ingest.since")}>
						{(id) => (
							<Input
								id={id}
								name="since"
								maxLength={40}
								placeholder="2026-11-03T00:00:00Z"
							/>
						)}
					</Field>
					<Field label={t("admin.ingest.limit")}>
						{(id) => <Input id={id} name="limit" type="number" min={1} />}
					</Field>
					<Field label={t("admin.ingest.reasonField")}>
						{(id) => <Input id={id} name="reason" maxLength={500} />}
					</Field>
					<Field
						label={t("admin.ingest.qids")}
						hint={t("admin.ingest.qidsHint")}
						className="md:col-span-2"
					>
						{(id) => <Textarea id={id} name="qids" rows={3} />}
					</Field>
					<div className="flex items-center gap-2 text-sm md:col-span-2">
						<Checkbox id={dropId} name="allow_count_drop" />
						<label htmlFor={dropId}>{t("admin.ingest.allowDrop")}</label>
					</div>
					<div className="grid gap-3 md:col-span-2">
						<ActionError message={action.error} />
						<div>
							<Button type="submit" disabled={action.pending}>
								{t("admin.ingest.start")}
							</Button>
						</div>
					</div>
				</form>
			</CardContent>
		</Card>
	);
}

export default function Ingest({ runs, ingest }: Props): ReactElement {
	const t = useT();
	const failures = ingest.failures ?? [];
	const asserts = ingest.asserts ?? [];
	const coverage = ingest.coverage ?? [];
	const link = buttonVariants({ variant: "outline", size: "sm" });
	return (
		<>
			<Head title={t("admin.ingest.title")} />
			<PageHeader
				title={t("admin.ingest.title")}
				description={t("admin.ingest.description")}
				crumbs={[
					{ label: t("admin.title"), href: "/admin/overview" },
					{ label: t("admin.ingest.title") },
				]}
				actions={
					<>
						<Link href="/admin/ingestion" className={link}>
							{t("admin.ingest.sources")}
						</Link>
						<Link href="/admin/ingest/reviews" className={link}>
							{t("admin.ingest.matchReviews")}
						</Link>
						<Link href="/admin/ingest/field-reviews" className={link}>
							{t("admin.ingest.fieldReviews")}
						</Link>
						<Link href="/admin/ingest/conflicts" className={link}>
							{t("admin.ingest.conflicts")}
						</Link>
					</>
				}
			/>
			<div className="grid gap-4">
				<RunsCard runs={runs ?? []} canRun={ingest.canRun} />
				<div className="grid gap-4 lg:grid-cols-2">
					<SectionCard
						title={t("admin.ingest.asserts")}
						empty={t("admin.ingest.noAsserts")}
						isEmpty={asserts.length === 0}
					>
						<Table>
							<TableHeader>
								<TableRow>
									<TableHead>{t("admin.ingest.code")}</TableHead>
									<TableHead>{t("admin.ingest.severity")}</TableHead>
									<TableHead>{t("admin.ingest.observed")}</TableHead>
								</TableRow>
							</TableHeader>
							<TableBody>
								{asserts.map((a) => (
									<TableRow key={a.code}>
										<TableCell className="font-medium">{a.code}</TableCell>
										<TableCell>{a.severity}</TableCell>
										<TableCell className="text-muted">{a.observed}</TableCell>
									</TableRow>
								))}
							</TableBody>
						</Table>
					</SectionCard>
					<SectionCard
						title={t("admin.ingest.deadLetter")}
						empty={t("admin.ingest.noFailures")}
						isEmpty={failures.length === 0}
					>
						<Table>
							<TableHeader>
								<TableRow>
									<TableHead>{t("admin.ingest.source")}</TableHead>
									<TableHead className="text-right">
										{t("admin.ingest.targets")}
									</TableHead>
									<TableHead className="text-right">
										{t("admin.ingest.parked")}
									</TableHead>
								</TableRow>
							</TableHeader>
							<TableBody>
								{failures.map((f) => (
									<TableRow key={f.source}>
										<TableCell className="font-medium">{f.source}</TableCell>
										<TableCell className="text-right tabular-nums">
											{f.total}
										</TableCell>
										<TableCell className="text-right tabular-nums">
											{f.parked}
										</TableCell>
									</TableRow>
								))}
							</TableBody>
						</Table>
					</SectionCard>
				</div>
				<Card>
					<CardHeader>
						<CardTitle>{t("admin.ingest.oldestTmdb")}</CardTitle>
					</CardHeader>
					<CardContent className="text-sm">
						{ingest.oldestTMDB ? (
							ingest.oldestTMDB
						) : (
							<span className="text-muted">{t("admin.ingest.noTmdb")}</span>
						)}
					</CardContent>
				</Card>
				<SectionCard
					title={t("admin.ingest.coverage")}
					empty={t("admin.ingest.noCoverage")}
					isEmpty={coverage.length === 0}
				>
					<Table>
						<TableHeader>
							<TableRow>
								<TableHead>{t("admin.ingest.day")}</TableHead>
								<TableHead>{t("admin.ingest.type")}</TableHead>
								<TableHead className="text-right">
									{t("admin.ingest.sampled")}
								</TableHead>
								<TableHead className="text-right">
									{t("admin.ingest.missing")}
								</TableHead>
							</TableRow>
						</TableHeader>
						<TableBody>
							{coverage.map((c) => (
								<TableRow key={`${c.day}-${c.type}`}>
									<TableCell>{c.day}</TableCell>
									<TableCell>{c.type}</TableCell>
									<TableCell className="text-right tabular-nums">
										{c.sampled}
									</TableCell>
									<TableCell className="text-right tabular-nums">
										{c.missing}
									</TableCell>
								</TableRow>
							))}
						</TableBody>
					</Table>
				</SectionCard>
				{ingest.canRun && <StartRunCard ingest={ingest} />}
			</div>
		</>
	);
}
