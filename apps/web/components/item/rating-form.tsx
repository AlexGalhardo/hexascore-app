"use client";

import { useTranslations } from "next-intl";
import { useState, useTransition } from "react";
import { deleteRating, saveRating } from "@/actions/ratings";
import { ScoreBadge } from "@/components/score/score-badge";
import { ScoreInput } from "@/components/score/score-input";
import { buttonClass, Notice } from "@/components/ui/form";
import type { ActionError } from "@/lib/action-result";
import { overall } from "@/lib/categories";

type Props = { itemId: string; labels: string[]; initial: number[] | null };
type Feedback = { tone: "success"; key: "saved" | "deleted" } | { tone: "error"; error: ActionError } | null;

const DEFAULT = [50, 50, 50, 50, 50, 50];

export function RatingForm({ itemId, labels, initial }: Props) {
	const t = useTranslations("ratings");
	const ts = useTranslations("score");
	const tAction = useTranslations("actions");
	const [values, setValues] = useState<number[]>(initial ?? DEFAULT);
	const [saved, setSaved] = useState<number[] | null>(initial);
	const [feedback, setFeedback] = useState<Feedback>(null);
	const [pending, startTransition] = useTransition();
	const live = overall(...values);
	const dirty = !saved || saved.some((v, i) => v !== values[i]);

	function onSave(): void {
		const previous = saved;
		setSaved(values); // optimistic: the panel shows "saved" state while the server confirms
		setFeedback(null);
		startTransition(async () => {
			const result = await saveRating({ itemId, scores: values });
			if (result.ok) setFeedback({ tone: "success", key: "saved" });
			else {
				setSaved(previous);
				setFeedback({ tone: "error", error: result.error });
			}
		});
	}

	function onDelete(): void {
		const previous = saved;
		setSaved(null);
		setFeedback(null);
		startTransition(async () => {
			const result = await deleteRating(itemId);
			if (result.ok) {
				setValues(DEFAULT);
				setFeedback({ tone: "success", key: "deleted" });
			} else {
				setSaved(previous);
				setFeedback({ tone: "error", error: result.error });
			}
		});
	}

	return (
		<form
			className="grid gap-4"
			onSubmit={(e) => {
				e.preventDefault();
				onSave();
			}}
		>
			<div className="flex items-center gap-3">
				<ScoreBadge score={live} size="md" label={ts("outOf", { score: live })} />
				<div>
					<p className="text-sm font-semibold">{t("yourOverall")}</p>
					<p className="text-sm text-muted">{t("overallHint")}</p>
				</div>
			</div>
			<div className="grid gap-3 sm:grid-cols-2">
				{labels.map((label, i) => (
					<ScoreInput
						key={label}
						label={label}
						value={values[i] ?? 0}
						disabled={pending}
						onChange={(v) => {
							setFeedback(null);
							setValues((prev) => prev.map((old, j) => (j === i ? v : old)));
						}}
					/>
				))}
			</div>
			<div aria-live="polite">
				{feedback?.tone === "error" ? <Notice tone="error">{tAction(feedback.error)}</Notice> : null}
				{feedback?.tone === "success" ? <Notice tone="success">{t(feedback.key)}</Notice> : null}
				{pending ? <p className="text-sm text-muted">{t("saving")}</p> : null}
			</div>
			<div className="flex flex-wrap gap-2">
				<button type="submit" disabled={pending || !dirty} className={buttonClass("primary")}>
					{saved ? t("update") : t("save")}
				</button>
				{saved ? (
					<button type="button" disabled={pending} onClick={onDelete} className={buttonClass("secondary")}>
						{t("delete")}
					</button>
				) : null}
			</div>
		</form>
	);
}
