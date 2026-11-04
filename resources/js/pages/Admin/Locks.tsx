import { Head } from "@inertiajs/react";
import type { ReactElement } from "react";

import { ActionButton, ActionForm } from "../../components/admin/ActionForm";
import { Field } from "../../components/admin/Field";
import { PageHeader } from "../../components/admin/PageHeader";
import { Card, CardContent, CardHeader, CardTitle } from "../../components/admin/ui/card";
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

interface Lock {
	id: string;
	targetType: string;
	targetId: string;
	field: string;
	reason: string;
	lockedBy: number;
	expiresAt: string;
}

export default function Locks({
	locks,
	targets,
}: {
	locks: Lock[] | null;
	targets: string[] | null;
}): ReactElement {
	const t = useT();
	const list = locks ?? [];
	return (
		<>
			<Head title={t("admin.locks.title")} />
			<PageHeader
				title={t("admin.locks.title")}
				description={t("admin.locks.description")}
				crumbs={[
					{ label: t("admin.title"), href: "/admin/overview" },
					{ label: t("admin.locks.title") },
				]}
			/>
			<div className="grid gap-4">
				<Card>
					<CardContent className="pt-0">
						{list.length === 0 ? (
							<p className="py-3 text-sm text-muted">{t("admin.locks.empty")}</p>
						) : (
							<Table>
								<TableHeader>
									<TableRow>
										<TableHead>{t("admin.locks.target")}</TableHead>
										<TableHead>{t("admin.locks.field")}</TableHead>
										<TableHead>{t("admin.locks.reason")}</TableHead>
										<TableHead>{t("admin.locks.by")}</TableHead>
										<TableHead>{t("admin.locks.expires")}</TableHead>
										<TableHead className="text-right">
											{t("admin.common.actions")}
										</TableHead>
									</TableRow>
								</TableHeader>
								<TableBody>
									{list.map((l) => (
										<TableRow key={l.id}>
											<TableCell>
												<span className="font-medium">{l.targetType}</span>{" "}
												<span className="text-muted">{l.targetId}</span>
											</TableCell>
											<TableCell>{l.field}</TableCell>
											<TableCell className="text-muted">{l.reason}</TableCell>
											<TableCell className="text-muted">
												{l.lockedBy ? `#${l.lockedBy}` : "–"}
											</TableCell>
											<TableCell className="whitespace-nowrap text-muted">
												{l.expiresAt
													? t("admin.locks.expiresUtc", {
															date: l.expiresAt,
														})
													: t("admin.common.never")}
											</TableCell>
											<TableCell className="text-right">
												<ActionButton
													action={`/admin/locks/${l.id}/delete`}
												>
													{t("admin.locks.unlock")}
												</ActionButton>
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
						<CardTitle>{t("admin.locks.lockTitle")}</CardTitle>
					</CardHeader>
					<CardContent>
						<ActionForm
							action="/admin/locks"
							submit={t("admin.locks.lock")}
							className="md:grid-cols-2"
						>
							<Field label={t("admin.locks.targetType")}>
								{(id) => (
									<Select id={id} name="target">
										{(targets ?? []).map((x) => (
											<option key={x}>{x}</option>
										))}
									</Select>
								)}
							</Field>
							<Field label={t("admin.locks.targetId")}>
								{(id) => (
									<Input
										id={id}
										name="target_id"
										required
										pattern="[0-9a-fA-F\-]{36}"
										maxLength={36}
									/>
								)}
							</Field>
							<Field label={t("admin.locks.field")}>
								{(id) => (
									<Input
										id={id}
										name="field"
										maxLength={100}
										placeholder={t("admin.locks.fieldPlaceholder")}
									/>
								)}
							</Field>
							<Field label={t("admin.locks.reason")}>
								{(id) => <Input id={id} name="reason" required maxLength={500} />}
							</Field>
							<Field label={t("admin.locks.expiresAfter")}>
								{(id) => (
									<Input id={id} name="days" type="number" min={1} max={365} />
								)}
							</Field>
						</ActionForm>
					</CardContent>
				</Card>
			</div>
		</>
	);
}
