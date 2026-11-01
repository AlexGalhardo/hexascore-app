import { type ReactElement, useState } from "react";
import { Pressable, Text, View } from "react-native";
import { t } from "../i18n";
import type { AppLocale } from "../i18n/locale";
import { type CatalogType, SINGULAR } from "../navigation/catalog";
import type { QueuePayload } from "../queue/store";
import { useRatingQueue } from "../rating/useRateMutation";
import { Btn } from "../ui/Btn";
import { Icon } from "../ui/Icon";
import { type ItemStatus, STATUSES } from "./cache";

const FG = "text-fg dark:text-fg-dark";
// v1 look of the rating sheet (no separate frame for status and time to beat).
export const SHEET =
	"gap-4 border-t-2 border-primary bg-surface p-4 dark:border-primary-dark dark:bg-surface-dark";
export const SHEET_TITLE = "font-display text-24 font-bold text-fg dark:text-fg-dark";

/** Optimistic write on the `status` or `timeToBeat` lane; false when the queue refused it (full or signed out). */
export function useStatusWrite(
	itemId: string,
	op: "status" | "timeToBeat",
): (value: unknown, payload: QueuePayload) => Promise<boolean> {
	const queue = useRatingQueue();
	return async (value, payload) => {
		queue.statusCache.optimistic(itemId, op, value);
		const entry = await queue.enqueue({ op, itemId, payload });
		if (entry === null) {
			queue.statusCache.restore(itemId, op);
			return false;
		}
		void queue.drain();
		return true;
	};
}

export type StatusSheetProps = {
	itemId: string;
	type: CatalogType;
	/** The user's status from `/items/{id}/me`, or null. */
	current: ItemStatus | null;
	locale: AppLocale;
	onDone?: () => void;
};

export function StatusSheet({
	itemId,
	type,
	current,
	locale,
	onDone,
}: StatusSheetProps): ReactElement {
	const write = useStatusWrite(itemId, "status");
	const [failed, setFailed] = useState(false);
	const finish = (accepted: boolean): void => {
		setFailed(!accepted);
		if (accepted) onDone?.();
	};

	return (
		<View testID="status-sheet" className={SHEET}>
			<Text accessibilityRole="header" className={SHEET_TITLE}>
				{t(locale, "item.status.title")}
			</Text>
			<View accessibilityRole="radiogroup" className="gap-2">
				{STATUSES.map((status) => {
					const on = status === current;
					const label = t(locale, `item.status.${SINGULAR[type]}.${status}`);
					return (
						<Pressable
							key={status}
							accessibilityRole="radio"
							accessibilityLabel={label}
							accessibilityState={{ checked: on }}
							className={`min-h-11 flex-row items-center justify-between gap-3 rounded-sm border border-b-2 border-fg px-4 active:border-b dark:border-fg-dark ${on ? "bg-fg dark:bg-fg-dark" : "bg-field-bg dark:bg-field-bg-dark"}`}
							onPress={() => void write(status, { method: "put", body: { status } }).then(finish)}
						>
							<Text
								className={`font-text text-16 font-semibold ${on ? "text-bg dark:text-bg-dark" : FG}`}
							>
								{label}
							</Text>
							{on ? (
								<View testID="status-checked">
									<Icon name="checkmark" size={20} className="text-bg dark:text-bg-dark" />
								</View>
							) : null}
						</Pressable>
					);
				})}
			</View>
			{failed ? <Text className={`text-16 ${FG}`}>{t(locale, "item.status.notSaved")}</Text> : null}
			{current !== null ? (
				<Btn
					label={t(locale, "item.status.none")}
					variant="ghost"
					onPress={() => void write(null, { method: "delete" }).then(finish)}
				/>
			) : null}
		</View>
	);
}
