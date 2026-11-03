import { Head, usePage } from "@inertiajs/react";
import { useQuery } from "@tanstack/react-query";
import { Activity, CircleAlert, Flag, OctagonAlert, Package, RefreshCw, Users } from "lucide-react";
import type { ReactElement } from "react";
import {
	fetchOverview,
	formatDuration,
	type OverviewData,
	runVariant,
} from "../../components/admin/overview";
import { PageHeader } from "../../components/admin/PageHeader";
import { StatCard } from "../../components/admin/StatCard";
import { Badge } from "../../components/admin/ui/badge";
import { Button } from "../../components/admin/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "../../components/admin/ui/card";
import {
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
} from "../../components/admin/ui/table";
import { withQueryClient } from "../../lib/queryClient";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";

const REFRESH_MS = 30_000;

function WorkerCard({ worker }: { worker: OverviewData["worker"] }): ReactElement {
	const t = useT();
	const variant =
		worker.status === "ok" ? "success" : worker.status === "stale" ? "destructive" : "warning";
	const hint =
		worker.status === "ok" && worker.ageSeconds !== null
			? t("admin.dashboard.worker.age", { age: formatDuration(worker.ageSeconds) })
			: worker.status === "stale"
				? t("admin.dashboard.worker.staleHint")
				: worker.status === "never"
					? t("admin.dashboard.worker.neverHint")
					: "";
	return (
		<Card>
			<CardHeader className="flex flex-row items-center justify-between">
				<div className="grid gap-1.5">
					<CardTitle>{t("admin.dashboard.worker.title")}</CardTitle>
					<CardDescription>{hint}</CardDescription>
				</div>
				<Badge variant={variant}>{t(`admin.dashboard.worker.${worker.status}`)}</Badge>
			</CardHeader>
		</Card>
	);
}

function RunsTable({ runs, locale }: { runs: OverviewData["runs"]; locale: string }): ReactElement {
	const t = useT();
	const when = new Intl.DateTimeFormat(locale, { dateStyle: "short", timeStyle: "short" });
	const num = new Intl.NumberFormat(locale);
	return (
		<Card>
			<CardHeader className="flex flex-row items-center justify-between">
				<CardTitle>{t("admin.dashboard.runs.title")}</CardTitle>
				<a
					href="/admin/job-runs"
					className="text-sm text-primary underline-offset-4 hover:underline"
				>
					{t("admin.dashboard.runs.viewAll")}
				</a>
			</CardHeader>
			<CardContent>
				{runs.length === 0 ? (
					<p className="py-6 text-center text-sm text-muted">
						{t("admin.dashboard.runs.empty")}
					</p>
				) : (
					<Table>
						<TableHeader>
							<TableRow>
								<TableHead>{t("admin.dashboard.runs.job")}</TableHead>
								<TableHead>{t("admin.dashboard.runs.source")}</TableHead>
								<TableHead>{t("admin.dashboard.runs.mode")}</TableHead>
								<TableHead>{t("admin.dashboard.runs.status")}</TableHead>
								<TableHead className="text-right">
									{t("admin.dashboard.runs.seen")}
								</TableHead>
								<TableHead className="text-right">
									{t("admin.dashboard.runs.changed")}
								</TableHead>
								<TableHead className="text-right">
									{t("admin.dashboard.runs.failed")}
								</TableHead>
								<TableHead>{t("admin.dashboard.runs.started")}</TableHead>
								<TableHead className="text-right">
									{t("admin.dashboard.runs.duration")}
								</TableHead>
							</TableRow>
						</TableHeader>
						<TableBody>
							{runs.map((r) => (
								<TableRow key={r.id}>
									<TableCell className="font-medium">{r.job}</TableCell>
									<TableCell className="text-muted">{r.source || "–"}</TableCell>
									<TableCell className="text-muted">{r.mode}</TableCell>
									<TableCell>
										<Badge variant={runVariant(r.status)}>
											{t(`admin.dashboard.runs.statuses.${r.status}`)}
										</Badge>
									</TableCell>
									<TableCell className="text-right tabular-nums">
										{num.format(r.seen)}
									</TableCell>
									<TableCell className="text-right tabular-nums">
										{num.format(r.changed)}
									</TableCell>
									<TableCell className="text-right tabular-nums">
										{num.format(r.failed)}
									</TableCell>
									<TableCell className="whitespace-nowrap text-muted">
										{r.startedAt ? when.format(new Date(r.startedAt)) : "–"}
									</TableCell>
									<TableCell className="text-right tabular-nums text-muted">
										{r.durationSeconds === null
											? "–"
											: formatDuration(r.durationSeconds)}
									</TableCell>
								</TableRow>
							))}
						</TableBody>
					</Table>
				)}
			</CardContent>
		</Card>
	);
}

