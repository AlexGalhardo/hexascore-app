import { router } from "expo-router";
import type { ReactElement } from "react";
import { appClient } from "../../src/auth/runtime";
import { useLocale } from "../../src/providers/root";
import { SearchBody } from "../../src/screens/Search";
import { searchClickSender } from "../../src/search/events";

const sendClick = searchClickSender((body) =>
	appClient.write(appClient.newIdempotencyKey(), (client) => client.searchEvents({ body })),
);

export default function SearchScreen(): ReactElement {
	const locale = useLocale();
	return (
		<SearchBody
			read={appClient.read}
			locale={locale}
			onOpen={(href) => router.push(href)}
			onCancel={() => router.navigate("/")}
			onClick={sendClick}
		/>
	);
}
