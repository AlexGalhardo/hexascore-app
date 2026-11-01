import { router } from "expo-router";
import type { ReactElement } from "react";
import type { ItemJSON } from "../../../../packages/app-client/src/api";
import { appClient } from "../../src/auth/runtime";
import { isCatalogType, itemHref } from "../../src/navigation/catalog";
import { useLocale } from "../../src/providers/root";
import { HomeBody } from "../../src/screens/Home";

function openItem(item: ItemJSON): void {
	if (isCatalogType(item.type)) router.push(itemHref(item.type, item));
}

export default function HomeScreen(): ReactElement {
	return <HomeBody read={appClient.read} locale={useLocale()} onOpen={openItem} />;
}
