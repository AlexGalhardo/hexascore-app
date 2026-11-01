import type { ReactElement } from "react";
import { Pressable, ScrollView, Text, View } from "react-native";
import { Btn } from "../Btn";

export interface NotificationRow {
	id: string;
	text: string;
	/** Already-formatted relative time. */
	time: string;
	unread: boolean;
}

export interface NotificationsLabels {
	unreadCount: (count: number) => string;
	markAll: string;
	/** Visible tag on unread rows, so colour is not the only signal. */
	unread: string;
	empty: string;
}

export interface NotificationsProps {
	title: string;
	rows: NotificationRow[];
	labels: NotificationsLabels;
	onOpen: (row: NotificationRow) => void;
	onMarkAll: () => void;
}

const FG = "text-fg dark:text-fg-dark";
const MUTED = "font-text text-14 text-muted dark:text-muted-dark";

function Row({
	row,
	labels,
	onOpen,
}: {
	row: NotificationRow;
	labels: NotificationsLabels;
	onOpen: (row: NotificationRow) => void;
}): ReactElement {
	return (
		<Pressable
			accessibilityRole="button"
			accessibilityLabel={[row.unread ? labels.unread : null, row.text, row.time]
				.filter(Boolean)
				.join(", ")}
			onPress={() => onOpen(row)}
			className="min-h-11 flex-row items-start gap-3 border-b border-border py-3 active:bg-surface dark:border-border-dark dark:active:bg-surface-dark"
		>
			<View
				className={`mt-2 h-2 w-2 rounded-full ${row.unread ? "bg-primary dark:bg-primary-dark" : "bg-transparent"}`}
			/>
			<View className="flex-1 gap-1">
				{row.unread ? (
					<Text className="font-text text-12 font-bold uppercase tracking-wider text-primary dark:text-primary-dark">
						{labels.unread}
					</Text>
				) : null}
				<Text className={`font-text text-16 ${row.unread ? "font-semibold" : ""} ${FG}`}>
					{row.text}
				</Text>
				<Text className={MUTED}>{row.time}</Text>
			</View>
		</Pressable>
	);
}

export function NotificationsPage({
	title,
	rows,
	labels,
	onOpen,
	onMarkAll,
}: NotificationsProps): ReactElement {
	const unread = rows.filter((row) => row.unread).length;
	return (
		<ScrollView
			className="flex-1 bg-bg dark:bg-bg-dark"
			contentContainerClassName="gap-4 px-4 pb-8 pt-6"
		>
			<View className="gap-2">
				<Text accessibilityRole="header" className={`font-display text-36 font-bold ${FG}`}>
					{title}
				</Text>
				<Text className={`${MUTED} tabular-nums`}>{labels.unreadCount(unread)}</Text>
			</View>
			<View className="self-start">
				<Btn
					variant="secondary"
					label={labels.markAll}
					onPress={onMarkAll}
					disabled={unread === 0}
				/>
			</View>
			{rows.length === 0 ? (
				<Text accessibilityLiveRegion="polite" className={`py-8 font-text text-16 ${FG}`}>
					{labels.empty}
				</Text>
			) : (
				<View className="border-t-2 border-fg dark:border-fg-dark">
					{rows.map((row) => (
						<Row key={row.id} row={row} labels={labels} onOpen={onOpen} />
					))}
				</View>
			)}
		</ScrollView>
	);
}
