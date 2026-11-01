import { router } from "expo-router";
import type { ReactElement } from "react";
import { listsDeps } from "../../src/auth/runtime";
import { ListsTab } from "../../src/lists/ListsTab";
import { useLocale } from "../../src/providers/root";

export default function ListsScreen(): ReactElement {
	const locale = useLocale();
	return (
		<ListsTab
			deps={listsDeps}
			locale={locale}
			onOpen={(list) => router.push(`/lists/${encodeURIComponent(list.id)}`)}
		/>
	);
}
