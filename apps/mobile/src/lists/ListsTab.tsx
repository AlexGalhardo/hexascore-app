import { type ReactElement, useState } from "react";
import { Pressable, ScrollView, Text, View } from "react-native";
import { type AppLocale, t } from "../i18n";
import { Btn } from "../ui/Btn";
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

// shortcut: plain primitives on tokens; impl-mobile restyles it (IMPL-mobile-v1-lists-ui).
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
				{rows.map((list) => (
					<Pressable
						key={list.id}
						accessibilityRole="button"
						onPress={() => onOpen(list)}
						className="min-h-11 justify-center gap-1 border-b border-border py-3 dark:border-border-dark"
					>
						<Text className={`font-text text-16 font-bold ${FG}`}>{list.title}</Text>
						<Text className={MUTED}>
							{t(locale, "lists.itemCount", { count: list.itemCount })} ·{" "}
							{t(locale, list.isPublic ? "lists.public" : "lists.private")}
						</Text>
					</Pressable>
				))}
			</View>
			<View className="gap-3">
				<Text accessibilityRole="header" className={`font-text text-20 font-bold ${FG}`}>
					{t(locale, "lists.new")}
				</Text>
				<TextField label={t(locale, "lists.fieldTitle")} value={title} onChangeText={setTitle} />
				<Text className={MUTED}>{t(locale, "lists.fieldType")}</Text>
				<View className="flex-row flex-wrap gap-2">
					{TYPES.map((value) => (
						<Btn
							key={value}
							variant={type === value ? "primary" : "secondary"}
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
