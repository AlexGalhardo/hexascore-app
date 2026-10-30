import { useLocalSearchParams } from "expo-router";
import type { ReactElement } from "react";
import { firstParam } from "../../src/navigation/catalog";
import { Placeholder } from "../../src/ui/Placeholder";

export default function EntityScreen(): ReactElement {
	const { id } = useLocalSearchParams<{ id?: string | string[] }>();
	return <Placeholder title={`Entity ${firstParam(id)}`} />;
}
