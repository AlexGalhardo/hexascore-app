import { router, useLocalSearchParams } from "expo-router";
import { type ReactElement, useState } from "react";
import { View } from "react-native";
import type { AppItemJSON } from "../../../../packages/app-client/src/api";
import { appClient } from "../../src/auth/runtime";
import { t } from "../../src/i18n";
import {
	CATALOG_TYPES,
	type CatalogType,
	firstParam,
	isCatalogType,
	itemHref,
	SINGULAR,
} from "../../src/navigation/catalog";
import { useLocale } from "../../src/providers/root";
import { ListingBody } from "../../src/screens/Listing";
import { Placeholder } from "../../src/ui/Placeholder";
import { SegmentedControl } from "../../src/ui/SegmentedControl";

export function SectionBody({ type }: { type: string }): ReactElement {
	const locale = useLocale();
	if (!isCatalogType(type)) return <Placeholder title="Unknown section" />;
	const openItem = (item: AppItemJSON): void => router.push(itemHref(type, item));
	return <ListingBody type={type} read={appClient.read} locale={locale} onOpen={openItem} />;
}

// The Sections tab has no route param: a segmented control picks the category (games first).
export default function SectionScreen(): ReactElement {
	const { type } = useLocalSearchParams<{ type?: string | string[] }>();
	const locale = useLocale();
	const initial = firstParam(type);
	const [picked, setPicked] = useState<CatalogType>(isCatalogType(initial) ? initial : "games");
	return (
		<View className="flex-1 bg-bg dark:bg-bg-dark">
			<View className="px-4 pt-3">
				<SegmentedControl
					label={t(locale, "nav.sections")}
					value={picked}
					onChange={setPicked}
					options={CATALOG_TYPES.map((value) => ({
						value,
						label: t(locale, `catalog.title.${SINGULAR[value]}`),
					}))}
				/>
			</View>
			<SectionBody type={picked} />
		</View>
	);
}
