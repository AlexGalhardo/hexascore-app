import { router, useLocalSearchParams } from "expo-router";
import type { ReactElement } from "react";
import type { ItemJSON } from "../../../../packages/app-client/src/api";
import { appClient } from "../../src/auth/runtime";
import { firstParam, isCatalogType, itemHref } from "../../src/navigation/catalog";
import { useLocale } from "../../src/providers/root";
import { ListingBody } from "../../src/screens/Listing";
import { Placeholder } from "../../src/ui/Placeholder";

export function SectionBody({ type }: { type: string }): ReactElement {
	const locale = useLocale();
	if (!isCatalogType(type)) return <Placeholder title="Unknown section" />;
	const openItem = (item: ItemJSON): void => router.push(itemHref(type, item));
	return <ListingBody type={type} read={appClient.read} locale={locale} onOpen={openItem} />;
}

export default function SectionScreen(): ReactElement {
	const { type } = useLocalSearchParams<{ type?: string | string[] }>();
	return <SectionBody type={firstParam(type)} />;
}
