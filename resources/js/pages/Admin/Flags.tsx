import { Head } from "@inertiajs/react";
import type { ReactElement } from "react";

import { ActionButton, ActionForm } from "../../components/admin/ActionForm";
import { PageHeader } from "../../components/admin/PageHeader";
import { Badge } from "../../components/admin/ui/badge";
import { Card, CardContent } from "../../components/admin/ui/card";
import { Input } from "../../components/admin/ui/input";
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

function FlagControls({ flag }: { flag: FlagRow }): ReactElement {
	const t = useT();
	if (flag.locked) return <span className="text-muted">–</span>;
	const action = `/admin/r/flags/${encodeURIComponent(flag.name)}`;
	return (
		<div className="flex flex-wrap items-center justify-end gap-2">
			{flag.kind === "bool" ? (
				<ActionButton
					action={action}
					fields={{ value: flag.value === "true" ? "false" : "true" }}
				>
					{flag.value === "true" ? t("admin.flags.turnOff") : t("admin.flags.turnOn")}
				</ActionButton>
			) : (
				<ActionForm
					action={action}
					submit={t("admin.common.save")}
					className="grid-cols-[minmax(8rem,1fr)_auto] items-end"
				>
					<Input
						name="value"
						defaultValue={flag.value}
						aria-label={flag.name}
						type={flag.kind === "int" ? "number" : "text"}
					/>
				</ActionForm>
			)}
			{flag.stored && (
				<ActionButton action={action} fields={{ reset: "1" }}>
					{t("admin.flags.reset")}
				</ActionButton>
			)}
		</div>
	);
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
								<TableHead className="text-right">
									{t("admin.common.actions")}
								</TableHead>
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
									<TableCell className="text-right">
										<FlagControls flag={f} />
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
