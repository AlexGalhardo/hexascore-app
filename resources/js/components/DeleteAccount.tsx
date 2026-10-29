import { router } from "@inertiajs/react";
import { useEffect, useState } from "react";

import { useT } from "../lib/useT";
import Dialog from "./Dialog";

const COUNTDOWN_SECONDS = 10;

// The confirm button stays disabled for 10 s so a stray click cannot start the grace period (SDD 03 FR-03-15).
export default function DeleteAccount({ className }: { className?: string }) {
	const t = useT();
	const [open, setOpen] = useState(false);
	const [seconds, setSeconds] = useState(COUNTDOWN_SECONDS);

	useEffect(() => {
		if (!open) {
			setSeconds(COUNTDOWN_SECONDS);
			return;
		}
		const timer = setInterval(() => setSeconds((s) => Math.max(0, s - 1)), 1000);
		return () => clearInterval(timer);
	}, [open]);

	return (
		<>
			<p>{t("account.delete.body")}</p>
			<button type="button" className={className} onClick={() => setOpen(true)}>
				{t("account.delete.button")}
			</button>
			<Dialog
				open={open}
				onOpenChange={setOpen}
				title={t("account.delete.dialogTitle")}
				description={t("account.delete.dialogBody")}
			>
				<div className="mt-4 flex flex-wrap gap-3">
					<button
						type="button"
						className={className}
						disabled={seconds > 0}
						onClick={() => router.post("/profile/deletion", { confirm: true })}
					>
						{seconds > 0
							? t("account.delete.wait", { seconds })
							: t("account.delete.confirm")}
					</button>
					<button
						type="button"
						className="rounded-md border px-4 py-2"
						onClick={() => setOpen(false)}
					>
						{t("account.delete.cancel")}
					</button>
				</div>
			</Dialog>
		</>
	);
}
