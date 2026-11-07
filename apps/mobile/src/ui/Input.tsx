import { type ReactElement, useState } from "react";
import { TextInput, type TextInputProps } from "react-native";

export interface InputProps extends Omit<TextInputProps, "className" | "style"> {
	label: string;
	invalid?: boolean;
	className?: string;
}

// shadcn/ui input: 44 pt box, hairline field border, 2 px ring colour while focused.
export function Input({
	label,
	invalid = false,
	className = "",
	...rest
}: InputProps): ReactElement {
	const [focused, setFocused] = useState(false);
	const border = invalid
		? "border-danger dark:border-danger-dark"
		: focused
			? "border-ring dark:border-ring-dark"
			: "border-field dark:border-field-dark";
	return (
		<TextInput
			accessibilityLabel={label}
			placeholderTextColor="#8a8a8a"
			{...rest}
			onFocus={(event) => {
				setFocused(true);
				rest.onFocus?.(event);
			}}
			onBlur={(event) => {
				setFocused(false);
				rest.onBlur?.(event);
			}}
			className={`min-h-11 rounded-md border bg-field-bg px-3 font-text text-16 text-fg dark:bg-field-bg-dark dark:text-fg-dark ${focused ? "border-2" : ""} ${border} ${className}`}
		/>
	);
}
