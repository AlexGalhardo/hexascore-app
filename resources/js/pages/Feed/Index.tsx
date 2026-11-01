import { Link } from "@inertiajs/react";
import Seo from "../../components/Seo";
import ActivityCard, { type ActivityEntry } from "../../components/social/ActivityCard";
import { btnClass } from "../../components/ui/Btn";
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
			<h1 className="mt-6 font-display text-5xl leading-none font-bold tracking-[-0.02em] md:mt-10 md:text-[64px]">
				{t("social.feed.title")}
			</h1>
			{events.length === 0 ? (
				<p className="mt-8 max-w-3xl border-t-2 border-fg pt-6 font-display text-2xl italic">
					{t("social.feed.empty")}
				</p>
			) : (
				<ul className="mt-8 max-w-3xl divide-y divide-border border-t-2 border-fg">
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
					className={btnClass("secondary", "mt-6")}
					preserveScroll
				>
					{t("social.feed.more")}
				</Link>
			)}
		</>
	);
}
