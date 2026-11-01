import { useForm } from "@inertiajs/react";
import { type FormEvent, useState } from "react";

import { useLocalized } from "../lib/useLocalized";
import { useT } from "../lib/useT";
import Dialog from "./Dialog";

const REASONS = ["spam", "offensive", "harassment", "impersonation", "other"] as const;

const LABELS = {
	list: "social.report.reportList",
	user: "social.report.reportUser",
	review: "social.report.reportReview",
	comment: "social.report.reportComment",
} as const;

interface Props {
	targetType: "list" | "user" | "review" | "comment";
	targetId: string;
}

// Report button plus dialog for a public list, profile, review or comment (SDD 14 FR-14-9). The
// server answers 303 with a flash; a repeat report is a silent 200.
export default function ReportDialog({ targetType, targetId }: Props) {
	const t = useT();
	const l = useLocalized();
	const [open, setOpen] = useState(false);
	const form = useForm({ targetType, targetId, reason: "spam", details: "" });

	const submit = (e: FormEvent) => {
		e.preventDefault();
		form.post(l("/reports"), {
			preserveScroll: true,
			onSuccess: () => {
				form.reset("reason", "details");
				setOpen(false);
			},
		});
	};

	return (
		<>
			<button
				type="button"
				onClick={() => setOpen(true)}
				className="inline-flex min-h-11 items-center text-sm text-muted underline hover:text-fg"
			>
				{t(LABELS[targetType])}
			</button>
			<Dialog open={open} onOpenChange={setOpen} title={t("social.report.title")} size="sm">
				<form onSubmit={submit} className="mt-4 grid gap-3">
					<label className="grid gap-1 text-sm">
						{t("social.report.reason")}
						<select
							value={form.data.reason}
							onChange={(e) => form.setData("reason", e.target.value)}
							className="min-h-11 rounded-md border border-border bg-bg px-2"
						>
							{REASONS.map((r) => (
								<option key={r} value={r}>
									{t(`social.report.reasons.${r}`)}
								</option>
							))}
						</select>
					</label>
					<label className="grid gap-1 text-sm">
						{t("social.report.details")}
						<textarea
							value={form.data.details}
							onChange={(e) => form.setData("details", e.target.value)}
							maxLength={500}
							rows={3}
							className="rounded-md border border-border bg-bg px-2 py-1"
						/>
					</label>
					{form.errors.reason && (
						<p className="text-sm text-danger">{form.errors.reason}</p>
					)}
					<div className="flex justify-end gap-2">
						<button
							type="button"
							onClick={() => setOpen(false)}
							className="min-h-11 rounded-md border border-border px-4 text-sm"
						>
							{t("social.report.close")}
						</button>
						<button
							type="submit"
							disabled={form.processing}
							className="min-h-11 rounded-md bg-fg px-4 text-sm text-bg disabled:opacity-50"
						>
							{t("social.report.submit")}
						</button>
					</div>
				</form>
			</Dialog>
		</>
	);
}
