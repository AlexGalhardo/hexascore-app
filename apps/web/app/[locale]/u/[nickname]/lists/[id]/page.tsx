import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { getTranslations, setRequestLocale } from "next-intl/server";
import { cache } from "react";
import { ListForm } from "@/components/lists/list-form";
import { DeleteListButton, ListEntryControls } from "@/components/lists/list-owner-controls";
import { ReportDialog } from "@/components/social/report-dialog";
import { Notice } from "@/components/ui/form";
import { asLocale } from "@/i18n/locale";
import { Link } from "@/i18n/navigation";
import type { Locale } from "@/i18n/routing";
import { db } from "@/lib/db/client";
import { itemHref } from "@/lib/sections";
import { getSessionUser } from "@/lib/session";
import { getList } from "@/lib/social/lists";

type Props = { params: Promise<{ locale: string; nickname: string; id: string }> };

// The list must belong to the nickname in the URL; otherwise (or when not visible) it is a 404.
const load = cache(async (nickname: string, id: string, locale: Locale) => {
	const me = await getSessionUser();
	const list = await getList(db, id, me?.id ?? null, locale);
	if (!list || list.ownerNickname.toLowerCase() !== decodeURIComponent(nickname).toLowerCase()) return null;
	return { list, me };
});

export async function generateMetadata({ params }: Props): Promise<Metadata> {
	const { locale, nickname, id } = await params;
	const loaded = await load(nickname, id, asLocale(locale));
	if (!loaded) return {};
	const { list } = loaded;
	return {
		title: `${list.title} · ${list.ownerNickname}`,
		description: list.description ?? undefined,
		robots: list.isPublic && !list.hidden ? undefined : { index: false },
	};
}

export default async function ListPage({ params }: Props) {
	const { locale, nickname, id } = await params;
	setRequestLocale(locale);
	const loaded = await load(nickname, id, asLocale(locale));
	if (!loaded) notFound();
	const { list, me } = loaded;
	const isOwner = me?.id === list.ownerId;
	const [t, tt] = await Promise.all([getTranslations("lists"), getTranslations("itemTypes")]);

	return (
		<article className="grid max-w-3xl gap-6">
			<header className="grid gap-2">
				<Link
					href={{ pathname: `/u/${encodeURIComponent(list.ownerNickname)}`, query: { tab: "lists" } }}
					className="text-sm text-muted hover:underline"
				>
					{t("by", { nickname: list.ownerNickname })}
				</Link>
				<h1 className="font-serif text-3xl font-bold break-words">{list.title}</h1>
				{list.description ? <p className="whitespace-pre-line">{list.description}</p> : null}
				<p className="text-sm text-muted">
					{[t("itemCount", { count: list.items.length }), list.isPublic ? t("public") : t("private")].join(
						" · ",
					)}
				</p>
				{isOwner && list.hidden ? <Notice tone="error">{t("hiddenNotice")}</Notice> : null}
				{me && !isOwner ? <ReportDialog target={{ type: "list", id: list.id }} /> : null}
			</header>

			{list.items.length ? (
				<ol className="grid">
					{list.items.map((item, i) => (
						<li
							key={item.id}
							className="flex flex-wrap items-center gap-x-3 gap-y-1 border-t border-border py-2"
						>
							<span className="w-8 font-mono text-sm tabular-nums text-muted">{i + 1}</span>
							<div className="min-w-0 flex-1">
								<Link href={itemHref(item.type, item.slug)} className="block truncate hover:underline">
									{item.title}
								</Link>
								<span className="text-xs text-muted">{tt(item.type)}</span>
							</div>
							{isOwner ? (
								<ListEntryControls
									listId={list.id}
									itemId={item.id}
									title={item.title}
									first={i === 0}
									last={i === list.items.length - 1}
								/>
							) : null}
						</li>
					))}
				</ol>
			) : (
				<p className="text-muted">{isOwner ? t("emptyOwner") : t("empty")}</p>
			)}

			{isOwner ? (
				<section aria-labelledby="edit-list" className="grid gap-4 border-t border-border pt-6">
					<h2 id="edit-list" className="font-serif text-xl font-bold">
						{t("edit")}
					</h2>
					<ListForm
						listId={list.id}
						initial={{ title: list.title, description: list.description ?? "", isPublic: list.isPublic }}
					/>
					<DeleteListButton listId={list.id} nickname={list.ownerNickname} />
				</section>
			) : null}
		</article>
	);
}
