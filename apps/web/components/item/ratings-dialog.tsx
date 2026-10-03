"use client";

import { Dialog } from "@base-ui/react/dialog";
import { useTranslations } from "next-intl";
import { useState } from "react";
import { z } from "zod";
import { ScoreBadge } from "@/components/score/score-badge";
import { Link } from "@/i18n/navigation";
import { abbreviateLabel } from "@/lib/categories";

const responseSchema = z.object({
	data: z.array(
		z.object({
			nickname: z.string(),
			overall: z.number(),
			categories: z.array(z.number()).length(6),
			ratedAt: z.string(),
		}),
	),
	next_cursor: z.string().nullable(),
});
type Rating = z.infer<typeof responseSchema>["data"][number];

type Props = { itemId: string; categoryLabels: string[] };

export function RatingsDialog({ itemId, categoryLabels }: Props) {
	const t = useTranslations("item");
	const ts = useTranslations("score");
	const [ratings, setRatings] = useState<Rating[]>([]);
	const [cursor, setCursor] = useState<string | null>(null);
	const [state, setState] = useState<"idle" | "loading" | "error" | "done">("idle");

	async function load(next: string | null): Promise<void> {
		setState("loading");
		try {
			const url = `/api/items/${itemId}/ratings${next ? `?cursor=${encodeURIComponent(next)}` : ""}`;
			const response = await fetch(url, { signal: AbortSignal.timeout(10_000) });
			if (!response.ok) throw new Error(`HTTP ${response.status}`);
			const body = responseSchema.parse(await response.json());
			setRatings((prev) => (next ? [...prev, ...body.data] : body.data));
			setCursor(body.next_cursor);
			setState("done");
		} catch (error) {
			console.error("[ratings] load failed", error);
			setState("error");
		}
	}

	return (
		<Dialog.Root
			onOpenChange={(open) => {
				if (open && state === "idle") void load(null);
			}}
		>
			<Dialog.Trigger className="inline-flex min-h-11 items-center rounded-md border border-border px-4 text-sm font-medium hover:bg-surface">
				{t("seeUserScores")}
			</Dialog.Trigger>
			<Dialog.Portal>
				<Dialog.Backdrop className="fixed inset-0 bg-black/40" />
				<Dialog.Popup className="fixed inset-x-2 top-[5dvh] mx-auto flex max-h-[90dvh] max-w-3xl flex-col rounded-lg border border-border bg-bg p-4 text-fg shadow-xl">
					<div className="flex items-start justify-between gap-4">
						<div>
							<Dialog.Title className="font-serif text-xl font-bold">{t("userScores")}</Dialog.Title>
							<Dialog.Description className="text-sm text-muted">
								{t("userScoresHint")}
							</Dialog.Description>
						</div>
						<Dialog.Close className="min-h-11 rounded-md border border-border px-3 text-sm">
							{t("close")}
						</Dialog.Close>
					</div>
					<div className="mt-3 overflow-auto" aria-live="polite" aria-busy={state === "loading"}>
						{ratings.length ? (
							<table className="w-full text-sm">
								<thead className="sticky top-0 bg-bg text-left text-xs text-muted">
									<tr>
										<th className="py-2 pr-2 font-medium">{t("user")}</th>
										<th className="py-2 pr-2 font-medium">{t("overall")}</th>
										{categoryLabels.map((label) => (
											<th key={label} className="py-2 pr-2 text-right font-medium">
												<abbr title={label} className="no-underline">
													{abbreviateLabel(label)}
												</abbr>
											</th>
										))}
									</tr>
								</thead>
								<tbody>
									{ratings.map((r) => (
										<tr key={`${r.nickname}-${r.ratedAt}`} className="border-t border-border">
											<td className="max-w-32 truncate py-2 pr-2">
												<Link
													href={`/u/${encodeURIComponent(r.nickname)}`}
													className="underline underline-offset-2"
												>
													{r.nickname}
												</Link>
											</td>
											<td className="py-2 pr-2">
												<ScoreBadge
													score={r.overall}
													label={ts("outOf", { score: r.overall })}
													size="sm"
												/>
											</td>
											{r.categories.map((value, i) => (
												<td
													key={categoryLabels[i]}
													className="py-2 pr-2 text-right font-mono tabular-nums"
												>
													{value}
												</td>
											))}
										</tr>
									))}
								</tbody>
							</table>
						) : null}
						{state === "done" && !ratings.length ? (
							<p className="py-6 text-center text-muted">{t("noPublicRatings")}</p>
						) : null}
						{state === "loading" ? <p className="py-3 text-center text-muted">{t("loading")}</p> : null}
						{state === "error" ? (
							<p role="alert" className="py-3 text-center">
								{t("loadError")}
							</p>
						) : null}
					</div>
					{(cursor && state === "done") || state === "error" ? (
						<button
							type="button"
							onClick={() => void load(state === "error" && !ratings.length ? null : cursor)}
							className="mt-3 min-h-11 self-center rounded-md border border-border px-4 text-sm"
						>
							{t("loadMore")}
						</button>
					) : null}
				</Dialog.Popup>
			</Dialog.Portal>
		</Dialog.Root>
	);
}
