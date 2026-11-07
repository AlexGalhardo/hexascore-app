import { Pressable, Text, View } from "react-native";

export interface SegmentedOption<T extends string> {
	value: T;
	label: string;
}

export interface SegmentedControlProps<T extends string> {
	options: SegmentedOption<T>[];
	value: T;
	onChange: (value: T) => void;
	label: string;
}

// shadcn tabs list: a muted track with the selected segment raised on the page colour.
export function SegmentedControl<T extends string>({
	options,
	value,
	onChange,
	label,
}: SegmentedControlProps<T>): React.ReactElement {
	return (
		<View
			accessibilityRole="tablist"
			accessibilityLabel={label}
			className="flex-row rounded-lg bg-hover p-1 dark:bg-hover-dark"
		>
			{options.map((option) => {
				const selected = option.value === value;
				return (
					<Pressable
						key={option.value}
						accessibilityRole="tab"
						accessibilityState={{ selected }}
						onPress={() => onChange(option.value)}
						className={`min-h-9 flex-1 items-center justify-center rounded-md px-3 ${selected ? "bg-bg dark:bg-bg-dark" : ""}`}
					>
						<Text
							className={`font-text text-14 font-semibold ${selected ? "text-fg dark:text-fg-dark" : "text-muted dark:text-muted-dark"}`}
						>
							{option.label}
						</Text>
					</Pressable>
				);
			})}
		</View>
	);
}
