import { router } from "expo-router";
import type { ReactElement } from "react";
import { Linking } from "react-native";
import { defaultDeviceName } from "../../src/auth/deviceName";
import { signInWithGoogle } from "../../src/auth/google";
import { googleClientId, promptGoogleIdToken } from "../../src/auth/googlePrompt";
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
			google={
				googleClientId === ""
					? undefined
					: {
							prompt: () =>
								promptGoogleIdToken(googleClientId, async () => {
									const issued = await appClient.write(appClient.newIdempotencyKey(), (client) =>
										client.authGoogleNonce({ body: {} }),
									);
									const nonce = (issued as { nonce?: unknown } | null)?.nonce;
									return typeof nonce === "string" ? nonce : null;
								}),
							submit: (input) =>
								appClient.write(appClient.newIdempotencyKey(), (client) =>
									signInWithGoogle(client, input),
								),
						}
			}
			openWeb={(path) => void Linking.openURL(new URL(path, apiBaseUrl).toString())}
		/>
	);
}
