import { Dialog as BaseDialog } from "@base-ui/react/dialog";
import { useEffect, useRef, useState } from "react";
import { abbreviateLabel, pageWindow } from "../../lib/pageWindow";
import { useLocalized } from "../../lib/useLocalized";
import { useT } from "../../lib/useT";
import { dialogBackdrop, dialogPopup } from "../Dialog";

type Row = {
	nickname: string;
	overall: number;
	categories: number[];
	ratedAt: string;
};

type RatingsPage = { data: Row[]; page: number; page_count: number };

function isRatingsPage(value: unknown): value is RatingsPage {
	if (!value || typeof value !== "object") return false;
	const body = value as { data?: unknown; page?: unknown; page_count?: unknown };
	return (
		Array.isArray(body.data) &&
		typeof body.page === "number" &&
		typeof body.page_count === "number"
	);
}

const pageButton =
	"inline-flex min-h-11 min-w-11 items-center justify-center rounded-md border border-border px-3 text-sm hover:bg-surface disabled:cursor-not-allowed disabled:opacity-50";

export default function RatingsDialog({
	itemId,
	categoryLabels,
}: {
	itemId: string;
	categoryLabels: string[];
}) {
	const t = useT();
	const l = useLocalized();
	const [open, setOpen] = useState(false);
	const [page, setPage] = useState(1);
	const [rows, setRows] = useState<Row[]>([]);
	const [pageCount, setPageCount] = useState(0);
	const [status, setStatus] = useState<"idle" | "loading" | "ready" | "error">("idle");
	const [attempt, setAttempt] = useState(0);
	const scroller = useRef<HTMLDivElement>(null);
	const busy = status === "loading";

	useEffect(() => {
		if (!open) return;
		const ctrl = new AbortController();
		const timer = setTimeout(() => ctrl.abort(), 10_000);
		let ignore = false;
		setStatus("loading");
		fetch(`/api/items/${itemId}/ratings?page=${page}&r=${attempt}`, { signal: ctrl.signal })
			.then(async (response) => {
				if (!response.ok) throw new Error(`HTTP ${response.status}`);
				const body: unknown = await response.json();
				if (!isRatingsPage(body)) throw new Error("shape");
				if (ignore) return;
				setRows(body.data);
				setPageCount(body.page_count);
				setStatus("ready");
			})
			.catch(() => {
				if (!ignore) setStatus("error");
			});
		return () => {
			ignore = true;
			clearTimeout(timer);
			ctrl.abort();
		};
	}, [open, page, itemId, attempt]);

	function goTo(next: number): void {
		setPage(next);
		scroller.current?.scrollTo({ top: 0 });
	}

	return (
		<BaseDialog.Root open={open} onOpenChange={setOpen}>
			<BaseDialog.Trigger className="inline-flex min-h-11 items-center rounded-md border border-border px-4 text-sm font-medium hover:bg-surface">
				{t("item.seeUserScores")}
			</BaseDialog.Trigger>
			<BaseDialog.Portal>
				<BaseDialog.Backdrop className={dialogBackdrop} />
				<BaseDialog.Popup
					className={`${dialogPopup("lg")} flex max-h-[min(40rem,calc(100%-2rem))] flex-col`}
				>
					<div className="flex items-start justify-between gap-4">
						<div>
							<BaseDialog.Title className="font-serif text-xl font-bold">
								{t("item.userScores")}
							</BaseDialog.Title>
							<BaseDialog.Description className="text-sm text-muted">
								{t("item.userScoresHint")}
							</BaseDialog.Description>
						</div>
						<BaseDialog.Close className="min-h-11 shrink-0 rounded-md border border-border px-3 text-sm">
							{t("item.close")}
						</BaseDialog.Close>
					</div>
					<div
						ref={scroller}
						className="mt-3 min-h-0 flex-1 overflow-auto"
						aria-busy={busy}
					>
						{rows.length > 0 ? (
							<table className={`w-full text-sm ${busy ? "opacity-60" : ""}`}>
								<thead className="sticky top-0 bg-bg text-left text-xs text-muted">
									<tr>
										<th className="py-2 pr-2 font-medium">{t("item.user")}</th>
										<th className="py-2 pr-2 font-medium">
											{t("item.overall")}
										</th>
										{categoryLabels.map((label) => (
											<th
												key={label}
												className="py-2 pr-2 text-right font-medium"
											>
												<abbr title={label} className="no-underline">
													{abbreviateLabel(label)}
												</abbr>
											</th>
										))}
									</tr>
								</thead>
								<tbody>
									{rows.map((row) => (
										<tr
											key={`${row.nickname}-${row.ratedAt}`}
											className="border-t border-border"
										>
											<td className="max-w-32 truncate py-2 pr-2">
												<a
													href={l(
														`/u/${encodeURIComponent(row.nickname)}`,
													)}
													className="underline underline-offset-2"
												>
													{row.nickname}
												</a>
											</td>
											<td className="py-2 pr-2 font-mono tabular-nums">
												{row.overall}
											</td>
											{row.categories.map((value, i) => (
												<td
													key={categoryLabels[i] ?? `category-${value}`}
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
						{status === "ready" && rows.length === 0 ? (
							<p className="py-6 text-center text-muted">
								{t("item.noPublicRatings")}
							</p>
						) : null}
						{status === "loading" && rows.length === 0 ? (
							<p className="py-3 text-center text-muted">{t("item.loading")}</p>
						) : null}
						{status === "error" ? (
							<div
								role="alert"
								className="flex flex-col items-center gap-2 py-3 text-center"
							>
								<p>{t("item.loadError")}</p>
								<button
									type="button"
									onClick={() => setAttempt((n) => n + 1)}
									className={pageButton}
								>
									{t("item.retry")}
								</button>
							</div>
						) : null}
					</div>
					{pageCount > 1 ? (
						<nav
							aria-label={t("item.ratingsPages")}
							className="mt-3 flex flex-col items-center gap-2"
						>
							<p className="text-sm text-muted tabular-nums" aria-live="polite">
								{t("item.pageOf", { page, count: pageCount })}
							</p>
							<div className="flex w-full items-center justify-between gap-2 sm:justify-center">
								<button
									type="button"
									onClick={() => goTo(page - 1)}
									disabled={busy || page <= 1}
									className={pageButton}
								>
									{t("item.previous")}
								</button>
								<ul className="hidden items-center gap-1 sm:flex">
									{pageWindow(page, pageCount).map((slot, i, slots) =>
										slot === "gap" ? (
											<li
												key={`gap-after-${String(slots[i - 1])}`}
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
													aria-current={
														slot === page ? "page" : undefined
													}
													aria-label={t("item.goToPage", { page: slot })}
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
									{t("item.next")}
								</button>
							</div>
						</nav>
					) : null}
				</BaseDialog.Popup>
			</BaseDialog.Portal>
		</BaseDialog.Root>
	);
}
