import { router } from "expo-router";
import type { ReactElement } from "react";
import { appClient } from "../../src/auth/runtime";
import { useLocale } from "../../src/providers/root";
import { SearchBody } from "../../src/screens/Search";

export default function SearchScreen(): ReactElement {
	const locale = useLocale();
	return (
		<SearchBody
			read={appClient.read}
			locale={locale}
			onOpen={(href) => router.push(href)}
			onCancel={() => router.navigate("/")}
		/>
	);
}
