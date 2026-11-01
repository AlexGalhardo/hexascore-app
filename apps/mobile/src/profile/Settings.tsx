import { useQuery, useQueryClient } from "@tanstack/react-query";
import { type ReactElement, type ReactNode, useState } from "react";
import { Pressable, Switch, Text, View } from "react-native";
import { describeError } from "../api/problems";
import { type AppLocale, t } from "../i18n";
import { APP_LOCALES } from "../i18n/locale";
import type { ThemePreference } from "../theme";
import { Btn } from "../ui/Btn";
import { SettingsPage, SettingsSection } from "../ui/screens/SettingsPage";
import { TextField } from "../ui/TextField";
import { Devices } from "./Devices";
import {
	isNicknameTaken,
	type Me,
	type MeChange,
	meKey,
	type ProfileApi,
	parseMe,
	patchBody,
} from "./me";

// Native names, as on the web profile: a language is listed in its own language.
const LANGUAGE_NAMES: Record<AppLocale, string> = {
	en: "English",
	"pt-BR": "Português",
	es: "Español",
};

const FG = "text-fg dark:text-fg-dark";

export type ProfileSettingsProps = {
	api: ProfileApi;
	signedIn: boolean;
	locale: AppLocale;
	/** Applies the locale override on the device; the UI re-renders, no restart (FR-10-6). */
	onLocale: (locale: AppLocale) => void;
	theme: ThemePreference;
	onTheme: (theme: ThemePreference) => void;
	crashReport: ReactNode;
	onSignOut: () => void;
	/** Local sign-out after this device's own token was revoked. */
	onCurrentRevoked: () => void;
	onSignIn: () => void;
};

function LanguagePicker({
	locale,
	label,
	onChoose,
}: {
	locale: AppLocale;
	label: string;
	onChoose: (locale: AppLocale) => void;
}): ReactElement {
	return (
		<View accessibilityRole="radiogroup" accessibilityLabel={label} className="flex-row">
			{APP_LOCALES.map((option) => {
				const on = option === locale;
				return (
					<Pressable
						key={option}
						accessibilityRole="radio"
						accessibilityLabel={LANGUAGE_NAMES[option]}
						accessibilityState={{ checked: on }}
						onPress={() => onChoose(option)}
						className={`min-h-11 flex-1 items-center justify-center border border-fg dark:border-fg-dark ${on ? "bg-fg dark:bg-fg-dark" : "bg-field-bg dark:bg-field-bg-dark"}`}
					>
						<Text
							className={`font-text text-16 font-semibold ${on ? "text-bg dark:text-bg-dark" : FG}`}
						>
							{LANGUAGE_NAMES[option]}
						</Text>
					</Pressable>
				);
			})}
		</View>
	);
}

function ProfileForm({
	me,
	locale,
	save,
	onLocale,
}: {
	me: Me;
	locale: AppLocale;
	save: (change: MeChange) => Promise<void>;
	onLocale: (locale: AppLocale) => void;
}): ReactElement {
	const [name, setName] = useState(me.name);
	const [nickname, setNickname] = useState("");
	const [nicknameError, setNicknameError] = useState<string | undefined>();
	const [error, setError] = useState<string | null>(null);
	const [saved, setSaved] = useState(false);

	const run = async (change: MeChange): Promise<void> => {
		setError(null);
		setNicknameError(undefined);
		setSaved(false);
		try {
			await save(change);
			setSaved(true);
		} catch (cause) {
			if (isNicknameTaken(cause)) setNicknameError(t(locale, "auth.errors.usernameTaken"));
			else setError(describeError(cause).key);
		}
	};

	return (
		<>
			<SettingsSection title={t(locale, "account.details.title")}>
				<TextField label={t(locale, "account.details.name")} value={name} onChangeText={setName} />
				{me.username === null ? (
					<TextField
						label={t(locale, "account.details.nickname")}
						value={nickname}
						onChangeText={setNickname}
						error={nicknameError}
					/>
				) : (
					<Text className={`font-text text-16 ${FG}`}>
						{`${t(locale, "account.details.nickname")}: ${me.username}`}
					</Text>
				)}
				<Btn
					label={t(locale, "account.details.save")}
					onPress={() =>
						void run({
							name: name.trim(),
							...(nickname.trim() ? { username: nickname.trim() } : {}),
						})
					}
				/>
				{saved ? (
					<Text accessibilityLiveRegion="polite" className={`font-text text-14 ${FG}`}>
						{t(locale, "account.saved")}
					</Text>
				) : null}
				{error ? (
					<Text accessibilityRole="alert" className={`font-text text-14 ${FG}`}>
						{t(locale, error)}
					</Text>
				) : null}
			</SettingsSection>
			<SettingsSection title={t(locale, "account.details.locale")}>
				<LanguagePicker
					locale={locale}
					label={t(locale, "account.details.locale")}
					onChoose={(next) => {
						onLocale(next);
						void run({ locale: next });
					}}
				/>
			</SettingsSection>
		</>
	);
}

