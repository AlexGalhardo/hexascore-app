import { router, useLocalSearchParams } from "expo-router";
import type { ReactElement } from "react";
import { appClient } from "../../src/auth/runtime";
import { t } from "../../src/i18n";
import { firstParam, isCatalogType } from "../../src/navigation/catalog";
import { useLocale } from "../../src/providers/root";
import { ItemView } from "../../src/screens/item/Item";
import { Placeholder } from "../../src/ui/Placeholder";

const UUID = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

// GET /{type}/{id} takes the item id only. In-app links add `?id=`; a bare slug (a web App Link) has no
// lookup route yet (FIX-app-item-by-slug), so it shows not-found instead of guessing.
export function itemIdOf(slug: string, id: string): string | null {
	if (UUID.test(id)) return id;
	if (UUID.test(slug)) return slug;
	return null;
}

export function ItemBody({
	type,
	slug,
	id = "",
}: {
	type: string;
	slug: string;
	id?: string;
}): ReactElement {
	const locale = useLocale();
	if (!isCatalogType(type)) return <Placeholder title="Unknown section" />;
	const itemId = itemIdOf(slug, id);
	if (itemId === null) return <Placeholder title={t(locale, "errors.notFoundTitle")} />;
	return (
		<ItemView
			type={type}
			id={itemId}
			read={appClient.read}
			locale={locale}
			// No typed-routes manifest is generated, so setParams is typed `never` (same as useSegments).
			onMerged={(next) => (router.setParams as (params: { id: string }) => void)({ id: next })}
			onOpenEntity={(entity) => router.push(`/entities/${entity}`)}
		/>
	);
}

export default function ItemScreen(): ReactElement {
	const params = useLocalSearchParams<{
		type?: string | string[];
		slug?: string | string[];
		id?: string | string[];
	}>();
	return (
		<ItemBody
			type={firstParam(params.type)}
			slug={firstParam(params.slug)}
			id={firstParam(params.id)}
		/>
	);
}
