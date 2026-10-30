import { useLocalSearchParams } from "expo-router";
import type { ReactElement } from "react";
import { firstParam, isCatalogType } from "../../src/navigation/catalog";
import { Placeholder } from "../../src/ui/Placeholder";

export function ItemBody({ type, slug }: { type: string; slug: string }): ReactElement {
	if (!isCatalogType(type)) return <Placeholder title="Unknown section" />;
	return <Placeholder title={`Item ${type} ${slug}`} />;
}

export default function ItemScreen(): ReactElement {
	const params = useLocalSearchParams<{ type?: string | string[]; slug?: string | string[] }>();
	return <ItemBody type={firstParam(params.type)} slug={firstParam(params.slug)} />;
}
