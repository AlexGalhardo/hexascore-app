import ForYouListView, { type ForYouList } from "../../components/catalog/ForYouList";
import Seo from "../../components/Seo";
import { useT } from "../../lib/useT";

const TYPES = ["game", "movie", "series", "book"] as const;

type Props = { optedOut: boolean } & { [K in (typeof TYPES)[number]]?: ForYouList };

export default function Index(props: Props) {
	const t = useT();

	return (
		<>
			<Seo title={t("forYou.title")} noindex />
			<h1 className="text-2xl font-semibold">{t("forYou.title")}</h1>
			{props.optedOut && <p className="mt-4">{t("forYou.optedOut")}</p>}
			{!props.optedOut &&
				TYPES.map((type) => {
					const list = props[type];
					return (
						<section key={type} aria-labelledby={`foryou-${type}`} className="mt-8">
							<h2
								id={`foryou-${type}`}
								className="border-b-2 border-border pb-1 text-xl font-bold"
							>
								{t(`catalog.title.${type}`)}
							</h2>
							{!list ? (
								<p className="mt-2 text-muted">{t("forYou.loading")}</p>
							) : (
								<>
									{list.basis === "onboarding" && (
										<p className="mt-2">
											{t("forYou.onboarding", {
												type: t(`itemTypes.${type}`),
											})}
										</p>
									)}
									{list.data.length === 0 && list.basis === "personal" ? (
										<p className="mt-2">{t("forYou.empty")}</p>
									) : (
										<ForYouListView type={type} list={list} />
									)}
								</>
							)}
						</section>
					);
				})}
		</>
	);
}
