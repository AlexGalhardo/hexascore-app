import type { ReactElement } from "react";
import { Linking } from "react-native";
import { logout } from "../src/auth/logout";
import { appClient, session } from "../src/auth/runtime";
import { t } from "../src/i18n";
import { storeListingUrl } from "../src/platform/store";
import { useLocale } from "../src/providers/root";
import { StatePage } from "../src/ui/screens/StatePage";

export default function UpdateRequiredScreen(): ReactElement {
	const locale = useLocale();
	return (
		<StatePage
			title={t(locale, "account.updateRequired.title")}
			body={[]}
			actions={[
				{
					label: storeListingUrl(),
					role: "link",
					onPress: () => void Linking.openURL(storeListingUrl()),
				},
				{
					label: t(locale, "account.sessions.revoke"),
					variant: "secondary",
					onPress: () =>
						void appClient.write(appClient.newIdempotencyKey(), (client) =>
							logout({ client, session }),
						),
				},
			]}
		/>
	);
}