/** Lives in the page's Privacy section, next to the crash-report switch (D-29: ships before "For you"). */
function PersonalizationSwitch({
	me,
	locale,
	save,
}: {
	me: Me;
	locale: AppLocale;
	save: (change: MeChange) => Promise<void>;
}): ReactElement {
	const [error, setError] = useState<string | null>(null);
	const label = t(locale, "account.privacy.personalization");
	return (
		<View className="gap-1">
			<View className="min-h-11 flex-row items-center gap-4">
				<View className="flex-1 gap-1">
					<Text className={`font-text text-16 font-semibold ${FG}`}>{label}</Text>
					<Text className="font-text text-14 text-muted dark:text-muted-dark">
						{t(locale, "account.privacy.personalizationHelp")}
					</Text>
				</View>
				<Switch
					accessibilityRole="switch"
					accessibilityLabel={label}
					value={me.personalizationOptOut}
					onValueChange={(next) => {
						setError(null);
						save({ personalizationOptOut: next }).catch((cause: unknown) =>
							setError(describeError(cause).key),
						);
					}}
				/>
			</View>
			{error ? (
				<Text accessibilityRole="alert" className={`font-text text-14 ${FG}`}>
					{t(locale, error)}
				</Text>
			) : null}
		</View>
	);
}

/** The profile tab (SDD 10 §6.1 Profile row): GET/PATCH /me, devices, consent, theme, logout. */
export function ProfileSettings(props: ProfileSettingsProps): ReactElement {
	const { api, locale } = props;
	const queryClient = useQueryClient();
	const me = useQuery({
		queryKey: meKey,
		queryFn: async () => parseMe(await api.read.meShow()),
		enabled: props.signedIn,
	});
	const profile = props.signedIn ? (me.data ?? null) : null;

	const save = async (change: MeChange): Promise<void> => {
		if (!profile) return;
		const body = await api.write((client) => client.meUpdate({ body: patchBody(profile, change) }));
		queryClient.setQueryData(meKey, parseMe(body) ?? profile);
	};

	return (
		<SettingsPage
			title={t(locale, "account.settings.title")}
			account={
				props.signedIn ? { name: profile?.name ?? null, email: profile?.email ?? null } : null
			}
			theme={props.theme}
			crashReport={
				<>
					{profile ? <PersonalizationSwitch me={profile} locale={locale} save={save} /> : null}
					{props.crashReport}
				</>
			}
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
			onTheme={props.onTheme}
			onSignOut={props.onSignOut}
			onSignIn={props.onSignIn}
		>
			{profile ? (
				<ProfileForm
					key={profile.username ?? ""}
					me={profile}
					locale={locale}
					save={save}
					onLocale={props.onLocale}
				/>
			) : null}
			{props.signedIn ? (
				<Devices api={api} locale={locale} onCurrentRevoked={props.onCurrentRevoked} />
			) : null}
		</SettingsPage>
	);
}
