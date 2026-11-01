import { router, useLocalSearchParams } from "expo-router";
import type { ReactElement } from "react";
import { appClient } from "../../src/auth/runtime";
import { firstParam } from "../../src/navigation/catalog";
import { useLocale } from "../../src/providers/root";
import { EntityView } from "../../src/screens/Entity";

export default function EntityScreen(): ReactElement {
	const { id } = useLocalSearchParams<{ id?: string | string[] }>();
	return (
		<EntityView
			id={firstParam(id)}
			read={appClient.read}
			locale={useLocale()}
			onOpen={(href) => router.push(href)}
			onBrowse={() => router.navigate("/")}
		/>
	);
}
