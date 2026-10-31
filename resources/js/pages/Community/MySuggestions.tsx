import { Link } from "@inertiajs/react";
import Seo from "../../components/Seo";
import { useLocalized } from "../../lib/useLocalized";
import { useT } from "../../lib/useT";

interface Item {
	id: string;
	kind: string;
	domain: string;
	status: string;
	title: string | null;
	updatedAt: string;
}

export default function MySuggestions({ suggestions }: { suggestions: Item[]; open: boolean }) {
	const t = useT();
	const l = useLocalized();

	return (
		<>
			<Seo title={t("community.mine.title")} noindex />
			<h1 className="text-2xl font-semibold">{t("community.mine.title")}</h1>
			{suggestions.length === 0 ? (
				<p className="mt-3">{t("community.mine.empty")}</p>
			) : (
				<ul className="mt-4 space-y-2">
					{suggestions.map((s) => (
						<li key={s.id}>
							<Link href={l(`/suggestions/${s.id}`)} className="underline">
								{s.title ?? t(`community.mine.kinds.${s.kind}`)}
							</Link>{" "}
							<span className="text-sm text-muted">
								{t(`community.mine.kinds.${s.kind}`)} · {s.domain} ·{" "}
								{t(`community.status.states.${s.status}`)} ·{" "}
								{new Date(s.updatedAt).toLocaleDateString()}
							</span>
						</li>
					))}
				</ul>
			)}
		</>
	);
}
