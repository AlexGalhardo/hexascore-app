import type { ReactElement } from "react";
import { Pressable, ScrollView, Text, View } from "react-native";
import { Btn } from "../Btn";
import { ItemRow, type ItemRowData } from "../ItemRow";

export interface EntityRole {
	key: string;
	label: string;
	count: number;
}

export interface EntityFact {
	label: string;
	value: string;
}

export interface EntityGroup {
	key: string;
	title: string;
	/** Already-formatted "2 of 2". */
	count: string;
	note?: string;
	rows: ItemRowData[];
}

export interface EntityPageProps {
	kind: string;
	name: string;
	roles: EntityRole[];
	activeRole: string | null;
	facts: EntityFact[];
	description: string | null;
	licence: string;
	sourceLabel: string;
	/** In `ITEM_TYPES` order; the wiring decides it, empty groups are hidden here. */
	groups: EntityGroup[];
	empty: { title: string; body: string; action: string };
	locale: string;
	onRole: (key: string) => void;
	onOpen: (row: ItemRowData) => void;
	onSource: () => void;
	onBrowse: () => void;
}

const FG = "text-fg dark:text-fg-dark";
const MUTED = "font-text text-14 text-muted dark:text-muted-dark";

// v1 web-mobile entity/default and entity/empty.
export function EntityPage(props: EntityPageProps): ReactElement {
	const groups = props.groups.filter((group) => group.rows.length > 0);
	return (
		<ScrollView
			className="flex-1 bg-bg dark:bg-bg-dark"
			contentContainerClassName="gap-6 px-4 pb-8 pt-6"
		>
			<View className="gap-3">
				<Text className="font-text text-12 font-bold uppercase tracking-wider text-secondary dark:text-secondary-dark">
					{props.kind}
				</Text>
				<Text accessibilityRole="header" className={`font-display text-36 font-bold ${FG}`}>
					{props.name}
				</Text>
				{props.roles.length > 0 ? (
					<View className="flex-row flex-wrap gap-2 pt-1">
						{props.roles.map((role) => {
							const selected = role.key === props.activeRole;
							return (
								<Pressable
									key={role.key}
									accessibilityRole="button"
									accessibilityLabel={`${role.label} (${role.count})`}
									accessibilityState={{ selected }}
									onPress={() => props.onRole(role.key)}
									className={`min-h-11 flex-row items-center gap-1.5 rounded-sm border px-3 ${selected ? "border-fg bg-surface dark:border-fg-dark dark:bg-surface-dark" : "border-border dark:border-border-dark"}`}
								>
									<Text className={`font-text text-14 font-semibold ${FG}`}>{role.label}</Text>
									<Text className={`${MUTED} tabular-nums`}>({role.count})</Text>
								</Pressable>
							);
						})}
					</View>
				) : null}
			</View>
			{props.facts.length > 0 || props.description ? (
				<View className="border-t-2 border-fg pt-2 dark:border-fg-dark">
					{props.facts.map((fact) => (
						<View
							key={fact.label}
							className="flex-row gap-4 border-b border-border py-2 dark:border-border-dark"
						>
							<Text className={`w-24 ${MUTED}`}>{fact.label}</Text>
							<Text className={`flex-1 font-text text-14 ${FG}`}>{fact.value}</Text>
						</View>
					))}
					{props.description ? (
						<View className="gap-2 pt-3">
							<Text className={`font-display text-16 leading-relaxed ${FG}`}>
								{props.description}
							</Text>
							<View className="flex-row flex-wrap items-center gap-1">
								<Text className="font-text text-12 text-muted dark:text-muted-dark">
									{props.licence} ·
								</Text>
								<Pressable
									accessibilityRole="link"
									accessibilityLabel={props.sourceLabel}
									onPress={props.onSource}
									className="min-h-11 justify-center"
								>
									<Text className="font-text text-12 font-medium text-primary underline dark:text-primary-dark">
										{props.sourceLabel}
									</Text>
								</Pressable>
							</View>
						</View>
					) : null}
				</View>
			) : null}
			{groups.length === 0 ? (
				<View
					accessibilityLiveRegion="polite"
					className="items-center gap-3 border-t-2 border-fg py-12 dark:border-fg-dark"
				>
					<Text className="font-display text-24 italic text-muted dark:text-muted-dark">
						{props.empty.title}
					</Text>
					<Text className={`text-center ${MUTED}`}>{props.empty.body}</Text>
					<Btn variant="secondary" label={props.empty.action} onPress={props.onBrowse} />
				</View>
			) : (
				groups.map((group) => (
					<View key={group.key} className="gap-1">
						<View className="flex-row flex-wrap items-baseline gap-x-3 border-t-2 border-fg pt-4 dark:border-fg-dark">
							<Text accessibilityRole="header" className={`font-display text-24 font-bold ${FG}`}>
								{group.title}
							</Text>
							<Text className={`${MUTED} tabular-nums`}>{group.count}</Text>
						</View>
						{group.note ? <Text className={MUTED}>{group.note}</Text> : null}
						{group.rows.map((row) => (
							<ItemRow key={row.id} row={row} locale={props.locale} onOpen={props.onOpen} />
						))}
					</View>
				))
			)}
		</ScrollView>
	);
}
