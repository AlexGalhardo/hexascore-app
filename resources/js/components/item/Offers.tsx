import type { ReactElement } from "react";

import { useT } from "../../lib/useT";

export interface OfferView {
	id: string;
	store?: string;
	name: string;
	kind: string;
	/** Day (UTC, YYYY-MM-DD) the offer was last confirmed. */
	verifiedAt?: string;
	/** The store pays HexaScore a commission on this link. */
	sponsored: boolean;
}

/** The `offers` prop of Catalog/Show (app/offers.Section). The target URL never reaches the page: links go through /go/{id}. */
export interface OffersData {
	countries: string[];
	defaultCountry: string;
	byCountry: Record<string, OfferView[]>;
	commission: boolean;
	amazonAssociate: boolean;
	attribution: string[];
}

const LINK =
	"inline-flex min-h-11 items-center rounded-md border border-border bg-surface px-3 text-[15px] font-semibold hover:bg-fg hover:text-bg focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fg";

// Same order as the server (app/offers kindOrder): subscriptions and free first, physical last.
const KIND_ORDER = ["stream", "free", "rent", "buy", "buy_digital", "mobile", "physical", "read"];

/**
 * Worldwide view of the offers: every country's offers are merged, so a title on a service in one country shows
 * for every visitor. A store and kind appear once; the default country's link wins when several have it.
 */
export function worldwideOffers(offers: OffersData): OfferView[] {
	const order = [
		offers.defaultCountry,
		...offers.countries.filter((c) => c !== offers.defaultCountry),
	];
	const seen = new Map<string, OfferView>();
	for (const country of order) {
		for (const offer of offers.byCountry[country] ?? []) {
			const key = `${offer.kind}|${offer.store ?? offer.name}`;
			if (!seen.has(key)) seen.set(key, offer);
		}
	}
	return [...seen.values()].sort(
		(a, b) => KIND_ORDER.indexOf(a.kind) - KIND_ORDER.indexOf(b.kind),
	);
}

/** Where to watch (films and series) or where to play (games), the same worldwide list for every visitor. */
export default function Offers({
	offers,
	type,
	title,
}: {
	offers: OffersData;
	type: "game" | "movie" | "series" | "book";
	title: string;
}): ReactElement {
	const t = useT();
	const list = worldwideOffers(offers);
	const kinds = [...new Set(list.map((o) => o.kind))];
	const canSearch = type === "movie" || type === "series";

	return (
		<div className="grid gap-4">
			<h3 className="text-xs font-semibold tracking-widest text-muted uppercase">
				{t(type === "game" ? "offers.titlePlay" : "offers.titleWatch")}
			</h3>

			{kinds.map((kind) => (
				<fieldset key={kind} className="m-0 grid min-w-0 gap-2 border-0 p-0">
					<legend className="mb-2 p-0 text-base font-bold">
						{t(`offers.kinds.${kind}`)}
					</legend>
					<ul className="flex flex-wrap gap-2">
						{list
							.filter((o) => o.kind === kind)
							.map((o) => (
								<li key={o.id}>
									<a
										href={`/go/${o.id}`}
										target="_blank"
										rel="sponsored nofollow noopener noreferrer"
										title={
											o.verifiedAt
												? t("offers.verified", { date: o.verifiedAt })
												: undefined
										}
										className={LINK}
									>
										{o.name}
										<span className="sr-only"> {t("offers.newTab")}</span>
									</a>
								</li>
							))}
					</ul>
				</fieldset>
			))}

			{list.length === 0 && (
				<p className="text-muted">
					{t("offers.noneWorldwide")}{" "}
					{canSearch && (
						<a
							href={`https://www.justwatch.com/${offers.defaultCountry.toLowerCase()}/search?q=${encodeURIComponent(title)}`}
							rel="noopener noreferrer nofollow"
							className="text-primary hover:underline"
						>
							{t("offers.searchJustWatch")}
						</a>
					)}
				</p>
			)}

			{offers.commission && <p className="text-sm text-muted">{t("offers.commission")}</p>}
			{offers.amazonAssociate && <p className="text-sm text-muted">{t("offers.amazon")}</p>}
			{offers.attribution.includes("JustWatch") && (
				<p className="text-sm text-muted">{t("offers.justwatch")}</p>
			)}
		</div>
	);
}