function Dashboard({ overview }: { overview: OverviewData }): ReactElement {
	const t = useT();
	const { locale } = usePage<SharedProps>().props;
	const num = new Intl.NumberFormat(locale);
	const { data, isFetching, isError, refetch, dataUpdatedAt } = useQuery({
		queryKey: ["admin", "overview"],
		queryFn: fetchOverview,
		initialData: overview,
		initialDataUpdatedAt: Date.parse(overview.generatedAt),
		staleTime: REFRESH_MS / 2,
		refetchInterval: REFRESH_MS,
	});
	const c = data.counts;
	const time = new Intl.DateTimeFormat(locale, { timeStyle: "medium" }).format(
		new Date(dataUpdatedAt),
	);

	return (
		<>
			<Head title={t("admin.dashboard.title")} />
			<PageHeader
				title={t("admin.dashboard.title")}
				description={`${t("admin.dashboard.subtitle")} ${t("admin.dashboard.updated", { time })}`}
				crumbs={[
					{ label: t("admin.title"), href: "/admin/overview" },
					{ label: t("admin.dashboard.title") },
				]}
				actions={
					<Button
						variant="outline"
						size="sm"
						onClick={() => void refetch()}
						disabled={isFetching}
					>
						<RefreshCw
							aria-hidden="true"
							className={isFetching ? "animate-spin motion-reduce:animate-none" : ""}
						/>
						{t("admin.dashboard.refresh")}
					</Button>
				}
			/>

			{isError && (
				<p role="status" className="mb-4 flex items-center gap-2 text-sm text-danger">
					<CircleAlert aria-hidden="true" className="size-4" />
					{t("admin.dashboard.refreshError")}
				</p>
			)}

			<div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
				<StatCard
					label={t("admin.dashboard.stats.items")}
					value={num.format(c.items)}
					icon={Package}
					hint={t("admin.dashboard.stats.approx")}
					href="/admin/items"
				/>
				<StatCard
					label={t("admin.dashboard.stats.users")}
					value={num.format(c.users)}
					icon={Users}
					hint={t("admin.dashboard.stats.approx")}
					href="/admin/users"
				/>
				<StatCard
					label={t("admin.dashboard.stats.openReports")}
					value={num.format(c.openReports)}
					icon={Flag}
					href="/admin/reports"
				/>
				<StatCard
					label={t("admin.dashboard.stats.runningRuns")}
					value={num.format(c.runningRuns)}
					icon={Activity}
					href="/admin/ingest"
				/>
				<StatCard
					label={t("admin.dashboard.stats.failedRuns")}
					value={num.format(c.failedRuns24h)}
					icon={OctagonAlert}
					tone={c.failedRuns24h > 0 ? "danger" : "default"}
					href="/admin/job-runs"
				/>
				<StatCard
					label={t("admin.dashboard.stats.parkedFailures")}
					value={num.format(c.parkedFailures)}
					icon={OctagonAlert}
					tone={c.parkedFailures > 0 ? "danger" : "default"}
					href="/admin/ingest-failures"
				/>
			</div>

			<div className="mt-4 grid gap-4">
				<WorkerCard worker={data.worker} />
				<RunsTable runs={data.runs} locale={locale} />
			</div>
		</>
	);
}

export default withQueryClient(Dashboard);
