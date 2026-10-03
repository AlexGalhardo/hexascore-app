import type { Metadata } from "next";
import { getFormatter, getTranslations, setRequestLocale } from "next-intl/server";
import { z } from "zod";
import { CategoryLabel } from "@/components/catalog/category-label";
import { CoverThumb } from "@/components/catalog/cover-thumb";
import { ScoreBadge } from "@/components/score/score-badge";
import { buttonClass } from "@/components/ui/form";
import { asLocale } from "@/i18n/locale";
import { getPathname, Link, redirect } from "@/i18n/navigation";
import { db } from "@/lib/db/client";
import { ITEM_TYPES } from "@/lib/db/schema/catalog";
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
	const [t, tcat, ts, format] = await Promise.all([
		getTranslations("social.feed"),
		getTranslations("catalog"),
		getTranslations("score"),
		getFormatter(),
	]);
	const feed = await getFeed(db, me.id, { cursor, locale: asLocale(locale) }).catch((error: unknown) => {
		// A tampered cursor restarts from the newest entries instead of failing the page.
		if (error instanceof Error && error.name === "InvalidCursorError")
			return getFeed(db, me.id, { locale: asLocale(locale) });
		throw error;
	});
	const suggestions = feed.data.length || cursor ? [] : await getTopRaters(db, me.id);
	// Listing pattern: each page of the (chronological, cursor-paginated) feed is split by category; "older"
	// still pages the whole feed, so a category group may be empty on one page and come back on the next.
	const groups = ITEM_TYPES.map((type) => ({
		type,
		entries: feed.data.filter((e) => e.item.type === type),
	})).filter((g) => g.entries.length);

	return (
		<div className="mx-auto grid max-w-2xl grid-cols-1 gap-6">
			<h1 className="font-serif text-3xl font-bold">{t("title")}</h1>
			{groups.length ? (
				groups.map((group) => (
					<section key={group.type} aria-labelledby={`feed-${group.type}`} className="grid gap-1">
						<h2 id={`feed-${group.type}`} className="font-serif text-lg font-bold">
							<CategoryLabel type={group.type} size="lg">
								{tcat(`title.${group.type}`)}
							</CategoryLabel>
						</h2>
						<ol className="grid">
							{group.entries.map((entry) => (
								<li
									key={`${entry.nickname}-${entry.item.id}`}
									className="flex items-center gap-3 border-t border-border py-2"
								>
									<CoverThumb src={entry.item.coverUrl} thumb={entry.item.coverThumbUrl} />
									<div className="min-w-0 flex-1">
										<p className="truncate">
											<Link
												href={`/u/${encodeURIComponent(entry.nickname)}`}
												className="font-semibold"
											>
												{entry.nickname}
											</Link>{" "}
											{t("rated")}{" "}
											<Link href={itemHref(entry.item.type, entry.item.slug)}>
												{entry.item.title}
											</Link>
										</p>
										<p className="text-xs text-muted">
											{format.dateTime(new Date(entry.ratedAt), {
												dateStyle: "medium",
												timeStyle: "short",
											})}
										</p>
									</div>
									<ScoreBadge
										score={entry.overall}
										size="sm"
										label={ts("outOf", { score: entry.overall })}
									/>
								</li>
							))}
						</ol>
					</section>
				))
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
								<Link href={`/u/${encodeURIComponent(s.nickname)}`} className="font-medium">
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
