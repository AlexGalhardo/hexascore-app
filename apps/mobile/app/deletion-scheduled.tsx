import type { ReactElement } from "react";
import { Linking } from "react-native";
import { DeletionScheduledBody } from "../src/account/Blocked";
import { logout } from "../src/auth/logout";
import { apiBaseUrl, appClient, session } from "../src/auth/runtime";
import { useLocale } from "../src/providers/root";

export default function DeletionScheduledScreen(): ReactElement {
	const locale = useLocale();
	const state = session.state();
	return (
		<DeletionScheduledBody
			locale={locale}
			scheduledAt={
				state.kind === "deletionScheduled" ? state.scheduledAt : new Date().toISOString()
			}
			openWeb={(path) => void Linking.openURL(new URL(path, apiBaseUrl).toString())}
			onLogout={() =>
				void appClient.write(appClient.newIdempotencyKey(), (client) => logout({ client, session }))
			}
		/>
	);
}
