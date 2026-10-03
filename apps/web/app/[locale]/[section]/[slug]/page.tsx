import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { getFormatter, getTranslations, setRequestLocale } from "next-intl/server";
import type { ReactNode } from "react";
import { RatingPanel } from "@/components/item/rating-panel";
import { RatingsDialog } from "@/components/item/ratings-dialog";
import { TimeToBeat } from "@/components/item/time-to-beat";
import { Trailers } from "@/components/item/trailers";
import { NavIcon } from "@/components/layout/nav-icons";
import { HexRadar } from "@/components/score/hex-radar";
import { ScoreBadge } from "@/components/score/score-badge";
import { ScoreBar } from "@/components/score/score-bar";
import { asLocale } from "@/i18n/locale";
import { getPathname, Link } from "@/i18n/navigation";
import { routing } from "@/i18n/routing";
import { getAffiliateLinks, getExternalScores, getItemBySlug } from "@/lib/catalog/cached";
import { buildExternalLinks } from "@/lib/catalog/external-links";
import { CATEGORY_KEYS, MIN_RATINGS_FOR_SCORE } from "@/lib/categories";
import { itemHref, SECTION_ACCENT, SECTION_PATHS, typeFromSection } from "@/lib/sections";

type Props = { params: Promise<{ locale: string; section: string; slug: string }> };

async function load(params: Props["params"]) {
	const { locale, section, slug } = await params;
	const type = typeFromSection(section);
	if (!type || slug.length > 200) return null;
	const item = await getItemBySlug(type, slug, asLocale(locale));
	return item ? { item, locale: asLocale(locale) } : null;
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
	const loaded = await load(params);
	if (!loaded) return {};
	const { item, locale } = loaded;
	const t = await getTranslations({ locale, namespace: "item" });
	const tt = await getTranslations({ locale, namespace: "itemTypes" });
	const title = item.year ? `${item.title} (${item.year})` : item.title;
	const description = (item.synopsis ?? t("metaDescription", { title: item.title, type: tt(item.type) })).slice(
		0,
		200,
	);
	const href = itemHref(item.type, item.slug);
	return {
		title,
		description,
		alternates: {
			canonical: getPathname({ href, locale }),
			languages: Object.fromEntries(routing.locales.map((l) => [l, getPathname({ href, locale: l })])),
		},
		openGraph: {
			title,
			description,
			type: "website",
			url: getPathname({ href, locale }),
			images: item.coverUrl ? [{ url: item.coverUrl }] : undefined,
		},
	};
}

