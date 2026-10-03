import type { Metadata } from "next";
import { getTranslations, setRequestLocale } from "next-intl/server";
import { ApiConsole } from "@/components/api/api-console";
import { UsageChart } from "@/components/api/usage-chart";
import { Link, redirect } from "@/i18n/navigation";
import { getQuotaUsed, getUsageSeries, listApiKeys } from "@/lib/api/dashboard";
import { PLAN_LIMITS } from "@/lib/api/plans";
import { getStore } from "@/lib/api/quota";
import { getBillingSummary } from "@/lib/billing/summary";
import { db } from "@/lib/db/client";
import { getSessionUser } from "@/lib/session";

type Props = { params: Promise<{ locale: string }> };

export async function generateMetadata({ params }: Props): Promise<Metadata> {
	const { locale } = await params;
	const t = await getTranslations({ locale, namespace: "api.profile" });
	return { title: t("metaTitle"), robots: { index: false } };
}

export default async function ProfileApiPage({ params }: Props) {
	const { locale } = await params;
	setRequestLocale(locale);
	const user = await getSessionUser();
	if (!user) {
		redirect({ href: { pathname: "/login", query: { next: "/profile/api" } }, locale });
		return null;
	}

	const store = getStore();
	const summary = await getBillingSummary(user.id);
	const [t, tp, keys, series, used] = await Promise.all([
		getTranslations("api.profile"),
		getTranslations("billing.plans"),
		listApiKeys(db, user.id),
		getUsageSeries(db, store, user.id),
		getQuotaUsed(store, user.id, summary.plan),
	]);
	const limits = PLAN_LIMITS[summary.plan];
	const nf = new Intl.NumberFormat(locale);
	const df = new Intl.DateTimeFormat(locale, { day: "numeric", month: "short", timeZone: "UTC" });
	const total = series.reduce((sum, d) => sum + d.count, 0);
	const today = series.at(-1)?.count ?? 0;
	const now = new Date();
	const resetsAt = new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth() + 1, 1));

	return (
		<section className="max-w-4xl space-y-10">
			<header className="space-y-2">
				<h1 className="font-serif text-3xl font-bold sm:text-4xl">{t("title")}</h1>
				<p className="text-muted">
					{t("intro")}{" "}
					<Link href="/docs" className="text-fg underline underline-offset-4">
						{t("docsLink")}
					</Link>
				</p>
			</header>

			<ApiConsole keys={keys}>
				<div className="grid gap-6 md:grid-cols-[1fr_2fr]">
					<section aria-labelledby="plan-title" className="space-y-2 rounded-lg border border-border p-5">
						<h2 id="plan-title" className="text-sm text-muted">
							{t("plan.title")}
						</h2>
						<p className="font-serif text-2xl font-bold">{tp(summary.plan)}</p>
						<p className="font-mono text-sm tabular-nums">
							{t("plan.quota", { used: nf.format(used), limit: nf.format(limits.quota) })}
						</p>
						<p className="text-xs text-muted">
							{limits.period === "lifetime"
								? t("plan.lifetime")
								: t("plan.resets", {
										date: new Intl.DateTimeFormat(locale, {
											dateStyle: "long",
											timeZone: "UTC",
										}).format(resetsAt),
									})}
						</p>
						<ul className="space-y-1 pt-2 text-sm">
							<li>{t("plan.rate", { count: limits.requestsPerSecond })}</li>
							<li>{t("plan.pageSize", { count: limits.maxPageSize })}</li>
							{summary.plan !== "free" && (
								<li>{t("plan.credits", { count: nf.format(summary.credits) })}</li>
							)}
						</ul>
						<div className="flex flex-wrap gap-3 pt-3 text-sm">
							<Link href="/pricing" className="underline underline-offset-4">
								{t("plan.upgrade")}
							</Link>
							<Link href="/profile/api/checkout" className="underline underline-offset-4">
								{t("plan.checkout")}
							</Link>
						</div>
					</section>

					<section aria-labelledby="usage-title" className="space-y-3 rounded-lg border border-border p-5">
						<h2 id="usage-title" className="text-sm text-muted">
							{t("usage.title")}
						</h2>
						<dl className="flex gap-8">
							<div>
								<dt className="text-xs text-muted">{t("usage.today")}</dt>
								<dd className="font-mono text-2xl tabular-nums">{nf.format(today)}</dd>
							</div>
							<div>
								<dt className="text-xs text-muted">{t("usage.last30")}</dt>
								<dd className="font-mono text-2xl tabular-nums">{nf.format(total)}</dd>
							</div>
						</dl>
						<UsageChart
							series={series}
							label={t("usage.chartLabel", { total: nf.format(total) })}
							barLabel={(d) =>
								t("usage.bar", {
									day: df.format(new Date(`${d.day}T00:00:00Z`)),
									count: nf.format(d.count),
								})
							}
						/>
						<p className="text-xs text-muted">{t("usage.note")}</p>
					</section>
				</div>
			</ApiConsole>
		</section>
	);
}
