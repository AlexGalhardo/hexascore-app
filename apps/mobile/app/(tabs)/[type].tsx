import { useLocalSearchParams } from "expo-router";
import type { ReactElement } from "react";
import { firstParam, isCatalogType } from "../../src/navigation/catalog";
import { Placeholder } from "../../src/ui/Placeholder";

export function SectionBody({ type }: { type: string }): ReactElement {
	if (!isCatalogType(type)) return <Placeholder title="Unknown section" />;
	return <Placeholder title={`Section ${type}`} />;
}

export default function SectionScreen(): ReactElement {
	const { type } = useLocalSearchParams<{ type?: string | string[] }>();
	return <SectionBody type={firstParam(type)} />;
}
