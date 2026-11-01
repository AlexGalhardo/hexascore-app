import { type ReactElement, useState } from "react";
import { Text, View } from "react-native";
import { t } from "../i18n";
import type { AppLocale } from "../i18n/locale";
import { Btn } from "../ui/Btn";
import { TextField } from "../ui/TextField";
import { parseHours, STYLES, type TimeToBeat } from "./cache";
import { SHEET, SHEET_TITLE, useStatusWrite } from "./StatusSheet";

const FG = "text-fg dark:text-fg-dark";
const LABEL = {
	mainHours: "item.timeToBeat.styles.main",
	extrasHours: "item.timeToBeat.styles.extras",
	completionistHours: "item.timeToBeat.styles.completionist",
} as const;

export type TimeToBeatSheetProps = {
	itemId: string;
	/** The user's report from `/items/{id}/me`, or null. Games only: the item screen hides this sheet otherwise. */
	current: TimeToBeat | null;
	locale: AppLocale;
	onDone?: () => void;
};

export function TimeToBeatSheet({
	itemId,
	current,
	locale,
	onDone,
}: TimeToBeatSheetProps): ReactElement {
	const write = useStatusWrite(itemId, "timeToBeat");
	const [text, setText] = useState<Record<string, string>>(() =>
		Object.fromEntries(
			STYLES.map((s) => [s, current?.[s] === undefined ? "" : String(current[s])]),
		),
	);
	const [invalid, setInvalid] = useState<string[]>([]);
	const [failed, setFailed] = useState(false);
	const finish = (accepted: boolean): void => {
		setFailed(!accepted);
		if (accepted) onDone?.();
	};

	const save = (): void => {
		const parsed = STYLES.map((style) => [style, parseHours(text[style] ?? "")] as const);
		const bad = parsed.filter(([, hours]) => hours === "invalid").map(([style]) => style);
		const hours: TimeToBeat = Object.fromEntries(
			parsed.filter(
				(pair): pair is [(typeof STYLES)[number], number] => typeof pair[1] === "number",
			),
		);
		// The server needs at least one style (422 required on mainHours).
		if (bad.length === 0 && Object.keys(hours).length === 0) bad.push("mainHours");
		setInvalid(bad);
		if (bad.length > 0) return;
		void write(hours, { method: "put", body: hours }).then(finish);
	};

	return (
		<View testID="time-to-beat-sheet" className={SHEET}>
			<Text accessibilityRole="header" className={SHEET_TITLE}>
				{t(locale, "item.timeToBeat.yours")}
			</Text>
			<Text className="font-text text-14 text-muted dark:text-muted-dark">
				{t(locale, "item.timeToBeat.yoursHint")}
			</Text>
			{STYLES.map((style) => (
				<TextField
					key={style}
					label={t(locale, LABEL[style])}
					value={text[style] ?? ""}
					keyboardType="decimal-pad"
					onChangeText={(value) => setText((prev) => ({ ...prev, [style]: value }))}
					error={invalid.includes(style) ? t(locale, "errors.timeToBeatHours") : undefined}
				/>
			))}
			{failed ? (
				<Text className={`text-16 ${FG}`}>{t(locale, "item.timeToBeat.notSaved")}</Text>
			) : null}
			<Btn label={t(locale, "item.timeToBeat.save")} onPress={save} />
			{current !== null ? (
				<Btn
					label={t(locale, "item.timeToBeat.delete")}
					variant="ghost"
					onPress={() => void write(null, { method: "delete" }).then(finish)}
				/>
			) : null}
		</View>
	);
}