export default async function ItemPage({ params }: Props) {
	const loaded = await load(params);
	if (!loaded) notFound();
	const { item, locale } = loaded;
	setRequestLocale(locale);

	const [scores, affiliates] = await Promise.all([
		getExternalScores(item.id),
		item.type === "book" || item.type === "game" ? getAffiliateLinks(item.id) : Promise.resolve([]),
	]);
	const [t, tc, ts, tt, tcat, format] = await Promise.all([
		getTranslations("item"),
		getTranslations("categories"),
		getTranslations("score"),
		getTranslations("itemTypes"),
		getTranslations("catalog"),
		getFormatter(),
	]);

	const accent = SECTION_ACCENT[item.type];
	const labels = CATEGORY_KEYS[item.type].map((k) => tc(`${item.type}.${k}`));
	const summary = item.categories.some((c) => c !== null)
		? ts("radar", { summary: labels.map((l, i) => `${l} ${item.categories[i]}`).join(", ") })
		: ts("radarEmpty");
	const external = buildExternalLinks(item.type, item.externalIds, scores);
	const { meta } = item;
	const details: [string, ReactNode][] = [
		[t("fields.type"), tt(item.type)],
		[
			t("fields.releaseDate"),
			item.releaseDate
				? format.dateTime(new Date(`${item.releaseDate}T00:00:00Z`), { dateStyle: "long", timeZone: "UTC" })
				: null,
		],
		[t("fields.genres"), meta.genres?.join(", ")],
		[t("fields.platforms"), meta.platforms?.join(", ")],
		[t("fields.creators"), meta.creators?.join(", ")],
		[t("fields.runtime"), meta.runtimeMinutes ? t("runtimeValue", { minutes: meta.runtimeMinutes }) : null],
		[t("fields.pages"), meta.pages],
		[t("fields.seasons"), meta.seasons],
		[t("fields.episodes"), meta.episodes],
		[t("fields.countries"), meta.countries?.join(", ")],
	];

	return (
		<article className="grid gap-8">
			{/* Cover + title + final score first: the item is recognisable at a glance on any screen size. */}
			<header className="grid grid-cols-[auto_minmax(0,1fr)] items-start gap-4 sm:gap-6">
				{item.coverUrl ? (
					// biome-ignore lint/performance/noImgElement: remote covers from ingestion; next/image needs a host allowlist (later)
					<img
						src={item.coverUrl}
						alt={t("coverAlt", { title: item.title })}
						width={160}
						height={240}
						fetchPriority="high"
						decoding="async"
						className="aspect-[2/3] w-28 rounded-md border border-border object-cover shadow-sm sm:w-40"
					/>
				) : (
					<div
						aria-hidden="true"
						className={`grid aspect-[2/3] w-28 place-items-center rounded-md border border-border bg-surface sm:w-40 ${accent.text}`}
					>
						<NavIcon name={SECTION_PATHS[item.type]} className="size-10" />
					</div>
				)}
				<div className="grid min-w-0 content-start gap-2">
					<Link
						href={`/${SECTION_PATHS[item.type]}`}
						className={`justify-self-start text-xs font-semibold uppercase tracking-wide ${accent.text}`}
					>
						{tcat(`title.${item.type}`)}
					</Link>
					<h1 className="font-serif text-3xl font-bold leading-tight break-words sm:text-5xl">
						{item.title}
					</h1>
					<p className="text-sm text-muted">
						{[item.year, tt(item.type), meta.genres?.join(", "), meta.platforms?.join(", ")]
							.filter(Boolean)
							.join(" · ")}
					</p>
					{meta.creators?.length ? <p className="text-sm">{meta.creators.join(", ")}</p> : null}
					<div className="mt-2 flex items-center gap-4">
						<ScoreBadge
							score={item.score}
							size="lg"
							label={item.score === null ? ts("none") : ts("outOf", { score: item.score })}
						/>
						<div>
							<p className="text-sm font-semibold">{t("communityScore")}</p>
							<p className="text-sm text-muted">{tcat("ratingCount", { count: item.ratingCount })}</p>
							{item.score === null ? (
								<p className="text-sm text-muted">
									{t("notEnough")} — {t("notEnoughHint", { min: MIN_RATINGS_FOR_SCORE })}
								</p>
							) : null}
						</div>
					</div>
				</div>
			</header>

			<section
				aria-label={t("communityScore")}
				className="grid gap-6 md:grid-cols-[minmax(0,1fr)_minmax(0,1.4fr)]"
			>
				<div className="grid content-start gap-4">
					<div className="flex flex-wrap gap-2">
						<RatingsDialog itemId={item.id} categoryLabels={labels} />
					</div>
					<div className="grid gap-2">
						<h2 className="text-sm font-semibold">{t("categoryScores")}</h2>
						{labels.map((label, i) => (
							<ScoreBar key={label} label={label} value={item.categories[i] ?? null} />
						))}
					</div>
				</div>
				<HexRadar
					values={item.categories}
					labels={labels}
					summary={summary}
					accentClass={accent.text}
					className="justify-self-center"
				/>
			</section>

			{item.timeToBeat ? <TimeToBeat stats={item.timeToBeat} /> : null}

			<RatingPanel
				itemId={item.id}
				type={item.type}
				labels={labels}
				path={getPathname({ href: itemHref(item.type, item.slug), locale })}
			/>

			{external.length ? (
				<section aria-labelledby="external" className="grid gap-2">
					<h2 id="external" className="font-serif text-xl font-bold">
						{t("externalScores")}
					</h2>
					<ul className="flex flex-wrap gap-2">
						{external.map((link) => (
							<li key={link.source}>
								<a
									href={link.url}
									target="_blank"
									rel="noopener noreferrer nofollow"
									className="inline-flex min-h-11 items-center gap-2 rounded-md border border-border px-3 text-sm hover:bg-surface"
								>
									{link.value ? (
										<span className="font-mono font-bold tabular-nums">{link.value}</span>
									) : null}
									<span>{t(`sources.${link.source}`)}</span>
									<span aria-hidden="true">↗</span>
								</a>
							</li>
						))}
					</ul>
					<p className="text-xs text-muted">{t("externalNotice")}</p>
				</section>
			) : null}

			<section aria-labelledby="synopsis" className="grid max-w-3xl gap-2">
				<h2 id="synopsis" className="font-serif text-xl font-bold">
					{t("synopsis")}
				</h2>
				{item.translationFallback ? (
					<p className="rounded-md border border-border bg-surface px-3 py-2 text-sm">
						{t("fallbackNotice")}
					</p>
				) : null}
				<p className="leading-relaxed" lang={item.translationFallback ? "en" : undefined}>
					{item.synopsis ?? t("noSynopsis")}
				</p>
				{item.synopsis && (item.synopsisLicense || item.synopsisSourceUrl) ? (
					<p className="text-xs text-muted">
						{item.synopsisLicense ? t("license", { license: item.synopsisLicense }) : null}
						{item.synopsisSourceUrl ? (
							<>
								{item.synopsisLicense ? " · " : null}
								<a
									href={item.synopsisSourceUrl}
									rel="noopener noreferrer nofollow"
									className="underline"
								>
									{t("source")}
								</a>
							</>
						) : null}
					</p>
				) : null}
			</section>

			{meta.trailers.length ? (
				<section aria-labelledby="trailers" className="grid gap-2">
					<h2 id="trailers" className="font-serif text-xl font-bold">
						{t("trailers")}
					</h2>
					<Trailers trailers={meta.trailers.slice(0, 3)} />
				</section>
			) : null}

			<section aria-labelledby="details" className="grid max-w-3xl gap-2">
				<h2 id="details" className="font-serif text-xl font-bold">
					{t("details")}
				</h2>
				<dl className="grid grid-cols-[minmax(0,10rem)_1fr] border-t border-border text-sm">
					{details
						.filter(([, value]) => value !== null && value !== undefined && value !== "")
						.map(([label, value]) => (
							<div key={label} className="contents">
								<dt className="border-b border-border py-2 pr-3 text-muted">{label}</dt>
								<dd className="border-b border-border py-2">{value}</dd>
							</div>
						))}
				</dl>
				{item.coverUrl && meta.coverAttribution ? (
					<p className="text-xs text-muted">
						{t("coverCredit", {
							credit: [
								meta.coverAttribution.source,
								meta.coverAttribution.author,
								meta.coverAttribution.license,
							]
								.filter(Boolean)
								.join(" · "),
						})}
						{meta.coverAttribution.url ? (
							<>
								{" · "}
								<a
									href={meta.coverAttribution.url}
									rel="noopener noreferrer nofollow"
									className="underline"
								>
									{t("source")}
								</a>
							</>
						) : null}
					</p>
				) : null}
			</section>

			{affiliates.length ? (
				<section aria-labelledby="buy" className="grid gap-2">
					<h2 id="buy" className="font-serif text-xl font-bold">
						{t("whereToBuy")}
					</h2>
					<ul className="flex flex-wrap gap-2">
						{affiliates.map((link) => (
							<li key={link.store}>
								<a
									href={link.url}
									target="_blank"
									rel="sponsored nofollow noopener"
									className={`inline-flex min-h-11 items-center rounded-md border-2 px-4 text-sm font-medium ${accent.border}`}
								>
									{t(`stores.${link.store}`)}
								</a>
							</li>
						))}
					</ul>
					<p className="text-xs text-muted">{t("affiliateNotice")}</p>
				</section>
			) : null}
		</article>
	);
}
