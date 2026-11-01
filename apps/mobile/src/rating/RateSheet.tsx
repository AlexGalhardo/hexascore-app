import { type ReactElement, useState } from "react";
import { Text, View } from "react-native";
import { t } from "../i18n";
import type { AppLocale } from "../i18n/locale";
import { type CatalogType, SINGULAR } from "../navigation/catalog";
import { Btn } from "../ui/Btn";
import type { Scores } from "./cache";
import { CATEGORY_KEYS } from "./categories";
import { StarInput } from "./StarInput";
import { useRateMutation } from "./useRateMutation";

export type RateSheetProps = {
	itemId: string;
	type: CatalogType;
	/** The user's current rating from `/items/{id}/me`, or null. */
	initial: Scores | null;
	locale: AppLocale;
	onDone?: () => void;
};

export function RateSheet({ itemId, type, initial, locale, onDone }: RateSheetProps): ReactElement {
	const keys = CATEGORY_KEYS[type];
	const [scores, setScores] = useState<Partial<Scores>>(initial ?? {});
	const [failed, setFailed] = useState(false);
	const { rate, remove } = useRateMutation(itemId);
	const complete = keys.every((key) => typeof scores[key] === "number");
	const hasRating = initial !== null && Object.keys(initial).length > 0;

	const finish = (accepted: boolean): void => {
		setFailed(!accepted);
		if (accepted) onDone?.();
	};

	return (
		<View testID="rate-sheet" className="gap-3 bg-bg px-4 py-4 dark:bg-bg-dark">
			<Text className="font-text text-20 font-semibold text-fg dark:text-fg-dark">
				{t(locale, "ratings.title")}
			</Text>
			{keys.map((key) => (
				<StarInput
					key={key}
					label={t(locale, `categories.${SINGULAR[type]}.${key}`)}
					value={scores[key] ?? null}
					locale={locale}
					onChange={(value) => setScores((current) => ({ ...current, [key]: value }))}
				/>
			))}
			{failed ? (
				<Text className="text-16 text-fg dark:text-fg-dark">{t(locale, "ratings.notSaved")}</Text>
			) : null}
			<Btn
				label={t(locale, hasRating ? "ratings.update" : "ratings.save")}
				disabled={!complete}
				onPress={() =>
					void rate(Object.fromEntries(keys.map((k) => [k, scores[k] ?? 0]))).then(finish)
				}
			/>
			{hasRating ? (
				<Btn
					label={t(locale, "ratings.delete")}
					variant="ghost"
					onPress={() => void remove().then(finish)}
				/>
			) : null}
		</View>
	);
}
