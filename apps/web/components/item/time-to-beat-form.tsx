"use client";

import { useTranslations } from "next-intl";
import { useId, useState, useTransition } from "react";
import { deleteTimeToBeat, saveTimeToBeat } from "@/actions/time-to-beat";
import { buttonClass, Field, Input, Notice } from "@/components/ui/form";
import type { ActionError } from "@/lib/action-result";
import { TIME_TO_BEAT_HOURS, TIME_TO_BEAT_STYLES, type TimeToBeatStyle } from "@/lib/social/constants";

type Values = Record<TimeToBeatStyle, number | null>;
type Props = { itemId: string; initial: Values | null };
type Feedback = { tone: "success"; key: "saved" | "deleted" } | { tone: "error"; error: ActionError } | null;

const EMPTY: Record<TimeToBeatStyle, string> = { main: "", extras: "", completionist: "" };
const toText = (v: Values | null): Record<TimeToBeatStyle, string> =>
	v
		? {
				main: v.main?.toString() ?? "",
				extras: v.extras?.toString() ?? "",
				completionist: v.completionist?.toString() ?? "",
			}
		: EMPTY;

/** Shown in the rating panel for games the viewer marked as completed. Empty fields are not reported. */
export function TimeToBeatForm({ itemId, initial }: Props) {
	const t = useTranslations("item.timeToBeat");
	const tAction = useTranslations("actions");
	const id = useId();
	const [text, setText] = useState(toText(initial));
	const [saved, setSaved] = useState(initial !== null);
	const [feedback, setFeedback] = useState<Feedback>(null);
	const [pending, startTransition] = useTransition();

	function onSave(): void {
		const hours = (s: string): number | null => (s.trim() === "" ? null : Number(s));
		const input = {
			itemId,
			main: hours(text.main),
			extras: hours(text.extras),
			completionist: hours(text.completionist),
		};
		setFeedback(null);
		startTransition(async () => {
			const result = await saveTimeToBeat(input);
			if (result.ok) {
				setSaved(true);
				setFeedback({ tone: "success", key: "saved" });
			} else setFeedback({ tone: "error", error: result.error });
		});
	}

	function onDelete(): void {
		setFeedback(null);
		startTransition(async () => {
			const result = await deleteTimeToBeat(itemId);
			if (result.ok) {
				setSaved(false);
				setText(EMPTY);
				setFeedback({ tone: "success", key: "deleted" });
			} else setFeedback({ tone: "error", error: result.error });
		});
	}

	return (
		<form
			aria-labelledby={`${id}-title`}
			className="grid gap-3 border-t border-border pt-4"
			onSubmit={(e) => {
				e.preventDefault();
				onSave();
			}}
		>
			<div>
				<h3 id={`${id}-title`} className="text-sm font-semibold">
					{t("yours")}
				</h3>
				<p className="text-sm text-muted">{t("yoursHint")}</p>
			</div>
			<div className="grid gap-3 sm:grid-cols-3">
				{TIME_TO_BEAT_STYLES.map((style) => (
					<Field key={style} id={`${id}-${style}`} label={t(`styles.${style}`)}>
						<Input
							id={`${id}-${style}`}
							type="number"
							inputMode="decimal"
							min={TIME_TO_BEAT_HOURS.min}
							max={TIME_TO_BEAT_HOURS.max}
							step={TIME_TO_BEAT_HOURS.step}
							placeholder={t("hoursPlaceholder")}
							value={text[style]}
							disabled={pending}
							onChange={(e) => {
								setFeedback(null);
								setText((prev) => ({ ...prev, [style]: e.target.value }));
							}}
						/>
					</Field>
				))}
			</div>
			<div aria-live="polite">
				{feedback?.tone === "error" ? <Notice tone="error">{tAction(feedback.error)}</Notice> : null}
				{feedback?.tone === "success" ? <Notice tone="success">{t(feedback.key)}</Notice> : null}
			</div>
			<div className="flex flex-wrap gap-2">
				<button
					type="submit"
					disabled={pending || Object.values(text).every((v) => v.trim() === "")}
					className={buttonClass("primary")}
				>
					{t("save")}
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
