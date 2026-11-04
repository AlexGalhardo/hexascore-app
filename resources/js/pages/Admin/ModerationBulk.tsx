import { Head } from "@inertiajs/react";
import type { ReactElement } from "react";

import { PageHeader } from "../../components/admin/PageHeader";
import { Badge } from "../../components/admin/ui/badge";
import { Card, CardContent } from "../../components/admin/ui/card";
import {
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
} from "../../components/admin/ui/table";
import { useT } from "../../lib/useT";

interface Result {
	id: string;
	ok: boolean;
	code: string;
}

export default function ModerationBulk({ results }: { results: Result[] | null }): ReactElement {
	const t = useT();
	return (
		<>
			<Head title={t("admin.bulk.title")} />
			<PageHeader
				title={t("admin.bulk.title")}
				crumbs={[
					{ label: t("admin.title"), href: "/admin/overview" },
					{ label: t("admin.moderation.title"), href: "/admin/moderation" },
					{ label: t("admin.bulk.title") },
				]}
			/>
			<Card>
				<CardContent>
					<Table>
						<TableHeader>
							<TableRow>
								<TableHead>{t("admin.bulk.suggestion")}</TableHead>
								<TableHead>{t("admin.bulk.outcome")}</TableHead>
							</TableRow>
						</TableHeader>
						<TableBody>
							{(results ?? []).map((r) => (
								<TableRow key={r.id}>
									<TableCell className="font-mono text-xs">{r.id}</TableCell>
									<TableCell>
										<Badge variant={r.ok ? "success" : "destructive"}>
											{r.ok ? t("admin.bulk.done") : r.code}
										</Badge>
									</TableCell>
								</TableRow>
							))}
						</TableBody>
					</Table>
				</CardContent>
			</Card>
		</>
	);
}
