import * as Haptics from "expo-haptics";
import type { ReactElement } from "react";
import { RatingInput, type RatingInputProps } from "../ui/RatingInput";

/** impl-mobile's RatingInput plus the behaviour SDD 10 §7.5 asks for: re-tap is a no-op, one selection haptic per change. */
export function StarInput({ value, onChange, ...rest }: RatingInputProps): ReactElement {
	const change = (next: number): void => {
		if (next === value) return;
		void Haptics.selectionAsync();
		onChange(next);
	};
	return <RatingInput {...rest} value={value} onChange={change} />;
}
