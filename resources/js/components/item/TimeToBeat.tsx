import { router } from "@inertiajs/react";
import { useState } from "react";

import { useT } from "../../lib/useT";
import type { Hours } from "./types";

export interface Medians extends Hours {
	reports: number;
}

const STYLES = [
	["main", "mainHours"],
	["extras", "extrasHours"],
	["completionist", "completionistHours"],
] as const;

interface Props {
	itemId: string;
	medians: Medians;
	mine: Hours | null | undefined;
	// The server accepts reports only while the viewer's status is "completed".
	canReport: boolean;
}

export default function TimeToBeat({ itemId, medians, mine, canReport }: Props) {
	const t = useT();
	const [draft, setDraft] = useState<Record<string, string>>(
		Object.fromEntries(STYLES.map(([, k]) => [k, mine?.[k] != null ? String(mine[k]) : ""])),
	);
	const [busy, setBusy] = useState(false);
	const none = STYLES.every(([, k]) => medians[k] === null);

	const save = () => {
		const body = Object.fromEntries(
			STYLES.map(([, k]) => [k, draft[k] === "" ? null : Number(draft[k])]),
		);
		setBusy(true);
		router.put(`/items/${itemId}/time-to-beat`, body, {
			preserveScroll: true,
			onFinish: () => setBusy(false),
		});
	};

	return (
		<section aria-labelledby="ttb-heading" className="grid gap-2">
			<h2 id="ttb-heading" className="text-xl font-bold">
				{t("item.timeToBeat.title")}
			</h2>
			{none ? (
				<p>{t("item.timeToBeat.notEnough", { min: 3 })}</p>
			) : (
				<dl className="grid grid-cols-2 gap-1">
					{STYLES.map(([style, k]) =>
						medians[k] === null ? null : (
							<div key={style} className="contents">
								<dt>{t(`item.timeToBeat.styles.${style}`)}</dt>
								<dd>{t("item.timeToBeat.hours", { hours: medians[k] })}</dd>
							</div>
						),
					)}
				</dl>
			)}
			<p className="text-sm">{t("item.timeToBeat.reports", { count: medians.reports })}</p>
			{canReport && (
				<form
					className="grid gap-2"
					onSubmit={(e) => {
						e.preventDefault();
						save();
					}}
				>
					<h3 className="font-semibold">{t("item.timeToBeat.yours")}</h3>
					<p className="text-sm">{t("item.timeToBeat.yoursHint")}</p>
					{STYLES.map(([style, k]) => (
						<label key={style} className="grid gap-1">
							<span>{t(`item.timeToBeat.styles.${style}`)}</span>
							<input
								type="number"
								min={0.5}
								max={10000}
								step={0.5}
								className="min-h-11 rounded border px-2"
								placeholder={t("item.timeToBeat.hoursPlaceholder")}
								value={draft[k]}
								onChange={(e) => setDraft((d) => ({ ...d, [k]: e.target.value }))}
							/>
						</label>
					))}
					<div className="flex gap-2">
						<button
							type="submit"
							className="min-h-11 rounded border px-4"
							disabled={busy}
						>
							{t("item.timeToBeat.save")}
						</button>
						{mine && (
							<button
								type="button"
								className="min-h-11 rounded border px-4"
								disabled={busy}
								onClick={() =>
									router.delete(`/items/${itemId}/time-to-beat`, {
										preserveScroll: true,
									})
								}
							>
								{t("item.timeToBeat.delete")}
							</button>
						)}
					</div>
				</form>
			)}
		</section>
	);
}
