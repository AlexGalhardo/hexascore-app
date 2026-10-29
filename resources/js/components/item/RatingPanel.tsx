import { router } from "@inertiajs/react";
import { useState } from "react";

import { useT } from "../../lib/useT";
import StarRating from "../rating/StarRating";
import { CATEGORY_KEYS, type ItemType, type RatingSummary } from "./types";

interface Props {
	itemId: string;
	type: ItemType;
	summary: RatingSummary;
	mine: Record<string, number> | null | undefined;
	canWrite: boolean;
}

// Community aggregate plus the viewer's own six-category rating. Local values
// are optimistic; the aggregate always comes from the server reload.
export default function RatingPanel({ itemId, type, summary, mine, canWrite }: Props) {
	const t = useT();
	const keys = CATEGORY_KEYS[type];
	const [values, setValues] = useState<Record<string, number | null>>(
		Object.fromEntries(keys.map((k) => [k, mine?.[k] ?? null])),
	);
	const [busy, setBusy] = useState(false);
	const complete = keys.every((k) => values[k] !== null);

	const visit = { preserveScroll: true, onFinish: () => setBusy(false) };
	const save = () => {
		setBusy(true);
		router.put(`/items/${itemId}/rating`, { scores: values }, visit);
	};
	const remove = () => {
		setBusy(true);
		setValues(Object.fromEntries(keys.map((k) => [k, null])));
		router.delete(`/items/${itemId}/rating`, visit);
	};

	return (
		<section aria-labelledby="rating-heading" className="grid gap-4">
			<h2 id="rating-heading" className="text-xl font-bold">
				{t("item.communityScore")}
			</h2>
			{summary.score === null ? (
				<p>
					<strong>{t("item.notEnough")}</strong>{" "}
					<span>{t("item.notEnoughHint", { min: 5 })}</span>
				</p>
			) : (
				<>
					<p className="text-4xl font-bold">
						{summary.score}
						<span className="sr-only">
							{" "}
							{t("rating.valuetext", { value: summary.score })}
						</span>
					</p>
					<h3 className="sr-only">{t("item.categoryScores")}</h3>
					<dl className="grid grid-cols-2 gap-1">
						{keys.map((k) => (
							<div key={k} className="contents">
								<dt>{t(`categories.${type}.${k}`)}</dt>
								<dd>{summary.scoreByCategory[k] ?? "-"}</dd>
							</div>
						))}
					</dl>
				</>
			)}
			<h3 className="font-semibold">{t("ratings.title")}</h3>
			{canWrite ? (
				<>
					{keys.map((k) => (
						<StarRating
							key={k}
							label={t(`categories.${type}.${k}`)}
							value={values[k]}
							onChange={(v) => setValues((s) => ({ ...s, [k]: v }))}
							disabled={busy}
						/>
					))}
					<div className="flex gap-2">
						<button
							type="button"
							className="min-h-11 rounded border px-4"
							disabled={!complete || busy}
							onClick={save}
						>
							{busy
								? t("ratings.saving")
								: mine
									? t("ratings.update")
									: t("ratings.save")}
						</button>
						{mine && (
							<button
								type="button"
								className="min-h-11 rounded border px-4"
								disabled={busy}
								onClick={remove}
							>
								{t("ratings.delete")}
							</button>
						)}
					</div>
				</>
			) : (
				<p>{t("item.loginToRate")}</p>
			)}
		</section>
	);
}
