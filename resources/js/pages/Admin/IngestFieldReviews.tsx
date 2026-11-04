import { Head } from "@inertiajs/react";
import type { ReactElement } from "react";

import { ActionError } from "../../components/admin/ActionError";
import { PageHeader } from "../../components/admin/PageHeader";
import { Button } from "../../components/admin/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "../../components/admin/ui/card";
import { useAction } from "../../lib/adminAction";
import { useT } from "../../lib/useT";

interface FieldReview {
	id: number;
	itemId: string;
	itemType: string;
	itemTitle: string;
	field: string;
	value: string;
	source: string;
	confidence: number;
	reason: string;
}

function ReviewCard({ review }: { review: FieldReview }): ReactElement {
	const t = useT();
	const action = useAction();
	const decide = (verb: "accept" | "reject") =>
		void action.run(`/admin/ingest/field-reviews/${review.id}`, { action: verb });
	return (
		<Card>
			<CardHeader>
				<CardTitle className="text-base">{review.field}</CardTitle>
				<CardDescription>
					{t("admin.fieldReviews.onItem", {
						type: review.itemType,
						title: review.itemTitle,
					})}
					{" · "}
					{t("admin.fieldReviews.details", {
						source: review.source,
						confidence: review.confidence,
					})}
					{review.reason ? `, ${review.reason}` : ""}
				</CardDescription>
			</CardHeader>
			<CardContent className="grid gap-3">
				<div className="grid gap-1 text-sm">
					<span className="text-xs text-muted">{t("admin.fieldReviews.proposed")}</span>
					<code className="rounded-md bg-surface px-2 py-1.5 break-words">
						{review.value}
					</code>
				</div>
				<ActionError message={action.error} />
				<div className="flex gap-2">
					<Button disabled={action.pending} onClick={() => decide("accept")}>
						{t("admin.fieldReviews.accept")}
					</Button>
					<Button
						variant="outline"
						disabled={action.pending}
						onClick={() => decide("reject")}
					>
						{t("admin.fieldReviews.reject")}
					</Button>
				</div>
			</CardContent>
		</Card>
	);
}

export default function IngestFieldReviews({
	reviews,
}: {
	reviews: FieldReview[] | null;
}): ReactElement {
	const t = useT();
	const list = reviews ?? [];
	return (
		<>
			<Head title={t("admin.fieldReviews.title")} />
			<PageHeader
				title={t("admin.fieldReviews.title")}
				description={t("admin.fieldReviews.description")}
				crumbs={[
					{ label: t("admin.title"), href: "/admin/overview" },
					{ label: t("admin.ingest.title"), href: "/admin/ingest" },
					{ label: t("admin.fieldReviews.title") },
				]}
			/>
			{list.length === 0 ? (
				<p className="text-sm text-muted">{t("admin.fieldReviews.empty")}</p>
			) : (
				<div className="grid gap-4">
					{list.map((r) => (
						<ReviewCard key={r.id} review={r} />
					))}
				</div>
			)}
		</>
	);
}
