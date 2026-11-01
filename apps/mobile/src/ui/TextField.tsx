import { type ReactElement, useState } from "react";
import { Pressable, Text, TextInput, type TextInputProps, View } from "react-native";

export interface TextFieldProps {
	label: string;
	value: string;
	onChangeText: (value: string) => void;
	/** Password field: hidden by default with a Show / Hide control. */
	secure?: { show: string; hide: string };
	error?: string;
	autoComplete?: TextInputProps["autoComplete"];
	keyboardType?: TextInputProps["keyboardType"];
}

const FG = "text-fg dark:text-fg-dark";

// Round 2 login/signup fields: visible label above, 44 px box, Show inside the password box.
export function TextField({
	label,
	value,
	onChangeText,
	secure,
	error,
	autoComplete,
	keyboardType,
}: TextFieldProps): ReactElement {
	const [hidden, setHidden] = useState(true);
	return (
		<View className="gap-2">
			<Text className={`font-text text-14 font-semibold ${FG}`}>{label}</Text>
			<View className="min-h-11 flex-row items-center rounded-sm border border-field bg-field-bg dark:border-field-dark dark:bg-field-bg-dark">
				<TextInput
					accessibilityLabel={label}
					autoCapitalize="none"
					autoComplete={autoComplete}
					keyboardType={keyboardType}
					secureTextEntry={secure !== undefined && hidden}
					value={value}
					onChangeText={onChangeText}
					className={`min-h-11 flex-1 px-3 font-text text-16 ${FG}`}
				/>
				{secure !== undefined && (
					<Pressable
						accessibilityRole="button"
						accessibilityLabel={hidden ? secure.show : secure.hide}
						onPress={() => setHidden(!hidden)}
						className="h-11 justify-center px-3"
					>
						<Text className="font-text text-14 font-semibold text-primary dark:text-primary-dark">
							{hidden ? secure.show : secure.hide}
						</Text>
					</Pressable>
				)}
			</View>
			{error !== undefined && (
				<Text
					accessibilityRole="alert"
					className="font-text text-14 text-danger dark:text-danger-dark"
				>
					{error}
				</Text>
			)}
		</View>
	);
}
