import type { ReactElement } from "react";
import { Pressable, ScrollView, Text, View } from "react-native";
import { Btn } from "../Btn";
import { Icon, type IconName } from "../Icon";
import { ItemRow, type ItemRowData } from "../ItemRow";

export interface ListDetailLabels {
	rename: string;
	remove: string;
	/** Shown under the owner controls while offline. */
	offline: string;
	moveUp: (row: ItemRowData) => string;
	moveDown: (row: ItemRowData) => string;
	removeItem: (row: ItemRowData) => string;
	empty: string;
}

export interface ListDetailProps {
	kicker: string;
	title: string;
	description: string | null;
	count: string;
	rows: ItemRowData[];
	/** The viewer owns the list. */
	editable: boolean;
	online: boolean;
	labels: ListDetailLabels;
	locale: string;
	onOpen: (row: ItemRowData) => void;
	onRename: () => void;
	onDelete: () => void;
	onMove: (row: ItemRowData, delta: -1 | 1) => void;
	onRemoveItem: (row: ItemRowData) => void;
}

const FG = "text-fg dark:text-fg-dark";
const MUTED = "font-text text-14 text-muted dark:text-muted-dark";

function IconBtn({
	name,
	label,
	disabled,
	onPress,
}: {
	name: IconName;
	label: string;
	disabled: boolean;
	onPress: () => void;
}): ReactElement {
	return (
		<Pressable
			accessibilityRole="button"
			accessibilityLabel={label}
			accessibilityState={{ disabled }}
			disabled={disabled}
			onPress={onPress}
			className={`h-11 w-11 items-center justify-center rounded-sm active:bg-surface dark:active:bg-surface-dark ${disabled ? "opacity-50" : ""}`}
		>
			<Icon name={name} size={20} className={FG} />
		</Pressable>
	);
}

// SDD 10 §6.1 Lists: create, rename, delete and reorder need the network; a single remove is queued.
export function ListDetailPage(props: ListDetailProps): ReactElement {
	const { labels, rows, editable, online } = props;
	return (
		<ScrollView
			className="flex-1 bg-bg dark:bg-bg-dark"
			contentContainerClassName="gap-6 px-4 pb-8 pt-6"
		>
			<View className="gap-2">
				<Text className="font-text text-12 font-bold uppercase tracking-wider text-primary dark:text-primary-dark">
					{props.kicker}
				</Text>
				<Text accessibilityRole="header" className={`font-display text-36 font-bold ${FG}`}>
					{props.title}
				</Text>
				{props.description ? (
					<Text className={`font-text text-16 ${FG}`}>{props.description}</Text>
				) : null}
				<Text className={`${MUTED} tabular-nums`}>{props.count}</Text>
			</View>
			{editable ? (
				<View className="gap-2">
					<View className="flex-row flex-wrap gap-2">
						<Btn
							variant="secondary"
							label={labels.rename}
							onPress={props.onRename}
							disabled={!online}
						/>
						<Btn
							variant="ghost"
							label={labels.remove}
							onPress={props.onDelete}
							disabled={!online}
						/>
					</View>
					{online ? null : <Text className={MUTED}>{labels.offline}</Text>}
				</View>
			) : null}
			{rows.length === 0 ? (
				<Text accessibilityLiveRegion="polite" className={`py-8 font-text text-16 ${FG}`}>
					{labels.empty}
				</Text>
			) : (
				<View className="border-t-2 border-fg dark:border-fg-dark">
					{rows.map((row, index) => (
						<ItemRow
							key={row.id}
							row={row}
							locale={props.locale}
							onOpen={props.onOpen}
							trailing={
								editable ? (
									<View className="flex-row">
										<IconBtn
											name="arrow-up"
											label={labels.moveUp(row)}
											disabled={!online || index === 0}
											onPress={() => props.onMove(row, -1)}
										/>
										<IconBtn
											name="arrow-down"
											label={labels.moveDown(row)}
											disabled={!online || index === rows.length - 1}
											onPress={() => props.onMove(row, 1)}
										/>
										<IconBtn
											name="close"
											label={labels.removeItem(row)}
											disabled={false}
											onPress={() => props.onRemoveItem(row)}
										/>
									</View>
								) : null
							}
						/>
					))}
				</View>
			)}
		</ScrollView>
	);
}
