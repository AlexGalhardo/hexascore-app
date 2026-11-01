import { router } from "expo-router";
import type { ReactElement } from "react";
import { Linking } from "react-native";
import { defaultDeviceName } from "../../src/auth/deviceName";
import { LoginForm } from "../../src/auth/LoginForm";
import { login } from "../../src/auth/login";
import { apiBaseUrl, appClient, session } from "../../src/auth/runtime";
import { useLocale } from "../../src/providers/root";

export default function LoginScreen(): ReactElement {
	const locale = useLocale();
	return (
		<LoginForm
			locale={locale}
			initialDeviceName={defaultDeviceName()}
			submit={(input) =>
				appClient.write(appClient.newIdempotencyKey(), (client) => login(client, input))
			}
			onSignedIn={async (token) => {
				await session.signIn(token);
				router.replace("/");
			}}
			openWeb={(path) => void Linking.openURL(new URL(path, apiBaseUrl).toString())}
		/>
	);
}
