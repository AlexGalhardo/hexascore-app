import { type ReactElement, useState } from "react";
import { type AccessibilityActionEvent, Pressable, Text, View } from "react-native";
import { formatScore } from "./format";

const STARS = 10;
const MIN = 0.5;

const POSITIONS = Array.from({ length: STARS }, (_, index) => index + 1);

const clamp = (value: number): number => Math.min(STARS, Math.max(MIN, value));

/** Touch x inside a row `width` wide to a score in 0.5 steps (each star is two halves). */
export function valueAt(x: number, width: number): number {
	return clamp(Math.ceil((x / width) * STARS * 2) / 2);
}

export interface RatingInputProps {
	label: string;
	value: number | null;
	onChange: (value: number) => void;
	locale: string;
}

function Star({ fill }: { fill: 0 | 0.5 | 1 }): ReactElement {
	return (
		<View className="h-11 flex-1 items-center justify-center">
			<Text className="text-24 text-star-line dark:text-star-line-dark">☆</Text>
			{fill > 0 && (
				<View
					className={`absolute bottom-0 left-0 top-0 items-start justify-center overflow-hidden ${fill === 1 ? "w-full" : "w-1/2"}`}
				>
					<View className="w-full items-center">
						<Text className="text-24 text-star dark:text-star-dark">★</Text>
					</View>
				</View>
			)}
		</View>
	);
}

// Round 2 item 6: ten stars, half steps, value printed beside, 44 px row.
// TalkBack sees one adjustable control (swipe up/down = ±0.5) instead of 20 radios.
export function RatingInput({ label, value, onChange, locale }: RatingInputProps): ReactElement {
	const [width, setWidth] = useState(0);
	const text = `${formatScore(value, locale)} / ${STARS}`;
	const onAction = (event: AccessibilityActionEvent): void => {
		const step = event.nativeEvent.actionName === "increment" ? MIN : -MIN;
		onChange(clamp((value ?? 0) + step));
	};
	return (
		<View
			accessible
			accessibilityRole="adjustable"
			accessibilityLabel={label}
			accessibilityValue={{ min: 0, max: STARS, now: value ?? 0, text }}
			accessibilityActions={[{ name: "increment" }, { name: "decrement" }]}
			onAccessibilityAction={onAction}
			className="flex-row items-center gap-3"
		>
			<Pressable
				testID="rating-stars"
				importantForAccessibility="no-hide-descendants"
				onLayout={(event) => setWidth(event.nativeEvent.layout.width)}
				onPress={(event) => width > 0 && onChange(valueAt(event.nativeEvent.locationX, width))}
				className="flex-1 flex-row"
			>
				{POSITIONS.map((position) => {
					const filled = (value ?? 0) - position + 1;
					return <Star key={position} fill={filled >= 1 ? 1 : filled >= 0.5 ? 0.5 : 0} />;
				})}
			</Pressable>
			<Text className="w-20 text-right text-16 font-semibold tabular-nums text-fg dark:text-fg-dark">
				{text}
			</Text>
		</View>
	);
}
