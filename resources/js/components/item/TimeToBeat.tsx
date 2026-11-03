import { Link, router } from "@inertiajs/react";
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
	["speedrun", "speedrunHours"],
] as const;

interface Props {
	itemId: string;
	medians: Medians;
	mine: Hours | null | undefined;
	canEdit: boolean;
}

export default function TimeToBeat({ itemId, medians, mine, canEdit }: Props) {
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
		<section aria-labelledby="ttb-heading" className="grid content-start gap-3">
			<h2 id="ttb-heading" className="text-2xl font-bold">
				{t("item.timeToBeat.title")}
			</h2>
			{none ? (
				<p>{t("item.timeToBeat.none")}</p>
			) : (
				<table className="w-full text-sm">
					<thead>
						<tr className="border-b border-border text-left text-muted">
							{STYLES.map(([style]) => (
								<th key={style} scope="col" className="py-2.5 pr-3 font-semibold">
									{t(`item.timeToBeat.styles.${style}`)}
								</th>
							))}
						</tr>
					</thead>
					<tbody>
						<tr>
							{STYLES.map(([style, k]) => (
								<td key={style} className="py-2.5 pr-3 tabular-nums">
									{medians[k] === null
										? "–"
										: t("item.timeToBeat.hours", { hours: medians[k] })}
								</td>
							))}
						</tr>
					</tbody>
				</table>
			)}
			{!none && (
				<p className="text-xs text-muted">
					{t("item.timeToBeat.hint")}{" "}
					{t("item.timeToBeat.reports", { count: medians.reports })}
				</p>
			)}
			{canEdit ? (
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
					<div className="grid grid-cols-2 gap-2 sm:grid-cols-4">
						{STYLES.map(([style, k]) => (
							<label
								key={style}
								className="grid content-end gap-1 text-xs text-muted"
							>
								<span>{t(`item.timeToBeat.styles.${style}`)}</span>
								<input
									type="number"
									min={0.5}
									max={1000}
									step={0.5}
									className="min-h-11 w-full rounded-md border border-field bg-field-bg px-2 text-sm text-fg"
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
			) : (
				<Link href={l("/login")} className="text-sm text-primary hover:underline">
					{t("item.timeToBeat.loginToReport")}
				</Link>
			)}
		</section>
	);
}
