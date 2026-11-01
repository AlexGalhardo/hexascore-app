import { useQueries, useQuery } from "@tanstack/react-query";
import type { ReactElement } from "react";
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

/** Entity screen (SDD 10 §6.1): one `/entities/{id}/items?type=` page per type, in `ITEM_TYPES` order. */
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
	const pages = useQueries({
		queries: CATALOG_TYPES.map((type) => ({
			queryKey: queryKeys.entityItems(id, type, locale),
			queryFn: () => read.listEntityItems({ params: { id }, query: { type, locale } }),
		})),
	});

	if (!entity.data) return <View testID="entity" className="flex-1 bg-bg dark:bg-bg-dark" />;
	const data = entity.data;
	const byId = new Map<string, [CatalogType, AppItemJSON]>();
	const groups: EntityGroup[] = CATALOG_TYPES.map((type, index) => {
		const items = pages[index]?.data?.data ?? [];
		for (const item of items) byId.set(item.id, [type, item]);
		const total = pages[index]?.data?.meta?.total ?? items.length;
		return {
			key: type,
			title: t(locale, `nav.${type}`),
			count: t(locale, "entity.showing", { count: items.length, total }),
			rows: items.map((item) => rowOf(item, locale)),
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
