import { useQueries, useQuery } from "@tanstack/react-query";
import { type ReactElement, useState } from "react";
import { Linking, View } from "react-native";
import type { AppItemJSON, EntityJSON } from "../../../../packages/app-client/src/api";
import type { ReadClient } from "../api/client";
import { formatScore } from "../components/ItemRow";
import { type AppLocale, t } from "../i18n";
import { CATALOG_TYPES, type CatalogType, itemHref } from "../navigation/catalog";
import { queryKeys } from "../query/keys";
import type { ItemRowData } from "../ui/ItemRow";
import { type EntityFact, type EntityGroup, EntityPage } from "../ui/screens/EntityPage";

export type EntityReader = Pick<ReadClient, "getEntity" | "listEntityItems">;

function factsOf(entity: EntityJSON, locale: AppLocale): EntityFact[] {
	const facts: [string, string | number | null | undefined][] = [
		["entity.born", entity.bornOn],
		["entity.died", entity.diedOn],
		["entity.founded", entity.foundedYear],
	];
	return facts
		.filter((fact): fact is [string, string | number] => fact[1] != null && fact[1] !== "")
		.map(([key, value]) => ({ label: t(locale, key), value: String(value) }));
}

function rowOf(item: AppItemJSON, locale: AppLocale): ItemRowData {
	const score = item.score ?? null;
	return {
		id: item.id,
		title: item.title ?? item.slug,
		meta: item.releaseDate?.slice(0, 4) ?? "",
		coverUri: null,
		score,
		scoreLabel:
			score === null
				? t(locale, "item.notEnough")
				: `${t(locale, "item.communityScore")}: ${formatScore(score, locale)}`,
	};
}

/**
 * Entity screen (SDD 10 §6.1): `/entities/{id}/items?type=` per type, in `ITEM_TYPES` order;
 * each group's Next appends the page at its last `meta.next_cursor`.
 */
export function EntityView({
	id,
	read,
	locale,
	onOpen,
	onBrowse,
}: {
	id: string;
	read: EntityReader;
	locale: AppLocale;
	onOpen: (href: ReturnType<typeof itemHref>) => void;
	onBrowse: () => void;
}): ReactElement {
	const entity = useQuery({
		queryKey: queryKeys.entity(id, locale),
		queryFn: () => read.getEntity({ params: { id }, query: { locale } }),
	});
	// Cursors of the extra pages per type; scoped so a new entity or locale restarts at page one.
	const scope = `${id}|${locale}`;
	const [more, setMore] = useState<{
		scope: string;
		cursors: Partial<Record<CatalogType, string[]>>;
	}>({
		scope,
		cursors: {},
	});
	const cursors = more.scope === scope ? more.cursors : {};
	const slots = CATALOG_TYPES.flatMap((type) =>
		[undefined, ...(cursors[type] ?? [])].map((cursor) => ({ type, cursor })),
	);
	const pages = useQueries({
		queries: slots.map(({ type, cursor }) => ({
			queryKey: queryKeys.entityItems(id, type, locale, cursor),
			// Opaque, server-signed cursor (VF-005): passed back as received.
			queryFn: () =>
				read.listEntityItems({
					params: { id },
					query: { type, locale, ...(cursor ? { cursor } : {}) },
				}),
		})),
	});
	const pagesOf = (type: CatalogType) => pages.filter((_, index) => slots[index]?.type === type);
	const nextCursor = (type: CatalogType) =>
		pagesOf(type).at(-1)?.data?.meta?.next_cursor || undefined;

	if (!entity.data) return <View testID="entity" className="flex-1 bg-bg dark:bg-bg-dark" />;
	const data = entity.data;
	const byId = new Map<string, [CatalogType, AppItemJSON]>();
	const groups: EntityGroup[] = CATALOG_TYPES.map((type) => {
		const typePages = pagesOf(type);
		const items = typePages.flatMap((page) => page.data?.data ?? []);
		for (const item of items) byId.set(item.id, [type, item]);
		const total = typePages[0]?.data?.meta?.total ?? items.length;
		return {
			key: type,
			title: t(locale, `nav.${type}`),
			count: t(locale, "entity.showing", { count: items.length, total }),
			rows: items.map((item) => rowOf(item, locale)),
			// Hidden while the next page loads, so a double tap cannot request it twice.
			hasMore: nextCursor(type) !== undefined,
		};
	});
	const description = data.description?.text ?? null;
	const sourceUrl = data.description?.sourceUrl;

	return (
		<EntityPage
			kind={t(locale, `entity.kinds.${data.kind}`)}
			name={data.name ?? data.slug}
			// No role counts in the API yet: the role chips stay hidden.
			roles={[]}
			activeRole={null}
			facts={factsOf(data, locale)}
			description={description}
			licence={t(locale, "entity.descriptionSource", { licence: data.description?.licence ?? "" })}
			sourceLabel={t(locale, "item.source")}
			groups={groups}
			moreLabel={t(locale, "entity.next")}
			onMore={(key) => {
				const type = key as CatalogType;
				const cursor = nextCursor(type);
				if (cursor)
					setMore({ scope, cursors: { ...cursors, [type]: [...(cursors[type] ?? []), cursor] } });
			}}
			empty={{ title: t(locale, "entity.empty"), body: "", action: t(locale, "errors.home") }}
			locale={locale}
			onRole={() => {}}
			onOpen={(row) => {
				const hit = byId.get(row.id);
				if (hit) onOpen(itemHref(hit[0], hit[1]));
			}}
			onSource={() => {
				if (sourceUrl) void Linking.openURL(sourceUrl);
			}}
			onBrowse={onBrowse}
		/>
	);
}
