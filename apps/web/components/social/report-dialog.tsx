"use client";

import { Dialog } from "@base-ui/react/dialog";
import { useTranslations } from "next-intl";
import { useId, useState, useTransition } from "react";
import { report } from "@/actions/reports";
import { DIALOG_BACKDROP, dialogPopup } from "@/components/ui/dialog";
import { buttonClass, inputClass, Notice } from "@/components/ui/form";
import type { ActionError } from "@/lib/action-result";
import { REPORT_DETAILS_MAX, REPORT_REASONS } from "@/lib/social/constants";

type Target = { type: "list"; id: string } | { type: "user"; nickname: string };

export function ReportDialog({ target }: { target: Target }) {
	const t = useTranslations("social.report");
	const tAction = useTranslations("actions");
	const id = useId();
	const [reason, setReason] = useState<(typeof REPORT_REASONS)[number]>("spam");
	const [details, setDetails] = useState("");
	const [result, setResult] = useState<"sent" | ActionError | null>(null);
	const [pending, startTransition] = useTransition();

	function submit(): void {
		setResult(null);
		startTransition(async () => {
			const response = await report({ target, reason, details });
			setResult(response.ok ? "sent" : response.error);
		});
	}

	return (
		<Dialog.Root onOpenChange={(open) => open && setResult(null)}>
			<Dialog.Trigger className="min-h-11 px-2 text-sm text-muted underline underline-offset-2">
				{target.type === "list" ? t("reportList") : t("reportUser")}
			</Dialog.Trigger>
			<Dialog.Portal>
				<Dialog.Backdrop className={DIALOG_BACKDROP} />
				<Dialog.Popup className={`${dialogPopup()} grid gap-4 p-4`}>
					<Dialog.Title className="font-serif text-xl font-bold">{t("title")}</Dialog.Title>
					{result === "sent" ? (
						<Notice tone="success">{t("sent")}</Notice>
					) : (
						<form
							className="grid gap-3"
							onSubmit={(e) => {
								e.preventDefault();
								submit();
							}}
						>
							<label htmlFor={`${id}-reason`} className="text-sm font-medium">
								{t("reason")}
							</label>
							<select
								id={`${id}-reason`}
								value={reason}
								onChange={(e) => setReason(e.target.value as (typeof REPORT_REASONS)[number])}
								className={inputClass}
							>
								{REPORT_REASONS.map((r) => (
									<option key={r} value={r}>
										{t(`reasons.${r}`)}
									</option>
								))}
							</select>
							<label htmlFor={`${id}-details`} className="text-sm font-medium">
								{t("details")}
							</label>
							<textarea
								id={`${id}-details`}
								rows={3}
								maxLength={REPORT_DETAILS_MAX}
								value={details}
								onChange={(e) => setDetails(e.target.value)}
								className={`${inputClass} py-2`}
							/>
							{result ? <Notice tone="error">{tAction(result)}</Notice> : null}
							<button type="submit" disabled={pending} className={buttonClass("danger")}>
								{t("submit")}
							</button>
						</form>
					)}
					<Dialog.Close className={buttonClass("secondary")}>{t("close")}</Dialog.Close>
				</Dialog.Popup>
			</Dialog.Portal>
		</Dialog.Root>
	);
}
