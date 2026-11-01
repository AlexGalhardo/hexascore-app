import { type ReactElement, useState } from "react";
import { Pressable, Text, View } from "react-native";
import { t } from "../i18n";
import type { AppLocale } from "../i18n/locale";
import { type CatalogType, SINGULAR } from "../navigation/catalog";
import type { QueuePayload } from "../queue/store";
import { useRatingQueue } from "../rating/useRateMutation";
import { Btn } from "../ui/Btn";
import { type ItemStatus, STATUSES } from "./cache";

const FG = "text-fg dark:text-fg-dark";

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
		<View testID="status-sheet" className="gap-3 bg-bg px-4 py-4 dark:bg-bg-dark">
			<Text className={`font-text text-20 font-semibold ${FG}`}>
				{t(locale, "item.status.title")}
			</Text>
			<View accessibilityRole="radiogroup" className="gap-2">
				{STATUSES.map((status) => {
					const on = status === current;
					return (
						<Pressable
							key={status}
							accessibilityRole="radio"
							accessibilityState={{ checked: on }}
							className={`min-h-11 justify-center border border-fg px-4 dark:border-fg-dark ${on ? "bg-fg dark:bg-fg-dark" : "bg-field-bg dark:bg-field-bg-dark"}`}
							onPress={() => void write(status, { method: "put", body: { status } }).then(finish)}
						>
							<Text
								className={`font-text text-16 font-semibold ${on ? "text-bg dark:text-bg-dark" : FG}`}
							>
								{t(locale, `item.status.${SINGULAR[type]}.${status}`)}
							</Text>
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
