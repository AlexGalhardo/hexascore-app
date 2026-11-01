import { Link } from "@inertiajs/react";
import Seo from "../../components/Seo";
import ActivityCard, { type ActivityEntry } from "../../components/social/ActivityCard";
import { useT } from "../../lib/useT";

interface Props {
	events: ActivityEntry[];
	nextCursor: string;
}

export default function Index({ events, nextCursor }: Props) {
	const t = useT();

	return (
		<>
			<Seo title={t("social.feed.title")} noindex />
			<h1 className="text-2xl font-semibold">{t("social.feed.title")}</h1>
			{events.length === 0 ? (
				<p className="mt-4">{t("social.feed.empty")}</p>
			) : (
				<ul className="mt-4 divide-y divide-border">
					{events.map((e) => (
						<li key={e.id}>
							<ActivityCard entry={e} />
						</li>
					))}
				</ul>
			)}
			{nextCursor && (
				<Link
					href={`?cursor=${nextCursor}`}
					className="mt-4 inline-block underline"
					preserveScroll
				>
					{t("social.feed.more")}
				</Link>
			)}
		</>
	);
}
