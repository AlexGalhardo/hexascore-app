import { type ReactElement, useState } from "react";
import { ScrollView, Text, View } from "react-native";
import { type AppLocale, t } from "../i18n";
import { Button } from "../ui/Button";
import { TextField } from "../ui/TextField";
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

	const step = needsCode ? "twoFactor" : "credentials";
	return (
		<ScrollView
			className="flex-1 bg-bg dark:bg-bg-dark"
			contentContainerClassName="grow justify-center gap-6 px-6 py-12"
			keyboardShouldPersistTaps="handled"
		>
			<Text
				accessibilityElementsHidden
				importantForAccessibility="no"
				className="text-center font-display text-24 font-semibold tracking-tight text-fg dark:text-fg-dark"
			>
				HexaScore
			</Text>
			<View className="gap-1">
				<Text
					accessibilityRole="header"
					className="font-display text-30 font-semibold tracking-tight text-fg dark:text-fg-dark"
				>
					{t(locale, step === "twoFactor" ? "auth.login.twoFactorTitle" : "auth.login.title")}
				</Text>
				<Text className="font-text text-16 text-muted dark:text-muted-dark">
					{t(
						locale,
						step === "twoFactor"
							? useRecovery
								? "auth.login.backupHelp"
								: "auth.login.twoFactorHelp"
							: "auth.login.heading",
					)}
				</Text>
			</View>
			{step === "credentials" ? (
				<View className="gap-4">
					<TextField
						label={t(locale, "auth.fields.identifier")}
						value={identifier}
						onChangeText={setIdentifier}
						autoComplete="username"
					/>
					<TextField
						label={t(locale, "auth.fields.password")}
						value={password}
						onChangeText={setPassword}
						autoComplete="current-password"
						secure={{
							show: t(locale, "auth.fields.showShort"),
							hide: t(locale, "auth.fields.hideShort"),
						}}
					/>
					<TextField
						label={t(locale, "auth.fields.deviceName")}
						value={deviceName}
						onChangeText={setDeviceName}
					/>
				</View>
			) : (
				<View className="gap-4">
					<TextField
						label={t(locale, useRecovery ? "auth.login.backupCode" : "auth.login.code")}
						value={code}
						onChangeText={setCode}
						keyboardType={useRecovery ? "default" : "number-pad"}
					/>
					<Button
						variant="ghost"
						label={t(locale, useRecovery ? "auth.login.useTotp" : "auth.login.useBackup")}
						onPress={() => {
							setUseRecovery(!useRecovery);
							setCode("");
						}}
					/>
				</View>
			)}
			{error ? (
				<Text
					accessibilityRole="alert"
					className="font-text text-14 text-danger dark:text-danger-dark"
				>
					{error}
				</Text>
			) : null}
			<Button
				size="lg"
				label={t(locale, needsCode ? "auth.login.verify" : "auth.login.submit")}
				disabled={!canSubmit}
				loading={busy}
				onPress={() => void send()}
			/>
			{step === "credentials" ? (
				<View className="items-center gap-1">
					<Button
						variant="ghost"
						size="sm"
						label={t(locale, "auth.login.forgot")}
						onPress={() => props.openWeb("/forgot-password")}
					/>
					<Text className="font-text text-14 text-muted dark:text-muted-dark">
						{t(locale, "auth.login.noAccount")}
					</Text>
					<Button
						variant="outline"
						label={t(locale, "auth.login.createAccount")}
						onPress={() => props.openWeb("/signup")}
					/>
				</View>
			) : null}
		</ScrollView>
	);
}
