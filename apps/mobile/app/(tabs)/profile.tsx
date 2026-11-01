import { router } from "expo-router";
import { type ReactElement, useState, useSyncExternalStore } from "react";
import { logout } from "../../src/auth/logout";
import { appClient, appLocale, session } from "../../src/auth/runtime";
import { notificationsApi } from "../../src/notifications/runtime";
import { memoryPreferences } from "../../src/platform/preferences";
import { AccountExtras } from "../../src/profile/AccountExtras";
import type { ProfileApi } from "../../src/profile/me";
import { ProfileSettings } from "../../src/profile/Settings";
import { useLocale } from "../../src/providers/root";
import { CrashReportSwitch } from "../../src/telemetry/CrashReportSwitch";
import { applyThemePreference, readThemePreference } from "../../src/theme";

const api: ProfileApi = {
	read: appClient.read,
	write: (call) => appClient.write(appClient.newIdempotencyKey(), call),
};

export default function ProfileScreen(): ReactElement {
	const locale = useLocale();
	const state = useSyncExternalStore(session.subscribe, session.state);
	const [theme, setTheme] = useState(() => readThemePreference(memoryPreferences));
	return (
		<ProfileSettings
			api={api}
			signedIn={state.kind === "active"}
			locale={locale}
			onLocale={appLocale.set}
			theme={theme}
			onTheme={(next) => {
				applyThemePreference(memoryPreferences, next);
				setTheme(next);
			}}
			crashReport={<CrashReportSwitch locale={locale} />}
			onSignOut={() =>
				void appClient.write(appClient.newIdempotencyKey(), (client) => logout({ client, session }))
			}
			onCurrentRevoked={() => void session.signOut()}
			onSignIn={() => router.navigate("/login")}
			extra={
				<AccountExtras
					locale={locale}
					signedIn={state.kind === "active"}
					notifications={notificationsApi}
					write={(call) => appClient.write(appClient.newIdempotencyKey(), call)}
					// The server revokes every token on schedule; the session shows the cancel screen (SDD 10 §6.2).
					onScheduled={(deletesAt) => session.observeMe({ deletionScheduledAt: deletesAt })}
					onOpen={(path) => router.push(path)}
				/>
			}
		/>
	);
}
