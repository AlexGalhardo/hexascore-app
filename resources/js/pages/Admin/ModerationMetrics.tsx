import { Head, Link } from "@inertiajs/react";
import type { ReactElement } from "react";

import { formatDuration } from "../../components/admin/overview";
import { PageHeader } from "../../components/admin/PageHeader";
import { buttonVariants } from "../../components/admin/ui/button";
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
import { useT } from "../../lib/useT";

interface Window {
	days: number;
	decided: number;
	approved: number;
	approvalRate: number;
	p50Seconds: number;
	p95Seconds: number;
}

interface Props {
	open: number;
	oldestOpenSeconds: number;
	week: Window;
	month: Window;
}

function WindowRow({ w }: { w: Window }): ReactElement {
	const t = useT();
	return (
		<TableRow>
			<TableCell>{t("admin.metrics.days", { n: w.days })}</TableCell>
			<TableCell className="tabular-nums">{w.decided}</TableCell>
			{w.decided > 0 ? (
				<>
					<TableCell className="tabular-nums">
						{w.approved} / {w.decided} ({w.approvalRate.toFixed(2)})
					</TableCell>
					<TableCell className="tabular-nums">{formatDuration(w.p50Seconds)}</TableCell>
					<TableCell className="tabular-nums">{formatDuration(w.p95Seconds)}</TableCell>
				</>
			) : (
				<>
					<TableCell>–</TableCell>
					<TableCell>–</TableCell>
					<TableCell>–</TableCell>
				</>
			)}
		</TableRow>
	);
}

export default function ModerationMetrics({
	open,
	oldestOpenSeconds,
	week,
	month,
}: Props): ReactElement {
	const t = useT();
	return (
		<>
			<Head title={t("admin.metrics.title")} />
			<PageHeader
				title={t("admin.metrics.title")}
				crumbs={[
					{ label: t("admin.title"), href: "/admin/overview" },
					{ label: t("admin.moderation.title"), href: "/admin/moderation" },
					{ label: t("admin.metrics.title") },
				]}
				actions={
					<Link
						href="/admin/moderation"
						className={buttonVariants({ variant: "outline", size: "sm" })}
					>
						{t("admin.metrics.openQueue")}
					</Link>
				}
			/>
			<div className="grid gap-4">
				<div className="grid gap-4 sm:grid-cols-2">
					<Card>
						<CardHeader>
							<CardDescription>{t("admin.metrics.queueSize")}</CardDescription>
							<CardTitle className="text-3xl tabular-nums">{open}</CardTitle>
						</CardHeader>
					</Card>
					<Card>
						<CardHeader>
							<CardDescription>{t("admin.metrics.oldestOpen")}</CardDescription>
							<CardTitle className="text-3xl tabular-nums">
								{open > 0
									? formatDuration(oldestOpenSeconds)
									: t("admin.metrics.none")}
							</CardTitle>
						</CardHeader>
					</Card>
				</div>
				<Card>
					<CardContent>
						<Table>
							<TableHeader>
								<TableRow>
									<TableHead>{t("admin.metrics.window")}</TableHead>
									<TableHead>{t("admin.metrics.decisions")}</TableHead>
									<TableHead>{t("admin.metrics.approvalRate")}</TableHead>
									<TableHead>{t("admin.metrics.p50")}</TableHead>
									<TableHead>{t("admin.metrics.p95")}</TableHead>
								</TableRow>
							</TableHeader>
							<TableBody>
								<WindowRow w={week} />
								<WindowRow w={month} />
							</TableBody>
						</Table>
						<p className="pt-3 text-sm text-muted">{t("admin.metrics.target")}</p>
					</CardContent>
				</Card>
			</div>
		</>
	);
}
