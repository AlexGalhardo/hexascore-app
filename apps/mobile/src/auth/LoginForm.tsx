import { type ReactElement, useState } from "react";
import { Pressable, Text, TextInput, View } from "react-native";
import { type AppLocale, t } from "../i18n";
import { isValidDeviceName } from "./deviceName";
import type { LoginInput, LoginOutcome } from "./login";
import type { StoredToken } from "./tokenStore";

export type LoginFormProps = {
	locale: AppLocale;
	initialDeviceName: string;
	submit: (input: LoginInput) => Promise<LoginOutcome>;
	onSignedIn: (token: StoredToken) => Promise<void>;
	/** Opens the web challenge, signup or recovery page. */
	openWeb: (path: "/login" | "/signup" | "/forgot-password") => void;
};

const field = "rounded-md border border-border px-3 py-3 text-16 text-fg dark:text-fg-dark";

export function LoginForm(props: LoginFormProps): ReactElement {
	const { locale } = props;
	const [identifier, setIdentifier] = useState("");
	const [password, setPassword] = useState("");
	const [deviceName, setDeviceName] = useState(props.initialDeviceName);
	const [needsCode, setNeedsCode] = useState(false);
	const [useRecovery, setUseRecovery] = useState(false);
	const [code, setCode] = useState("");
	const [error, setError] = useState<string | null>(null);
	const [busy, setBusy] = useState(false);

	const canSubmit =
		!busy &&
		identifier.length > 0 &&
		password.length > 0 &&
		isValidDeviceName(deviceName) &&
		(!needsCode || code.trim().length > 0);

	const send = async (): Promise<void> => {
		setBusy(true);
		setError(null);
		const second = needsCode
			? useRecovery
				? { recoveryCode: code.trim() }
				: { code: code.trim() }
			: {};
		const outcome = await props.submit({
			identifier,
			password,
			deviceName: deviceName.trim(),
			...second,
		});
		setBusy(false);
		if (outcome.kind === "ok") await props.onSignedIn(outcome.token);
		else if (outcome.kind === "twoFactor") setNeedsCode(true);
		else if (outcome.kind === "challenge") props.openWeb("/login");
		else setError(t(locale, outcome.key));
	};

	return (
		<View className="flex-1 gap-3 bg-bg px-4 pt-12 dark:bg-bg-dark">
			<Text accessibilityRole="header" className="text-20 text-fg dark:text-fg-dark">
				{t(locale, "auth.login.title")}
			</Text>
			<TextInput
				accessibilityLabel={t(locale, "auth.fields.identifier")}
				autoCapitalize="none"
				autoComplete="username"
				className={field}
				onChangeText={setIdentifier}
				value={identifier}
			/>
			<TextInput
				accessibilityLabel={t(locale, "auth.fields.password")}
				autoCapitalize="none"
				autoComplete="current-password"
				className={field}
				onChangeText={setPassword}
				secureTextEntry
				value={password}
			/>
			{needsCode ? (
				<>
					<Text className="text-fg dark:text-fg-dark">
						{t(locale, useRecovery ? "auth.login.backupHelp" : "auth.login.twoFactorHelp")}
					</Text>
					<TextInput
						accessibilityLabel={t(
							locale,
							useRecovery ? "auth.login.backupCode" : "auth.login.code",
						)}
						autoCapitalize="none"
						className={field}
						keyboardType={useRecovery ? "default" : "number-pad"}
						onChangeText={setCode}
						value={code}
					/>
					<Pressable
						accessibilityRole="button"
						onPress={() => {
							setUseRecovery(!useRecovery);
							setCode("");
						}}
					>
						<Text className="text-fg dark:text-fg-dark">
							{t(locale, useRecovery ? "auth.login.useTotp" : "auth.login.useBackup")}
						</Text>
					</Pressable>
				</>
			) : null}
			<TextInput
				accessibilityLabel="Device name"
				className={field}
				maxLength={64}
				onChangeText={setDeviceName}
				value={deviceName}
			/>
			{error ? (
				<Text accessibilityRole="alert" className="text-fg dark:text-fg-dark">
					{error}
				</Text>
			) : null}
			<Pressable
				accessibilityRole="button"
				accessibilityState={{ disabled: !canSubmit }}
				disabled={!canSubmit}
				onPress={() => void send()}
			>
				<Text className="text-fg dark:text-fg-dark">
					{t(locale, needsCode ? "auth.login.verify" : "auth.login.submit")}
				</Text>
			</Pressable>
			<Pressable accessibilityRole="link" onPress={() => props.openWeb("/forgot-password")}>
				<Text className="text-fg dark:text-fg-dark">{t(locale, "auth.login.forgot")}</Text>
			</Pressable>
			<Pressable accessibilityRole="link" onPress={() => props.openWeb("/signup")}>
				<Text className="text-fg dark:text-fg-dark">{t(locale, "auth.login.createAccount")}</Text>
			</Pressable>
		</View>
	);
}
