import { router } from "@inertiajs/react";
import { useState } from "react";

import { useT } from "../../lib/useT";
import type { ItemType } from "./types";

const STATUSES = ["want", "in_progress", "completed", "dropped"] as const;

export default function StatusPicker({
	itemId,
	type,
	status,
}: {
	itemId: string;
	type: ItemType;
	status: string | null;
}) {
	const t = useT();
	const [value, setValue] = useState(status ?? "");
	const [busy, setBusy] = useState(false);

	const change = (next: string) => {
		const previous = value;
		setValue(next);
		setBusy(true);
		const opts = {
			preserveScroll: true,
			onError: () => setValue(previous),
			onFinish: () => setBusy(false),
		};
		if (next === "") router.delete(`/items/${itemId}/status`, opts);
		else router.put(`/items/${itemId}/status`, { status: next }, opts);
	};

	return (
		<label className="grid gap-1">
			<span className="font-semibold">{t("social.status.label")}</span>
			<select
				className="min-h-11 rounded border px-2"
				value={value}
				disabled={busy}
				onChange={(e) => change(e.target.value)}
			>
				<option value="">{t("social.status.none")}</option>
				{STATUSES.map((s) => (
					<option key={s} value={s}>
						{t(`social.status.${type}.${s}`)}
					</option>
				))}
			</select>
		</label>
	);
}
