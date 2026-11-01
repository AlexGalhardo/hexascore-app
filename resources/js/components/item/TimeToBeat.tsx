import { router } from "@inertiajs/react";
import { useState } from "react";

import { useLocalized } from "../../lib/useLocalized";
import { useT } from "../../lib/useT";
import Btn from "../ui/Btn";
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
	const l = useLocalized();
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
		router.put(l(`/items/${itemId}/time-to-beat`), body, {
			preserveScroll: true,
			onFinish: () => setBusy(false),
		});
	};

	return (
		<section
			aria-labelledby="ttb-heading"
			className="grid content-start gap-3 border-t-2 border-fg pt-4"
		>
			<h2 id="ttb-heading" className="text-2xl">
				{t("item.timeToBeat.title")}
			</h2>
			{none ? (
				<p>{t("item.timeToBeat.notEnough", { min: 3 })}</p>
			) : (
				<dl className="grid">
					{STYLES.map(([style, k]) =>
						medians[k] === null ? null : (
							<div
								key={style}
								className="grid grid-cols-[8rem_1fr] gap-4 border-b border-border py-2.5 text-sm"
							>
								<dt className="text-muted">
									{t(`item.timeToBeat.styles.${style}`)}
								</dt>
								<dd className="tabular-nums">
									{t("item.timeToBeat.hours", { hours: medians[k] })}
								</dd>
							</div>
						),
					)}
				</dl>
			)}
			<p className="text-xs text-muted">
				{t("item.timeToBeat.hint")}{" "}
				{t("item.timeToBeat.reports", { count: medians.reports })}
			</p>
			{canReport && (
				<form
					className="grid gap-2"
					onSubmit={(e) => {
						e.preventDefault();
						save();
					}}
				>
					<h3 className="text-xs font-semibold tracking-widest uppercase">
						{t("item.timeToBeat.yours")}
					</h3>
					<p className="text-xs text-muted">{t("item.timeToBeat.yoursHint")}</p>
					<div className="grid grid-cols-3 gap-2">
						{STYLES.map(([style, k]) => (
							<label
								key={style}
								className="grid content-end gap-1 text-xs text-muted"
							>
								<span>{t(`item.timeToBeat.styles.${style}`)}</span>
								<input
									type="number"
									min={0.5}
									max={10000}
									step={0.5}
									className="min-h-11 w-full rounded-xs border border-field bg-field-bg px-2 text-sm text-fg"
									placeholder={t("item.timeToBeat.hoursPlaceholder")}
									value={draft[k]}
									onChange={(e) =>
										setDraft((d) => ({ ...d, [k]: e.target.value }))
									}
								/>
							</label>
						))}
					</div>
					<div className="flex flex-wrap gap-2">
						<Btn type="submit" kind="secondary" disabled={busy}>
							{t("item.timeToBeat.save")}
						</Btn>
						{mine && (
							<Btn
								kind="ghost"
								disabled={busy}
								onClick={() =>
									router.delete(l(`/items/${itemId}/time-to-beat`), {
										preserveScroll: true,
									})
								}
							>
								{t("item.timeToBeat.delete")}
							</Btn>
						)}
					</div>
				</form>
			)}
		</section>
	);
}
