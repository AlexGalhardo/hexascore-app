import { getFormatter, getTranslations } from "next-intl/server";
import type { TimeToBeat as TimeToBeatStats } from "@/lib/catalog/types";
import { MIN_REPORTS_FOR_TIME_TO_BEAT, TIME_TO_BEAT_STYLES } from "@/lib/social/constants";

/** Community median per play style (games). Medians are null below MIN_REPORTS_FOR_TIME_TO_BEAT reports. */
export async function TimeToBeat({ stats }: { stats: TimeToBeatStats }) {
	const [t, format] = await Promise.all([getTranslations("item.timeToBeat"), getFormatter()]);
	return (
		<section aria-labelledby="time-to-beat" className="grid max-w-3xl gap-2">
			<h2 id="time-to-beat" className="font-serif text-xl font-bold">
				{t("title")}
			</h2>
			<dl className="grid gap-2 sm:grid-cols-3">
				{TIME_TO_BEAT_STYLES.map((style) => {
					const { medianHours, reports } = stats[style];
					return (
						<div key={style} className="grid gap-1 rounded-md border border-border p-3">
							<dt className="text-sm text-muted">{t(`styles.${style}`)}</dt>
							<dd className="grid">
								{medianHours === null ? (
									<span className="text-sm">
										{t("notEnough", { min: MIN_REPORTS_FOR_TIME_TO_BEAT })}
									</span>
								) : (
									<span className="font-mono text-2xl font-bold tabular-nums">
										{t("hours", {
											hours: format.number(medianHours, { maximumFractionDigits: 1 }),
										})}
									</span>
								)}
								<span className="text-xs text-muted">{t("reports", { count: reports })}</span>
							</dd>
						</div>
					);
				})}
			</dl>
			<p className="text-xs text-muted">{t("hint")}</p>
		</section>
	);
}
