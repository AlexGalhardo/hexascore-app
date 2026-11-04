import { Head } from "@inertiajs/react";
import type { ReactElement } from "react";

import { PageHeader } from "../../components/admin/PageHeader";
import { Badge } from "../../components/admin/ui/badge";
import { buttonVariants } from "../../components/admin/ui/button";
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

interface FlagRow {
	name: string;
	description: string;
	kind: string;
	default: string;
	value: string;
	locked: boolean;
	stored: boolean;
}

export default function Flags({ flags }: { flags: FlagRow[] | null }): ReactElement {
	const t = useT();
	const list = flags ?? [];
	return (
		<>
			<Head title={t("admin.flags.title")} />
			<PageHeader
				title={t("admin.flags.title")}
				description={t("admin.flags.description")}
				crumbs={[
					{ label: t("admin.title"), href: "/admin/overview" },
					{ label: t("admin.flags.title") },
				]}
				actions={
					<a
						href="/admin/_flags"
						className={buttonVariants({ variant: "outline", size: "sm" })}
					>
						{t("admin.flags.change")}
					</a>
				}
			/>
			<Card>
				<CardContent>
					<Table>
						<TableHeader>
							<TableRow>
								<TableHead>{t("admin.flags.name")}</TableHead>
								<TableHead>{t("admin.flags.value")}</TableHead>
								<TableHead>{t("admin.flags.default")}</TableHead>
								<TableHead>{t("admin.res.col.status")}</TableHead>
							</TableRow>
						</TableHeader>
						<TableBody>
							{list.map((f) => (
								<TableRow key={f.name}>
									<TableCell>
										<div className="font-medium">{f.name}</div>
										<div className="max-w-xl text-xs text-muted">
											{f.description}
										</div>
									</TableCell>
									<TableCell>
										<Badge
											variant={f.value === "true" ? "success" : "secondary"}
										>
											{f.value}
										</Badge>
									</TableCell>
									<TableCell className="text-muted">{f.default}</TableCell>
									<TableCell>
										<span className="flex flex-wrap gap-1">
											{f.locked && (
												<Badge variant="outline">
													{t("admin.flags.locked")}
												</Badge>
											)}
											{f.stored && !f.locked && (
												<Badge variant="warning">
													{t("admin.flags.overridden")}
												</Badge>
											)}
										</span>
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
