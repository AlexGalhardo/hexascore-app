import { type ReactElement, useState } from "react";
import { Alert, View } from "react-native";
import { type AppLocale, t } from "../i18n";
import { Btn } from "../ui/Btn";
import type { ItemRowData } from "../ui/ItemRow";
import { ListDetailPage } from "../ui/screens/ListDetailPage";
import { TextField } from "../ui/TextField";
import {
	type ListEntry,
	type ListsDeps,
	useList,
	useListActions,
	useMyLists,
	useOnline,
} from "./data";

const toRow = (entry: ListEntry): ItemRowData => ({
	id: entry.item.id,
	title: entry.item.title,
	meta: entry.item.year === null ? "" : String(entry.item.year),
	coverUri: entry.item.cover?.thumb || null,
	score: null,
	scoreLabel: "",
});

export function ListDetailScreen({
	deps,
	id,
	locale,
	onOpenItem,
	onDeleted,
}: {
	deps: ListsDeps;
	id: string;
	locale: AppLocale;
	onOpenItem: (entry: ListEntry) => void;
	onDeleted: () => void;
}): ReactElement | null {
	const list = useList(deps, id);
	const mine = useMyLists(deps);
	const actions = useListActions(deps, locale);
	const online = useOnline();
	const [draft, setDraft] = useState<string | null>(null);

	const data = list.data;
	if (!data) return null;
	// GET /lists/{id} has no owner field; the viewer owns it when it is in GET /me/lists.
	const editable = (mine.data ?? []).some((l) => l.id === id);
	const byId = new Map(data.items.map((e) => [e.item.id, e]));
	const titled = (key: string) => (row: ItemRowData) => t(locale, key, { title: row.title });

	const confirmDelete = (): void => {
		Alert.alert(t(locale, "lists.delete"), t(locale, "lists.deleteConfirm"), [
			{ text: t(locale, "lists.cancel"), style: "cancel" },
			{
				text: t(locale, "lists.delete"),
				style: "destructive",
				onPress: () => void actions.remove(id).then((ok) => ok && onDeleted()),
			},
		]);
	};

	const saveTitle = async (): Promise<void> => {
		if (draft !== null && (await actions.rename(id, draft))) {
			setDraft(null);
			deps.notify(t(locale, "lists.saved"));
		}
	};

	return (
		<View className="flex-1 bg-bg dark:bg-bg-dark">
			{draft === null ? null : (
				<View className="gap-2 px-4 pt-4">
					<TextField label={t(locale, "lists.fieldTitle")} value={draft} onChangeText={setDraft} />
					<Btn
						label={t(locale, "lists.save")}
						onPress={() => void saveTitle()}
						disabled={!online || draft.trim() === ""}
					/>
				</View>
			)}
			<ListDetailPage
				kicker={t(locale, data.isPublic ? "lists.public" : "lists.private")}
				title={data.title}
				description={data.description || null}
				count={t(locale, "lists.itemCount", { count: data.items.length })}
				rows={data.items.map(toRow)}
				editable={editable}
				online={online}
				locale={locale}
				labels={{
					rename: t(locale, "lists.edit"),
					remove: t(locale, "lists.delete"),
					offline: t(locale, "lists.offline"),
					moveUp: titled("lists.moveUpLabel"),
					moveDown: titled("lists.moveDownLabel"),
					removeItem: titled("lists.removeLabel"),
					empty: t(locale, editable ? "lists.emptyOwner" : "lists.empty"),
				}}
				onOpen={(row) => {
					const entry = byId.get(row.id);
					if (entry) onOpenItem(entry);
				}}
				onRename={() => setDraft(data.title)}
				onDelete={confirmDelete}
				onMove={(row, delta) => void actions.move(data, row.id, delta)}
				onRemoveItem={(row) => void actions.removeItem(id, row.id)}
			/>
		</View>
	);
}
