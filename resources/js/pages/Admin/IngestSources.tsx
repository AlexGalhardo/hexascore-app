import { Head } from "@inertiajs/react";
import type { ReactElement } from "react";

import { runVariant } from "../../components/admin/overview";
import { PageHeader } from "../../components/admin/PageHeader";
import type { RunRow, SourceRow } from "../../components/admin/types";
import { Badge } from "../../components/admin/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "../../components/admin/ui/card";
import {
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
} from "../../components/admin/ui/table";
import { useT } from "../../lib/useT";

interface Props {
	ready: boolean;
	sources?: SourceRow[] | null;
	runs?: RunRow[] | null;
}

export default function IngestSources({ ready, sources, runs }: Props): ReactElement {
	const t = useT();
	const list = sources ?? [];
	const last = runs ?? [];
	return (
		<>
			<Head title={t("admin.sources.title")} />
			<PageHeader
				title={t("admin.sources.title")}
				description={t("admin.sources.description")}
				crumbs={[
					{ label: t("admin.title"), href: "/admin/overview" },
					{ label: t("admin.ingest.title"), href: "/admin/ingest" },
					{ label: t("admin.sources.title") },
				]}
			/>
			{!ready ? (
				<p className="text-sm text-muted">{t("admin.sources.notReady")}</p>
			) : (
				<div className="grid gap-4">
					<Card>
						<CardHeader>
							<CardTitle>{t("admin.sources.title")}</CardTitle>
						</CardHeader>
						<CardContent>
							{list.length === 0 ? (
								<p className="py-3 text-sm text-muted">
									{t("admin.sources.empty")}
								</p>
							) : (
								<Table>
									<TableHeader>
										<TableRow>
											<TableHead>{t("admin.ingest.source")}</TableHead>
											<TableHead className="text-right">
												{t("admin.sources.records")}
											</TableHead>
											<TableHead className="text-right">
												{t("admin.sources.pending")}
											</TableHead>
											<TableHead>{t("admin.sources.lastFetched")}</TableHead>
											<TableHead>{t("admin.sources.lastChecked")}</TableHead>
										</TableRow>
									</TableHeader>
									<TableBody>
										{list.map((s) => (
											<TableRow key={s.source}>
												<TableCell className="font-medium">
													{s.source}
												</TableCell>
												<TableCell className="text-right tabular-nums">
													{s.records}
												</TableCell>
												<TableCell className="text-right tabular-nums">
													{s.pending}
												</TableCell>
												<TableCell className="text-muted">
													{s.lastFetched || "–"}
												</TableCell>
												<TableCell className="text-muted">
													{s.lastChecked || "–"}
												</TableCell>
											</TableRow>
										))}
									</TableBody>
								</Table>
							)}
						</CardContent>
					</Card>
					<Card>
						<CardHeader>
							<CardTitle>{t("admin.sources.lastRuns")}</CardTitle>
						</CardHeader>
						<CardContent>
							{last.length === 0 ? (
								<p className="py-3 text-sm text-muted">
									{t("admin.ingest.noRuns")}
								</p>
							) : (
								<Table>
									<TableHeader>
										<TableRow>
											<TableHead>{t("admin.ingest.job")}</TableHead>
											<TableHead>{t("admin.ingest.source")}</TableHead>
											<TableHead>{t("admin.ingest.statusCol")}</TableHead>
											<TableHead className="text-right">
												{t("admin.ingest.seen")}
											</TableHead>
											<TableHead className="text-right">
												{t("admin.ingest.changed")}
											</TableHead>
											<TableHead className="text-right">
												{t("admin.ingest.failed")}
											</TableHead>
											<TableHead>{t("admin.ingest.started")}</TableHead>
											<TableHead>{t("admin.ingest.finished")}</TableHead>
										</TableRow>
									</TableHeader>
									<TableBody>
										{last.map((r) => (
											<TableRow key={r.id}>
												<TableCell className="font-medium">
													{r.job}
												</TableCell>
												<TableCell className="text-muted">
													{r.source || "–"}
												</TableCell>
												<TableCell>
													<Badge variant={runVariant(r.status)}>
														{t(
															`admin.dashboard.runs.statuses.${r.status}`,
														)}
													</Badge>
												</TableCell>
												<TableCell className="text-right tabular-nums">
													{r.seen}
												</TableCell>
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
											</TableRow>
										))}
									</TableBody>
								</Table>
							)}
						</CardContent>
					</Card>
				</div>
			)}
		</>
	);
}
