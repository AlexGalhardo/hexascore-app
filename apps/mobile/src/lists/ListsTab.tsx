import { type ReactElement, useState } from "react";
import { Pressable, ScrollView, Text, View } from "react-native";
import { type AppLocale, t } from "../i18n";
import { Btn } from "../ui/Btn";
import { Icon } from "../ui/Icon";
import { TextField } from "../ui/TextField";
import {
	type ListSummary,
	type ListsDeps,
	type ListType,
	useListActions,
	useMyLists,
	useOnline,
} from "./data";

const TYPES: ListType[] = ["game", "movie", "series", "book"];
const FG = "text-fg dark:text-fg-dark";
const MUTED = "font-text text-14 text-muted dark:text-muted-dark";
// Same panel as the status and add-to-list sheets.
const PANEL =
	"gap-4 border-t-2 border-primary bg-surface p-4 dark:border-primary-dark dark:bg-surface-dark";

function TypeChip({
	label,
	selected,
	onPress,
}: {
	label: string;
	selected: boolean;
	onPress: () => void;
}): ReactElement {
	return (
		<Pressable
			accessibilityRole="button"
			accessibilityState={{ selected }}
			onPress={onPress}
			className={`min-h-11 justify-center rounded-full border px-4 ${selected ? "border-fg bg-fg dark:border-fg-dark dark:bg-fg-dark" : "border-border dark:border-border-dark"}`}
		>
			<Text
				className={`font-text text-14 ${selected ? "text-bg dark:text-bg-dark" : "text-fg dark:text-fg-dark"}`}
			>
				{label}
			</Text>
		</Pressable>
	);
}

export function ListsTab({
	deps,
	locale,
	onOpen,
}: {
	deps: ListsDeps;
	locale: AppLocale;
	onOpen: (list: ListSummary) => void;
}): ReactElement {
	const lists = useMyLists(deps);
	const actions = useListActions(deps, locale);
	const online = useOnline();
	const [title, setTitle] = useState("");
	const [type, setType] = useState<ListType>("game");
	const rows = lists.data ?? [];

	const create = async (): Promise<void> => {
		if (await actions.create({ type, title })) {
			setTitle("");
			deps.notify(t(locale, "lists.created"));
		}
	};

	return (
		<ScrollView
			className="flex-1 bg-bg dark:bg-bg-dark"
			contentContainerClassName="gap-6 px-4 pb-8 pt-6"
		>
			<Text accessibilityRole="header" className={`font-display text-36 font-bold ${FG}`}>
				{t(locale, "nav.lists")}
			</Text>
			{rows.length === 0 && !lists.isPending ? (
				<Text className={`font-text text-16 ${FG}`}>{t(locale, "lists.noLists")}</Text>
			) : null}
			<View className="border-t-2 border-fg dark:border-fg-dark">
				{rows.map((list) => {
					const meta = [
						t(locale, `catalog.title.${list.type}`),
						t(locale, "lists.itemCount", { count: list.itemCount }),
						t(locale, list.isPublic ? "lists.public" : "lists.private"),
					];
					return (
						<Pressable
							key={list.id}
							accessibilityRole="button"
							accessibilityLabel={[list.title, ...meta].join(", ")}
							onPress={() => onOpen(list)}
							className="min-h-11 flex-row items-center gap-4 border-b border-border py-3 active:bg-surface dark:border-border-dark dark:active:bg-surface-dark"
						>
							<View className="flex-1 gap-1">
								<Text className={`font-text text-16 font-bold ${FG}`}>{list.title}</Text>
								<Text className={`${MUTED} tabular-nums`}>{meta.join(" · ")}</Text>
							</View>
							<Icon name="chevron-forward" size={20} className="text-muted dark:text-muted-dark" />
						</Pressable>
					);
				})}
			</View>
			<View className={PANEL}>
				<Text accessibilityRole="header" className={`font-display text-24 font-bold ${FG}`}>
					{t(locale, "lists.new")}
				</Text>
				<TextField label={t(locale, "lists.fieldTitle")} value={title} onChangeText={setTitle} />
				<Text className={MUTED}>{t(locale, "lists.fieldType")}</Text>
				<View className="flex-row flex-wrap gap-2">
					{TYPES.map((value) => (
						<TypeChip
							key={value}
							selected={type === value}
							label={t(locale, `catalog.title.${value}`)}
							onPress={() => setType(value)}
						/>
					))}
				</View>
				<Btn
					label={t(locale, "lists.create")}
					onPress={() => void create()}
					disabled={!online || title.trim() === ""}
				/>
				{online ? null : <Text className={MUTED}>{t(locale, "lists.offline")}</Text>}
			</View>
		</ScrollView>
	);
}
