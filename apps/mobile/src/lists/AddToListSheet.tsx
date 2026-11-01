import { type ReactElement, useState } from "react";
import { Pressable, Text, View } from "react-native";
import { type AppLocale, t } from "../i18n";
import { Btn } from "../ui/Btn";
import { TextField } from "../ui/TextField";
import { type ListsDeps, type ListType, useListActions, useMyLists, useOnline } from "./data";

const FG = "text-fg dark:text-fg-dark";
const MUTED = "font-text text-14 text-muted dark:text-muted-dark";

/** The item's own type only: the server answers 422 `list-type-mismatch` otherwise (SDD 14 §7.1). */
export function AddToListSheet({
	deps,
	itemId,
	type,
	locale,
	onDone,
}: {
	deps: ListsDeps;
	itemId: string;
	type: ListType;
	locale: AppLocale;
	onDone?: () => void;
}): ReactElement {
	const lists = useMyLists(deps);
	const actions = useListActions(deps, locale);
	const online = useOnline();
	const [title, setTitle] = useState("");
	const rows = (lists.data ?? []).filter((l) => l.type === type);

	const createWithItem = async (): Promise<void> => {
		if (await actions.create({ type, title, firstItemId: itemId })) {
			deps.notify(t(locale, "lists.created"));
			onDone?.();
		}
	};

	return (
		<View className="gap-3 bg-bg px-4 py-4 dark:bg-bg-dark">
			<Text accessibilityRole="header" className={`font-text text-20 font-bold ${FG}`}>
				{t(locale, "lists.addToList")}
			</Text>
			{rows.length === 0 && !lists.isPending ? (
				<Text className={MUTED}>
					{t(locale, `lists.noListsOfType.${type === "book" ? "other" : type}`)}
				</Text>
			) : null}
			{rows.map((list) => (
				<Pressable
					key={list.id}
					accessibilityRole="button"
					onPress={() => void actions.addItem(list, itemId).then((ok) => ok && onDone?.())}
					className="min-h-11 justify-center border-b border-border py-2 dark:border-border-dark"
				>
					<Text className={`font-text text-16 ${FG}`}>
						{list.title} · {t(locale, "lists.itemCount", { count: list.itemCount })}
					</Text>
				</Pressable>
			))}
			<TextField label={t(locale, "lists.fieldTitle")} value={title} onChangeText={setTitle} />
			<Btn
				variant="secondary"
				label={t(locale, "lists.newWithItem")}
				onPress={() => void createWithItem()}
				disabled={!online || title.trim() === ""}
			/>
			{online ? null : <Text className={MUTED}>{t(locale, "lists.offline")}</Text>}
		</View>
	);
}
