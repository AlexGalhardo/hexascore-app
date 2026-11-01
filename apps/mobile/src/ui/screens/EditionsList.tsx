import type { ReactElement } from "react";
import { Text, View } from "react-native";

export interface EditionFact {
	label: string;
	value: string;
}

export interface Edition {
	id: string;
	title: string;
	facts: EditionFact[];
}

export interface EditionsListProps {
	title: string;
	editions: Edition[];
}

const FG = "text-fg dark:text-fg-dark";

// Block for the book item screen (`/books/{id}/editions`); rows follow web-mobile edition/default.
export function EditionsList({ title, editions }: EditionsListProps): ReactElement | null {
	if (editions.length === 0) return null;
	return (
		<View className="gap-2">
			<Text accessibilityRole="header" className={`font-display text-24 font-bold ${FG}`}>
				{title}
			</Text>
			{editions.map((edition) => (
				<View key={edition.id} className="border-t-2 border-fg pt-2 dark:border-fg-dark">
					<Text className={`font-text text-16 font-semibold ${FG}`}>{edition.title}</Text>
					{edition.facts.map((fact) => (
						<View
							key={fact.label}
							className="flex-row gap-4 border-b border-border py-2 dark:border-border-dark"
						>
							<Text className="w-24 font-text text-14 text-muted dark:text-muted-dark">
								{fact.label}
							</Text>
							<Text className={`flex-1 font-text text-14 tabular-nums ${FG}`}>{fact.value}</Text>
						</View>
					))}
				</View>
			))}
		</View>
	);
}
