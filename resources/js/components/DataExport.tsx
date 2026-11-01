import { useMutation } from "@tanstack/react-query";
import { withQueryClient } from "../lib/queryClient";

import { useT } from "../lib/useT";

type State = "idle" | "sending" | "requested" | "cooldown" | "failed";

// A fetch rather than an Inertia visit: the 429 export-cooldown answer is problem+json, which Inertia would show as an invalid-response modal.
function DataExport({ className }: { className?: string }) {
	const t = useT();
	const exp = useMutation({
		mutationFn: async (): Promise<number> => {
			const res = await fetch("/profile/export", {
				method: "POST",
				headers: { Accept: "application/json" },
				credentials: "same-origin",
			});
			return res.status;
		},
	});
	const state: State = exp.isPending
		? "sending"
		: exp.isError
			? "failed"
			: exp.data === undefined
				? "idle"
				: exp.data === 429
					? "cooldown"
					: exp.data >= 200 && exp.data < 300
						? "requested"
						: "failed";

	return (
		<>
			<p>{t("account.export.body")}</p>
			<button
				type="button"
				className={className}
				disabled={state === "sending"}
				onClick={() => exp.mutate()}
			>
				{t("account.export.button")}
			</button>
			{state !== "idle" && state !== "sending" && (
				<p role="status">{t(`account.export.${state}`)}</p>
			)}
		</>
	);
}

export default withQueryClient(DataExport);
