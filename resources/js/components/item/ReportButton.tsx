import { router } from "@inertiajs/react";
import { useState } from "react";

import { useT } from "../../lib/useT";

const REASONS = ["spam", "offensive", "other"] as const;

export default function ReportButton({ itemId }: { itemId: string }) {
	const t = useT();
	const [open, setOpen] = useState(false);
	const [reason, setReason] = useState<string>(REASONS[0]);
	const [sent, setSent] = useState(false);

	if (sent) return <p role="status">{t("social.report.sent")}</p>;
	if (!open) {
		return (
			<button type="button" className="min-h-11 underline" onClick={() => setOpen(true)}>
				{t("social.report.title")}
			</button>
		);
	}

	return (
		<form
			className="grid gap-2"
			onSubmit={(e) => {
				e.preventDefault();
				router.post(
					`/items/${itemId}/report`,
					{ reason },
					{ preserveScroll: true, onSuccess: () => setSent(true) },
				);
			}}
		>
			<label className="grid gap-1">
				<span>{t("social.report.reason")}</span>
				<select
					className="min-h-11 rounded border px-2"
					value={reason}
					onChange={(e) => setReason(e.target.value)}
				>
					{REASONS.map((r) => (
						<option key={r} value={r}>
							{t(`social.report.reasons.${r}`)}
						</option>
					))}
				</select>
			</label>
			<div className="flex gap-2">
				<button type="submit" className="min-h-11 rounded border px-4">
					{t("social.report.submit")}
				</button>
				<button type="button" className="min-h-11 px-4" onClick={() => setOpen(false)}>
					{t("social.report.close")}
				</button>
			</div>
		</form>
	);
}
