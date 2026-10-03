"use client";

import { Dialog } from "@base-ui/react/dialog";
import { keepPreviousData, useQuery } from "@tanstack/react-query";
import { useTranslations } from "next-intl";
import { useRef, useState } from "react";
import { z } from "zod";
import { ScoreBadge } from "@/components/score/score-badge";
import { DIALOG_BACKDROP, dialogPopup } from "@/components/ui/dialog";
import { Link } from "@/i18n/navigation";
import { abbreviateLabel } from "@/lib/categories";
import { pageWindow } from "@/lib/page-window";

const responseSchema = z.object({
	data: z.array(
		z.object({
			nickname: z.string(),
			overall: z.number(),
			categories: z.array(z.number()).length(6),
			ratedAt: z.string(),
		}),
	),
	page: z.number().int().min(1),
	page_count: z.number().int().min(0),
});
type RatingsPage = z.infer<typeof responseSchema>;

async function fetchRatingsPage(itemId: string, page: number, signal: AbortSignal): Promise<RatingsPage> {
	try {
		const response = await fetch(`/api/items/${itemId}/ratings?page=${page}`, {
			signal: AbortSignal.any([signal, AbortSignal.timeout(10_000)]),
		});
		if (!response.ok) throw new Error(`HTTP ${response.status}`);
		return responseSchema.parse(await response.json());
	} catch (error) {
		if (!signal.aborted) console.error("[ratings] load failed", error);
		throw error;
	}
}

type Props = { itemId: string; categoryLabels: string[] };

const pageButton =
	"inline-flex min-h-11 min-w-11 items-center justify-center rounded-md border border-border px-3 text-sm hover:bg-surface disabled:cursor-not-allowed disabled:opacity-50";

export function RatingsDialog({ itemId, categoryLabels }: Props) {
	const t = useTranslations("item");
	const ts = useTranslations("score");
	const [open, setOpen] = useState(false);
	const [page, setPage] = useState(1);
	const scroller = useRef<HTMLDivElement>(null);
	const query = useQuery({
		queryKey: ["item-ratings", itemId, page],
		queryFn: ({ signal }) => fetchRatingsPage(itemId, page, signal),
		placeholderData: keepPreviousData,
		enabled: open,
	});
	const ratings = query.data?.data ?? [];
	const pageCount = query.data?.page_count ?? 0;
	const busy = query.isFetching;

	function goTo(next: number): void {
		setPage(next);
		scroller.current?.scrollTo({ top: 0 });
	}

	return (
		<Dialog.Root open={open} onOpenChange={setOpen}>
			<Dialog.Trigger className="inline-flex min-h-11 items-center rounded-md border border-border px-4 text-sm font-medium hover:bg-surface">
				{t("seeUserScores")}
			</Dialog.Trigger>
			<Dialog.Portal>
				{/* z-60: above the sticky navbar (z-40) and the cookie banner (z-50) */}
				<Dialog.Backdrop className={DIALOG_BACKDROP} />
				<Dialog.Popup className={`${dialogPopup("lg")} flex flex-col p-4`}>
					<div className="flex items-start justify-between gap-4">
						<div>
							<Dialog.Title className="font-serif text-xl font-bold">{t("userScores")}</Dialog.Title>
							<Dialog.Description className="text-sm text-muted">
								{t("userScoresHint")}
							</Dialog.Description>
						</div>
						<Dialog.Close className="min-h-11 shrink-0 rounded-md border border-border px-3 text-sm">
							{t("close")}
						</Dialog.Close>
					</div>
					<div ref={scroller} className="mt-3 min-h-0 flex-1 overflow-auto" aria-busy={busy}>
						{ratings.length ? (
							<table className={`w-full text-sm ${query.isPlaceholderData ? "opacity-60" : ""}`}>
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
						{query.isSuccess && !ratings.length ? (
							<p className="py-6 text-center text-muted">{t("noPublicRatings")}</p>
						) : null}
						{query.isPending && busy ? <p className="py-3 text-center text-muted">{t("loading")}</p> : null}
						{query.isError ? (
							<div role="alert" className="flex flex-col items-center gap-2 py-3 text-center">
								<p>{t("loadError")}</p>
								<button type="button" onClick={() => void query.refetch()} className={pageButton}>
									{t("retry")}
								</button>
							</div>
						) : null}
					</div>
					{pageCount > 1 ? (
						<nav aria-label={t("ratingsPages")} className="mt-3 flex flex-col items-center gap-2">
							<p className="text-sm text-muted tabular-nums" aria-live="polite">
								{t("pageOf", { page, count: pageCount })}
							</p>
							<div className="flex w-full items-center justify-between gap-2 sm:justify-center">
								<button
									type="button"
									onClick={() => goTo(page - 1)}
									disabled={busy || page <= 1}
									className={pageButton}
								>
									{t("previous")}
								</button>
								{/* numbered pages need ~400px; on phones Previous/Next plus the status line are enough */}
								<ul className="hidden items-center gap-1 sm:flex">
									{pageWindow(page, pageCount).map((slot, i, slots) =>
										slot === "gap" ? (
											<li
												key={`gap-after-${slots[i - 1]}`}
												aria-hidden="true"
												className="px-1 text-muted"
											>
												…
											</li>
										) : (
											<li key={slot}>
												<button
													type="button"
													onClick={() => goTo(slot)}
													disabled={busy && slot !== page}
													aria-current={slot === page ? "page" : undefined}
													aria-label={t("goToPage", { page: slot })}
													className={`${pageButton} ${slot === page ? "bg-fg font-semibold text-bg hover:bg-fg" : ""}`}
												>
													{slot}
												</button>
											</li>
										),
									)}
								</ul>
								<button
									type="button"
									onClick={() => goTo(page + 1)}
									disabled={busy || page >= pageCount}
									className={pageButton}
								>
									{t("next")}
								</button>
							</div>
						</nav>
					) : null}
				</Dialog.Popup>
			</Dialog.Portal>
		</Dialog.Root>
	);
}
