import type { ReactElement } from "react";
import { Linking, Pressable, Text, View } from "react-native";
import { logout } from "../src/auth/logout";
import { appClient, session } from "../src/auth/runtime";
import { t } from "../src/i18n";
import { storeListingUrl } from "../src/platform/store";

export default function UpdateRequiredScreen(): ReactElement {
	return (
		<View className="flex-1 items-center justify-center gap-4 bg-bg px-4 dark:bg-bg-dark">
			<Text className="text-20 text-fg dark:text-fg-dark">Update required</Text>
			<Pressable accessibilityRole="link" onPress={() => void Linking.openURL(storeListingUrl())}>
				<Text className="text-fg dark:text-fg-dark">{storeListingUrl()}</Text>
			</Pressable>
			<Pressable
				accessibilityRole="button"
				onPress={() =>
					void appClient.write(appClient.newIdempotencyKey(), (client) =>
						logout({ client, session }),
					)
				}
			>
				<Text className="text-fg dark:text-fg-dark">{t("en", "account.sessions.revoke")}</Text>
			</Pressable>
		</View>
	);
}
