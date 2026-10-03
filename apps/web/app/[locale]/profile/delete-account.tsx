"use client";

import { Dialog } from "@base-ui/react/dialog";
import { useLocale, useTranslations } from "next-intl";
import { useEffect, useState, useTransition } from "react";
import { scheduleDeletion } from "@/actions/account";
import { DIALOG_BACKDROP, dialogPopup } from "@/components/ui/dialog";
import { buttonClass, Notice } from "@/components/ui/form";
import { getPathname } from "@/i18n/navigation";
import { DELETION_CONFIRM_SECONDS, secondsLeft } from "@/lib/account-deletion";
import type { ActionError } from "@/lib/action-result";

// The countdown restarts every time the dialog opens so the confirm button can't be clicked by reflex.
function useCountdown(active: boolean): number {
	const [left, setLeft] = useState(DELETION_CONFIRM_SECONDS);
	useEffect(() => {
		if (!active) return;
		const started = Date.now();
		setLeft(DELETION_CONFIRM_SECONDS);
		const timer = setInterval(() => {
			const remaining = secondsLeft(started, Date.now());
			setLeft(remaining);
			if (remaining === 0) clearInterval(timer);
		}, 250);
		return () => clearInterval(timer);
	}, [active]);
	return left;
}

export function DeleteAccount() {
	const t = useTranslations("account.delete");
	const tAction = useTranslations("actions");
	const locale = useLocale();
	const [open, setOpen] = useState(false);
	const [error, setError] = useState<ActionError | null>(null);
	const [pending, startTransition] = useTransition();
	const left = useCountdown(open);

	function confirm(): void {
		startTransition(async () => {
			const result = await scheduleDeletion();
			if (result.ok) window.location.assign(getPathname({ href: "/", locale }));
			else setError(result.error);
		});
	}

	return (
		<div className="flex flex-col gap-4">
			<p className="text-sm text-muted">{t("body")}</p>
			<Dialog.Root
				open={open}
				onOpenChange={(value) => {
					setOpen(value);
					setError(null);
				}}
			>
				<Dialog.Trigger className={buttonClass("danger", "self-start")}>{t("button")}</Dialog.Trigger>
				<Dialog.Portal>
					<Dialog.Backdrop className={DIALOG_BACKDROP} />
					<Dialog.Popup className={`${dialogPopup()} flex flex-col gap-4 p-6`}>
						<Dialog.Title className="font-serif text-xl font-bold">{t("dialogTitle")}</Dialog.Title>
						<Dialog.Description className="text-sm">{t("dialogBody")}</Dialog.Description>
						{error && <Notice tone="error">{tAction(error)}</Notice>}
						<div className="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
							<Dialog.Close className={buttonClass("secondary")}>{t("cancel")}</Dialog.Close>
							<button
								type="button"
								onClick={confirm}
								disabled={left > 0 || pending}
								className={buttonClass("danger", "tabular-nums")}
							>
								{left > 0 ? t("wait", { seconds: left }) : t("confirm")}
							</button>
						</div>
					</Dialog.Popup>
				</Dialog.Portal>
			</Dialog.Root>
		</div>
	);
}
