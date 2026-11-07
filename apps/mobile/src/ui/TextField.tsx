import { type ReactElement, useState } from "react";
import { Pressable, Text, type TextInputProps, View } from "react-native";
import { Input } from "./Input";
import { Label } from "./Label";

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

// Label above, Input, optional Show / Hide inside the password box, inline error below.
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
			<Label text={label} />
			<View className="justify-center">
				<Input
					label={label}
					autoCapitalize="none"
					autoComplete={autoComplete}
					keyboardType={keyboardType}
					secureTextEntry={secure !== undefined && hidden}
					value={value}
					onChangeText={onChangeText}
					invalid={error !== undefined}
					className={secure ? "pr-20" : ""}
				/>
				{secure !== undefined && (
					<Pressable
						accessibilityRole="button"
						accessibilityLabel={hidden ? secure.show : secure.hide}
						onPress={() => setHidden(!hidden)}
						className="absolute right-0 h-11 justify-center px-3"
					>
						<Text className="font-text text-14 font-semibold text-fg dark:text-fg-dark">
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
