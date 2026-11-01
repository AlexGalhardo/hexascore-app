import { router } from "expo-router";
import { type ReactElement, useState, useSyncExternalStore } from "react";
import { logout } from "../../src/auth/logout";
import { appClient, session } from "../../src/auth/runtime";
import { t } from "../../src/i18n";
import { memoryPreferences } from "../../src/platform/preferences";
import { useLocale } from "../../src/providers/root";
import { CrashReportSwitch } from "../../src/telemetry/CrashReportSwitch";
import { applyThemePreference, readThemePreference } from "../../src/theme";
import { SettingsPage } from "../../src/ui/screens/SettingsPage";

export default function ProfileScreen(): ReactElement {
	const locale = useLocale();
	const state = useSyncExternalStore(session.subscribe, session.state);
	const [theme, setTheme] = useState(() => readThemePreference(memoryPreferences));
	return (
		<SettingsPage
			title={t(locale, "account.settings.title")}
			// Name and e-mail come from GET /me, which P5.B.4.2 wires.
			account={state.kind === "active" ? { name: null, email: null } : null}
			theme={theme}
			crashReport={<CrashReportSwitch locale={locale} />}
			labels={{
				account: t(locale, "account.settings.account"),
				appearance: t(locale, "account.settings.appearance"),
				theme: {
					system: t(locale, "account.settings.themeSystem"),
					light: t(locale, "account.settings.themeLight"),
					dark: t(locale, "account.settings.themeDark"),
				},
				privacy: t(locale, "account.privacy.title"),
				signOut: t(locale, "account.sessions.revoke"),
				signIn: t(locale, "auth.login.title"),
			}}
			onTheme={(next) => {
				applyThemePreference(memoryPreferences, next);
				setTheme(next);
			}}
			onSignOut={() =>
				void appClient.write(appClient.newIdempotencyKey(), (client) => logout({ client, session }))
			}
			onSignIn={() => router.navigate("/login")}
		/>
	);
}
