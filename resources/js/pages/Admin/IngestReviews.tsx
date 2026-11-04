import { Head } from "@inertiajs/react";
import type { ReactElement } from "react";

import { ActionError } from "../../components/admin/ActionError";
import { Field } from "../../components/admin/Field";
import { PageHeader } from "../../components/admin/PageHeader";
import { Button } from "../../components/admin/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "../../components/admin/ui/card";
import { Select } from "../../components/admin/ui/select";
import { useAction } from "../../lib/adminAction";
import { useT } from "../../lib/useT";

interface Candidate {
	id: string;
	name: string;
	kind: string;
	qid: string;
}

interface Review {
	id: number;
	itemId: string;
	itemType: string;
	itemTitle: string;
	role: string;
	source: string;
	rawName: string;
	rawExternalID: string;
	rawYear: number;
	candidates: Candidate[] | null;
}

function ReviewCard({ review }: { review: Review }): ReactElement {
	const t = useT();
	const action = useAction();
	const candidates = review.candidates ?? [];
	const url = `/admin/ingest/reviews/${review.id}`;
	const decide = (verb: "link" | "create" | "dismiss", entityId?: string) =>
		void action.run(url, { action: verb, entity_id: entityId });
	return (
		<Card>
			<CardHeader>
				<CardTitle className="text-base">
					{review.rawName}
					{review.rawYear ? ` (${review.rawYear})` : ""}
				</CardTitle>
				<CardDescription>
					{t("admin.reviews.asRole", {
						role: review.role,
						type: review.itemType,
						title: review.itemTitle,
					})}
					{" · "}
					{t("admin.reviews.source", { source: review.source })}
					{review.rawExternalID
						? `, ${t("admin.reviews.externalId", { id: review.rawExternalID })}`
						: ""}
				</CardDescription>
			</CardHeader>
			<CardContent>
				<form
					className="grid gap-3"
					onSubmit={(e) => {
						e.preventDefault();
						const entity = String(new FormData(e.currentTarget).get("entity_id") ?? "");
						decide("link", entity || undefined);
					}}
				>
					{candidates.length > 0 && (
						<Field label={t("admin.reviews.candidate")} className="max-w-xl">
							{(id) => (
								<Select id={id} name="entity_id">
									{candidates.map((c) => (
										<option key={c.id} value={c.id}>
											{`${c.name} (${c.kind}${c.qid ? `, ${c.qid}` : ""}) ${c.id}`}
										</option>
									))}
								</Select>
							)}
						</Field>
					)}
					<ActionError message={action.error} />
					<div className="flex flex-wrap gap-2">
						<Button type="submit" disabled={action.pending || candidates.length === 0}>
							{t("admin.reviews.link")}
						</Button>
						<Button
							variant="outline"
							disabled={action.pending}
							onClick={() => decide("create")}
						>
							{t("admin.reviews.create")}
						</Button>
						<Button
							variant="outline"
							disabled={action.pending}
							onClick={() => decide("dismiss")}
						>
							{t("admin.reviews.dismiss")}
						</Button>
					</div>
				</form>
			</CardContent>
		</Card>
	);
}

export default function IngestReviews({ reviews }: { reviews: Review[] | null }): ReactElement {
	const t = useT();
	const list = reviews ?? [];
	return (
		<>
			<Head title={t("admin.reviews.title")} />
			<PageHeader
				title={t("admin.reviews.title")}
				description={t("admin.reviews.description")}
				crumbs={[
					{ label: t("admin.title"), href: "/admin/overview" },
					{ label: t("admin.ingest.title"), href: "/admin/ingest" },
					{ label: t("admin.reviews.title") },
				]}
			/>
			{list.length === 0 ? (
				<p className="text-sm text-muted">{t("admin.reviews.empty")}</p>
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
