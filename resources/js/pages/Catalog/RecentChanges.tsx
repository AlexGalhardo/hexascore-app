import { Link, usePage } from "@inertiajs/react";
import { fieldLabel } from "../../components/community/fields";
import Seo from "../../components/Seo";
import { btnClass } from "../../components/ui/Btn";
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

const th = "py-2 pr-4 text-left text-xs font-bold tracking-[0.1em] text-muted uppercase";
const tab =
	"-mb-px inline-flex min-h-11 items-center border-b-2 px-1 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fg";

export default function RecentChanges({ target, entries, page, nextPage }: Props) {
	const t = useT();
	const { url, props } = usePage<SharedProps>();
	const path = url.split("?")[0];
	const back = path.replace(/\/changes$/, "");
	const date = new Intl.DateTimeFormat(props.locale, { dateStyle: "medium", timeZone: "UTC" });
	const title = t("community.recent.title", { title: target.title });
	// Audit fields without a catalogue label (rare scalar columns) fall back to the raw name.
	const label = (f: string) => {
		const l = fieldLabel(t, f);
		return l.startsWith("community.fields.") ? f : l;
	};

	const actor = (e: Entry) => {
		if (e.actor === "user" && e.nickname) {
			return (
				<Link
					href={localized(`/u/${e.nickname.toLowerCase()}`, props.locale)}
					className="font-medium text-primary hover:underline"
				>
					{e.nickname}
				</Link>
			);
		}
		return t(`community.recent.actors.${e.actor}`);
	};

	return (
		<>
			<Seo title={title} noindex={page > 1} />
			<nav aria-label={t("item.breadcrumb")} className="mb-10 text-sm text-muted">
				<ol className="flex flex-wrap gap-2">
					<li>
						<Link href={back} className="hover:text-fg hover:underline">
							{target.title}
						</Link>
					</li>
					<li aria-hidden="true">/</li>
					<li aria-current="page" className="text-fg">
						{t("community.recent.link")}
					</li>
				</ol>
			</nav>
			<h1 className="text-4xl leading-none md:text-6xl">{title}</h1>
			<p className="mt-3 text-muted">{t("community.recent.lede")}</p>
			<div className="mt-8 flex gap-6 border-b border-border">
				<Link href={back} className={`${tab} border-transparent text-muted hover:text-fg`}>
					{t("community.recent.overview")}
				</Link>
				<span aria-current="page" className={`${tab} border-fg font-semibold`}>
					{t("community.recent.link")}
				</span>
			</div>
			{entries.length === 0 ? (
				<p className="mt-10 text-center font-display text-2xl text-muted italic">
					{t("community.recent.empty")}
				</p>
			) : (
				<table className="mt-8 w-full border-collapse text-sm max-md:block">
					<thead className="border-b-2 border-fg max-md:sr-only">
						<tr>
							<th className={th}>{t("community.recent.columns.date")}</th>
							<th className={th}>{t("community.recent.columns.field")}</th>
							<th className={th}>{t("community.recent.columns.before")}</th>
							<th className={th}>{t("community.recent.columns.after")}</th>
							<th className={th}>{t("community.recent.columns.by")}</th>
						</tr>
					</thead>
					<tbody className="max-md:block max-md:border-t-2 max-md:border-fg">
						{entries.map((e) => (
							<tr
								key={e.id}
								className="border-b border-border align-top max-md:grid max-md:grid-cols-[1fr_auto] max-md:gap-x-4 max-md:gap-y-1 max-md:py-3 [&>td]:md:py-3.5 [&>td]:md:pr-4"
							>
								<td className="text-muted tabular-nums max-md:order-2 max-md:text-xs">
									<time dateTime={e.occurredAt}>
										{date.format(new Date(e.occurredAt))}
									</time>
								</td>
								<td className="font-semibold max-md:order-1">{label(e.field)}</td>
								{e.source === "merge" ? (
									<td colSpan={2} className="max-md:order-3 max-md:col-span-2">
										{t("community.recent.merged")}
									</td>
								) : (
									<>
										<td className="break-words text-muted max-md:order-3 max-md:col-span-2">
											<del className="no-underline">{show(e.before)}</del>
										</td>
										<td className="break-words max-md:order-4 max-md:col-span-2">
											<span className="sr-only">
												{t("community.recent.changedTo")}{" "}
											</span>
											<span aria-hidden="true" className="md:hidden">
												→{" "}
											</span>
											<ins className="no-underline">{show(e.after)}</ins>
										</td>
									</>
								)}
								<td className="max-md:order-5 max-md:col-span-2 max-md:text-xs">
									{actor(e)}
								</td>
							</tr>
						))}
					</tbody>
				</table>
			)}
			<nav
				aria-label={t("community.recent.pages")}
				className="mt-8 flex justify-center gap-4"
			>
				{page > 1 && (
					<Link href={`${path}?page=${page - 1}`} className={btnClass("secondary")}>
						{t("item.previous")}
					</Link>
				)}
				{nextPage && (
					<Link href={`${path}?page=${nextPage}`} className={btnClass("secondary")}>
						{t("item.next")}
					</Link>
				)}
			</nav>
		</>
	);
}
