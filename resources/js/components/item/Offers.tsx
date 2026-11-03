import { type ChangeEvent, type ReactElement, useEffect, useState } from "react";

import { useT } from "../../lib/useT";

export interface OfferView {
	id: string;
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

const STORAGE_KEY = "offers-country";
const LINK =
	"inline-flex min-h-11 items-center rounded-md border border-border bg-surface px-3 text-[15px] font-semibold hover:bg-fg hover:text-bg focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fg";
const SELECT =
	"min-h-11 rounded-md border border-field bg-field-bg px-3 text-[15px] text-fg focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-primary";

/** The first country named by the URL, then the one remembered, that the page offers. */
function chosenCountry(countries: string[]): string | null {
	const fromUrl = new URLSearchParams(window.location.search).get("country");
	let remembered: string | null = null;
	try {
		remembered = localStorage.getItem(STORAGE_KEY);
	} catch {
		/* storage blocked: the URL and the default still work */
	}
	return [fromUrl, remembered].find((c) => c !== null && countries.includes(c)) ?? null;
}

/**
 * Where to watch or buy (RFC-003 section 4). The server sends every country, so the cached HTML is the same for all
 * visitors; the first render uses the locale's country and an effect applies the URL or the remembered choice.
 */
export default function Offers({
	offers,
	type,
	title,
}: {
	offers: OffersData;
	type: "game" | "movie" | "series" | "book";
	title: string;
	/** Reserved for date formatting; names come from the catalogues so the server and client markup match. */
	locale?: string;
}): ReactElement {
	const t = useT();
	const [country, setCountry] = useState(offers.defaultCountry);

	useEffect(() => {
		const picked = chosenCountry(offers.countries);
		if (picked) setCountry(picked);
	}, [offers.countries]);

	const change = (e: ChangeEvent<HTMLSelectElement>) => {
		const next = e.target.value;
		setCountry(next);
		try {
			localStorage.setItem(STORAGE_KEY, next);
		} catch {
			/* storage blocked */
		}
		const url = new URL(window.location.href);
		url.searchParams.set("country", next);
		window.history.replaceState(window.history.state, "", url);
	};

	const list = offers.byCountry[country] ?? [];
	const kinds = [...new Set(list.map((o) => o.kind))];
	const canSearch = type === "movie" || type === "series";

	return (
		<div className="grid gap-4">
			<label className="flex flex-wrap items-center gap-3 text-sm font-semibold">
				{t("offers.country")}
				<select value={country} onChange={change} className={SELECT}>
					{offers.countries.map((c) => (
						<option key={c} value={c}>
							{t(`offers.countryNames.${c}`)}
						</option>
					))}
				</select>
			</label>

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
					{t("offers.none", { country: t(`offers.countryNames.${country}`) })}{" "}
					{canSearch && (
						<a
							href={`https://www.justwatch.com/${country.toLowerCase()}/search?q=${encodeURIComponent(title)}`}
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
