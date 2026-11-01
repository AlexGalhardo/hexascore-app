import { router } from "expo-router";
import type { ReactElement } from "react";
import type { ItemJSON } from "../../../../packages/app-client/src/api";
import { appClient, appLocale } from "../../src/auth/runtime";
import { SINGULAR } from "../../src/components/CategorySection";
import { CATALOG_TYPES } from "../../src/navigation/catalog";
import { HomeBody } from "../../src/screens/Home";

function openItem(item: ItemJSON): void {
	const type = CATALOG_TYPES.find((plural) => SINGULAR[plural] === item.type);
	if (type) router.push(`/${type}/${item.slug}`);
}

export default function HomeScreen(): ReactElement {
	return <HomeBody read={appClient.read} locale={appLocale} onOpen={openItem} />;
}
