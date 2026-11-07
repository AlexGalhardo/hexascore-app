import { type ReactElement, useState } from "react";
import { Pressable, ScrollView, Text, View } from "react-native";
import { type AppLocale, t } from "../i18n";
import { Button } from "../ui/Button";
import { Card } from "../ui/Card";
import { EmptyState } from "../ui/EmptyState";
import { Icon } from "../ui/Icon";
import { ListItem } from "../ui/ListItem";
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
const _PANEL =
	"gap-4 rounded-lg border border-border bg-surface p-4 dark:border-border-dark dark:bg-surface-dark";

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
			className={`min-h-11 justify-center rounded-full border px-4 ${selected ? "border-primary bg-primary dark:border-primary-dark dark:bg-primary-dark" : "border-border dark:border-border-dark"}`}
		>
			<Text
				className={`font-text text-14 ${selected ? "text-on-primary dark:text-on-primary-dark" : "text-fg dark:text-fg-dark"}`}
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
			<Text
				accessibilityRole="header"
				className={`font-display text-36 font-semibold tracking-tight ${FG}`}
			>
				{t(locale, "nav.lists")}
			</Text>
			{rows.length === 0 && !lists.isPending ? (
				<EmptyState title={t(locale, "lists.noLists")} />
			) : null}
			<View className="border-t border-border dark:border-border-dark">
				{rows.map((list) => {
					const meta = [
						t(locale, `catalog.title.${list.type}`),
						t(locale, "lists.itemCount", { count: list.itemCount }),
						t(locale, list.isPublic ? "lists.public" : "lists.private"),
					];
					return (
						<ListItem
							key={list.id}
							title={list.title}
							subtitle={meta.join(" · ")}
							accessibilityLabel={[list.title, ...meta].join(", ")}
							onPress={() => onOpen(list)}
							trailing={
								<Icon
									name="chevron-forward"
									size={20}
									className="text-muted dark:text-muted-dark"
								/>
							}
						/>
					);
				})}
			</View>
			<Card>
				<Text
					accessibilityRole="header"
					className={`font-display text-24 font-semibold tracking-tight ${FG}`}
				>
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
				<Button
					label={t(locale, "lists.create")}
					onPress={() => void create()}
					disabled={!online || title.trim() === ""}
				/>
				{online ? null : <Text className={MUTED}>{t(locale, "lists.offline")}</Text>}
			</Card>
		</ScrollView>
	);
}
