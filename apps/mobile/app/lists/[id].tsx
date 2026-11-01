import { router, useLocalSearchParams } from "expo-router";
import type { ReactElement } from "react";
import { listsDeps } from "../../src/auth/runtime";
import { ListDetailScreen } from "../../src/lists/ListDetail";
import { CATALOG_TYPES, firstParam, itemHref, SINGULAR } from "../../src/navigation/catalog";
import { useLocale } from "../../src/providers/root";

export default function ListScreen(): ReactElement | null {
	const { id } = useLocalSearchParams<{ id?: string | string[] }>();
	const locale = useLocale();
	return (
		<ListDetailScreen
			deps={listsDeps}
			id={firstParam(id)}
			locale={locale}
			onOpenItem={({ item }) => {
				// List rows carry the singular catalogue type; routes use the plural segment.
				const type = CATALOG_TYPES.find((p) => p === item.type || SINGULAR[p] === item.type);
				if (type) router.push(itemHref(type, item));
			}}
			onDeleted={() => router.back()}
		/>
	);
}
