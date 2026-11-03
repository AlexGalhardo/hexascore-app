import { router } from "@inertiajs/react";
import { useState } from "react";

import { useLocalized } from "../../lib/useLocalized";
import { useT } from "../../lib/useT";
import { btnClass } from "../ui/Btn";
import type { ItemType } from "./types";

const STATUSES = ["want", "in_progress", "completed", "dropped"] as const;

// Four toggle buttons (design v1 status picker); pressing the active one clears the status.
// Unreleased items only allow "want"; an already-set status can still be cleared so it is not stuck.
export default function StatusPicker({
	itemId,
	type,
	status,
	unreleased = false,
}: {
	itemId: string;
	type: ItemType;
	status: string | null;
	unreleased?: boolean;
}) {
	const t = useT();
	const l = useLocalized();
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
		if (next === "") router.delete(l(`/items/${itemId}/status`), opts);
		else router.put(l(`/items/${itemId}/status`), { status: next }, opts);
	};

	return (
		<fieldset className="grid gap-2 border-t border-border pt-3">
			<legend className="mb-2 text-xs font-semibold tracking-widest uppercase">
				{t("social.status.label")}
			</legend>
			<div className="grid grid-cols-2 gap-2">
				{STATUSES.map((s) => (
					<button
						key={s}
						type="button"
						aria-pressed={value === s}
						disabled={busy || (unreleased && s !== "want" && value !== s)}
						onClick={() => change(value === s ? "" : s)}
						className={btnClass(value === s ? "primary" : "secondary", "px-2 text-sm")}
					>
						{t(`social.status.${type}.${s}`)}
					</button>
				))}
			</div>
		</fieldset>
	);
}
