import { router } from "expo-router";
import type { ReactElement } from "react";
import type { ItemJSON } from "../../../../packages/app-client/src/api";
import { appClient, appLocale } from "../../src/auth/runtime";
import { type CatalogType, isCatalogType, itemHref } from "../../src/navigation/catalog";
import { HomeBody } from "../../src/screens/Home";

function openItem(item: ItemJSON): void {
	if (isCatalogType(item.type)) router.push(itemHref(item.type, item));
}

function openSection(type: CatalogType): void {
	router.push(`/${type}`);
}

export default function HomeScreen(): ReactElement {
	return (
		<HomeBody read={appClient.read} locale={appLocale} onOpen={openItem} onSeeAll={openSection} />
	);
}
