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

// Plain text links (owner 2026-11-04): no boxes around the services.
const LINK =
	"inline-flex min-h-8 items-center text-[15px] text-blue-700 dark:text-blue-400 underline underline-offset-4 hover:no-underline focus-visible:rounded-sm focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fg";

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

/** Where to watch (films and series) or where to play (games), the same worldwide list for every visitor. Renders nothing without offers. */
export default function Offers({
	offers,
	type,
}: {
	offers: OffersData;
	type: "game" | "movie" | "series" | "book";
}): ReactElement | null {
	const t = useT();
	const list = worldwideOffers(offers);
	if (list.length === 0) return null;
	const kinds = [...new Set(list.map((o) => o.kind))];

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
					<ul className="flex flex-wrap gap-x-5 gap-y-1">
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

			{offers.commission && <p className="text-sm text-muted">{t("offers.commission")}</p>}
			{offers.amazonAssociate && <p className="text-sm text-muted">{t("offers.amazon")}</p>}
			{offers.attribution.includes("JustWatch") && (
				<p className="text-sm text-muted">{t("offers.justwatch")}</p>
			)}
		</div>
	);
}
