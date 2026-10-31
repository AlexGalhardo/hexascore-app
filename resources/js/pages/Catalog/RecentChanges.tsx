import { Link, usePage } from "@inertiajs/react";
import Seo from "../../components/Seo";
import { localized } from "../../lib/route";
import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";

interface Entry {
	id: number;
	occurredAt: string;
	source: string;
	field: string;
	before: string;
	after: string;
	actor: "user" | "team" | "deleted" | "system";
	nickname?: string;
}

interface Props {
	target: { type: string; id: string; title: string };
	entries: Entry[];
	page: number;
	nextPage: number | null;
}

// Values are JSON as stored in the audit log; a plain string shows without quotes, anything else as compact JSON.
function show(raw: string): string {
	if (raw === "" || raw === "null") return "—";
	try {
		const v: unknown = JSON.parse(raw);
		return typeof v === "string" ? v : JSON.stringify(v);
	} catch {
		return raw;
	}
}

export default function RecentChanges({ target, entries, page, nextPage }: Props) {
	const t = useT();
	const { url, props } = usePage<SharedProps>();
	const path = url.split("?")[0];
	const back = path.replace(/\/changes$/, "");

	const actor = (e: Entry) => {
		if (e.actor === "user" && e.nickname) {
			return (
				<Link
					href={localized(`/u/${e.nickname.toLowerCase()}`, props.locale)}
					className="underline"
				>
					{e.nickname}
				</Link>
			);
		}
		return t(`community.recent.actors.${e.actor}`);
	};

	return (
		<>
			<Seo title={t("community.recent.title", { title: target.title })} noindex={page > 1} />
			<h1 className="text-2xl font-semibold">
				{t("community.recent.title", { title: target.title })}
			</h1>
			<p className="mt-1 text-sm">
				<Link href={back} className="underline">
					{t("community.recent.back")}
				</Link>
			</p>
			{entries.length === 0 ? (
				<p className="mt-4">{t("community.recent.empty")}</p>
			) : (
				<ol className="mt-4 grid gap-3">
					{entries.map((e) => (
						<li key={e.id} className="border-b border-border pb-3">
							<p className="text-sm text-muted">
								<time dateTime={e.occurredAt}>{e.occurredAt.slice(0, 10)}</time> ·{" "}
								<span>{actor(e)}</span>
							</p>
							<p className="mt-1">
								<span className="font-medium">{e.field}</span>
								{e.source === "merge" ? (
									<> · {t("community.recent.merged")}</>
								) : (
									<>
										{": "}
										<del className="text-muted">{show(e.before)}</del>
										<span aria-hidden="true"> → </span>
										<span className="sr-only">
											{t("community.recent.changedTo")}
										</span>
										<ins className="no-underline">{show(e.after)}</ins>
									</>
								)}
							</p>
						</li>
					))}
				</ol>
			)}
			<nav aria-label={t("community.recent.pages")} className="mt-4 flex gap-4">
				{page > 1 && (
					<Link href={`${path}?page=${page - 1}`} className="underline">
						{t("item.previous")}
					</Link>
				)}
				{nextPage && (
					<Link href={`${path}?page=${nextPage}`} className="underline">
						{t("item.next")}
					</Link>
				)}
			</nav>
		</>
	);
}
