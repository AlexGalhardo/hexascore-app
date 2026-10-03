import type { Metadata } from "next";
import { getFormatter, getTranslations, setRequestLocale } from "next-intl/server";
import { z } from "zod";
import { ScoreBadge } from "@/components/score/score-badge";
import { buttonClass } from "@/components/ui/form";
import { getPathname, Link, redirect } from "@/i18n/navigation";
import type { Locale } from "@/i18n/routing";
import { db } from "@/lib/db/client";
import { itemHref } from "@/lib/sections";
import { getSessionUser } from "@/lib/session";
import { getFeed, getTopRaters } from "@/lib/social/repository";

type Props = {
	params: Promise<{ locale: string }>;
	searchParams: Promise<Record<string, string | string[] | undefined>>;
};

export async function generateMetadata({ params }: Props): Promise<Metadata> {
	const { locale } = await params;
	const t = await getTranslations({ locale, namespace: "social.feed" });
	return { title: t("title"), robots: { index: false } };
}

const cursorSchema = z.string().min(1).max(512).optional().catch(undefined);

export default async function FeedPage({ params, searchParams }: Props) {
	const { locale } = await params;
	setRequestLocale(locale);
	const me = await getSessionUser();
	const here = getPathname({ href: "/feed", locale });
	if (!me) return redirect({ href: { pathname: "/login", query: { next: here } }, locale });
	if (!me.username) return redirect({ href: { pathname: "/onboarding", query: { next: here } }, locale });

	const cursor = cursorSchema.parse((await searchParams).cursor);
	const [t, tt, ts, format] = await Promise.all([
		getTranslations("social.feed"),
		getTranslations("itemTypes"),
		getTranslations("score"),
		getFormatter(),
	]);
	const feed = await getFeed(db, me.id, { cursor, locale: locale as Locale }).catch((error: unknown) => {
		// A tampered cursor restarts from the newest entries instead of failing the page.
		if (error instanceof Error && error.name === "InvalidCursorError")
			return getFeed(db, me.id, { locale: locale as Locale });
		throw error;
	});
	const suggestions = feed.data.length || cursor ? [] : await getTopRaters(db, me.id);

	return (
		<div className="mx-auto grid max-w-2xl gap-6">
			<h1 className="font-serif text-3xl font-bold">{t("title")}</h1>
			{feed.data.length ? (
				<ol className="grid">
					{feed.data.map((entry) => (
						<li
							key={`${entry.nickname}-${entry.item.id}`}
							className="grid grid-cols-[auto_1fr] items-center gap-3 border-t border-border py-3"
						>
							<ScoreBadge score={entry.overall} size="sm" label={ts("outOf", { score: entry.overall })} />
							<div className="min-w-0">
								<p className="truncate">
									<Link
										href={`/u/${encodeURIComponent(entry.nickname)}`}
										className="font-semibold hover:underline"
									>
										{entry.nickname}
									</Link>{" "}
									{t("rated")}{" "}
									<Link href={itemHref(entry.item.type, entry.item.slug)} className="hover:underline">
										{entry.item.title}
									</Link>
								</p>
								<p className="text-xs text-muted">
									{tt(entry.item.type)} ·{" "}
									{format.dateTime(new Date(entry.ratedAt), {
										dateStyle: "medium",
										timeStyle: "short",
									})}
								</p>
							</div>
						</li>
					))}
				</ol>
			) : (
				<p className="text-muted">{cursor ? t("end") : t("empty")}</p>
			)}
			{feed.nextCursor ? (
				<Link
					href={{ pathname: "/feed", query: { cursor: feed.nextCursor } }}
					className={buttonClass("secondary", "w-fit")}
				>
					{t("older")}
				</Link>
			) : null}
			{suggestions.length ? (
				<section aria-labelledby="suggestions" className="grid gap-2">
					<h2 id="suggestions" className="font-serif text-xl font-bold">
						{t("suggestions")}
					</h2>
					<ul className="grid">
						{suggestions.map((s) => (
							<li key={s.nickname} className="flex min-h-11 items-center gap-3 border-t border-border">
								<Link
									href={`/u/${encodeURIComponent(s.nickname)}`}
									className="font-medium hover:underline"
								>
									{s.nickname}
								</Link>
								<span className="ml-auto text-sm text-muted">
									{t("ratingCount", { count: s.ratings })}
								</span>
							</li>
						))}
					</ul>
				</section>
			) : null}
		</div>
	);
}
