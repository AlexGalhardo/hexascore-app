import { Head } from "@inertiajs/react";
import type { ReactElement } from "react";

import { ActionError } from "../../components/admin/ActionError";
import { PageHeader } from "../../components/admin/PageHeader";
import { Badge } from "../../components/admin/ui/badge";
import { Button } from "../../components/admin/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "../../components/admin/ui/card";
import { useAction } from "../../lib/adminAction";
import { useT } from "../../lib/useT";

interface Conflict {
	id: number;
	qid: string;
	type: string;
	kind: string;
	key: string;
	value: string;
	otherID: string;
	otherSlug: string;
	otherTitle: string;
	similarity: number;
}

function ConflictCard({ conflict }: { conflict: Conflict }): ReactElement {
	const t = useT();
	const action = useAction();
	const decide = (verb: "allow" | "dismiss") =>
		void action.run(`/admin/ingest/conflicts/${conflict.id}`, { action: verb });
	const other = (
		<a
			href={`/admin/r/items/${conflict.otherID}`}
			className="font-medium text-primary underline-offset-4 hover:underline"
		>
			{conflict.otherTitle}
		</a>
	);
	return (
		<Card>
			<CardHeader className="flex flex-row items-center gap-2">
				<CardTitle className="text-base">{conflict.qid}</CardTitle>
				<Badge variant="outline">{conflict.type}</Badge>
			</CardHeader>
			<CardContent className="grid gap-3 text-sm">
				{conflict.kind === "external_id" ? (
					<p>
						{t("admin.conflicts.externalId", { key: conflict.key })}
						<code className="mx-1 rounded-md bg-surface px-1.5 py-0.5">
							{conflict.value}
						</code>
						{t("admin.conflicts.belongsTo")} {other}.
					</p>
				) : (
					<p>
						{t("admin.conflicts.similarTitle")}
						<code className="mx-1 rounded-md bg-surface px-1.5 py-0.5">
							{conflict.value}
						</code>
						{t("admin.conflicts.withinYear")} {other} (
						{t("admin.conflicts.similarity", { n: conflict.similarity.toFixed(2) })}).
					</p>
				)}
				<ActionError message={action.error} />
				<div className="flex gap-2">
					<Button disabled={action.pending} onClick={() => decide("allow")}>
						{t("admin.conflicts.allow")}
					</Button>
					<Button
						variant="outline"
						disabled={action.pending}
						onClick={() => decide("dismiss")}
					>
						{t("admin.conflicts.dismiss")}
					</Button>
				</div>
			</CardContent>
		</Card>
	);
}

export default function IngestConflicts({
	conflicts,
}: {
	conflicts: Conflict[] | null;
}): ReactElement {
	const t = useT();
	const list = conflicts ?? [];
	return (
		<>
			<Head title={t("admin.conflicts.title")} />
			<PageHeader
				title={t("admin.conflicts.title")}
				description={t("admin.conflicts.description")}
				crumbs={[
					{ label: t("admin.title"), href: "/admin/overview" },
					{ label: t("admin.ingest.title"), href: "/admin/ingest" },
					{ label: t("admin.conflicts.title") },
				]}
			/>
			{list.length === 0 ? (
				<p className="text-sm text-muted">{t("admin.conflicts.empty")}</p>
			) : (
				<div className="grid gap-4">
					{list.map((c) => (
						<ConflictCard key={c.id} conflict={c} />
					))}
				</div>
			)}
		</>
	);
}
