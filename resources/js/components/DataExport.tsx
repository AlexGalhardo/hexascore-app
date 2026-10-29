import { useState } from "react";

import { useT } from "../lib/useT";

type State = "idle" | "sending" | "requested" | "cooldown" | "failed";

// A fetch rather than an Inertia visit: the 429 export-cooldown answer is problem+json, which Inertia would show as an invalid-response modal.
export default function DataExport({ className }: { className?: string }) {
	const t = useT();
	const [state, setState] = useState<State>("idle");

	const request = async () => {
		setState("sending");
		try {
			const res = await fetch("/profile/export", {
				method: "POST",
				headers: { Accept: "application/json" },
				credentials: "same-origin",
			});
			if (res.status === 429) setState("cooldown");
			else setState(res.ok ? "requested" : "failed");
		} catch {
			setState("failed");
		}
	};

	return (
		<>
			<p>{t("account.export.body")}</p>
			<button
				type="button"
				className={className}
				disabled={state === "sending"}
				onClick={request}
			>
				{t("account.export.button")}
			</button>
			{state !== "idle" && state !== "sending" && (
				<p role="status">{t(`account.export.${state}`)}</p>
			)}
		</>
	);
}
