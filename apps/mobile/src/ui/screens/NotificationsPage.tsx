import type { ReactElement } from "react";
import { ScrollView, Text, View } from "react-native";
import { Badge } from "../Badge";
import { Button } from "../Button";
import { EmptyState } from "../EmptyState";
import { ListItem } from "../ListItem";

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
		<ListItem
			title={row.text}
			subtitle={row.time}
			accessibilityLabel={[row.unread ? labels.unread : null, row.text, row.time]
				.filter(Boolean)
				.join(", ")}
			onPress={() => onOpen(row)}
			leading={
				<View
					className={`h-2 w-2 rounded-full ${row.unread ? "bg-primary dark:bg-primary-dark" : "bg-transparent"}`}
				/>
			}
			trailing={row.unread ? <Badge label={labels.unread} /> : undefined}
		/>
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
				<Text
					accessibilityRole="header"
					className={`font-display text-36 font-semibold tracking-tight ${FG}`}
				>
					{title}
				</Text>
				<Text className={`${MUTED} tabular-nums`}>{labels.unreadCount(unread)}</Text>
			</View>
			<View className="self-start">
				<Button
					variant="outline"
					label={labels.markAll}
					onPress={onMarkAll}
					disabled={unread === 0}
				/>
			</View>
			{rows.length === 0 ? (
				<EmptyState title={labels.empty} />
			) : (
				<View className="border-t border-border dark:border-border-dark">
					{rows.map((row) => (
						<Row key={row.id} row={row} labels={labels} onOpen={onOpen} />
					))}
				</View>
			)}
		</ScrollView>
	);
}
