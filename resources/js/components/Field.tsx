import type { InputHTMLAttributes } from "react";

import { cn } from "../lib/utils";
import { inputClass } from "./ui/input";

interface Props extends InputHTMLAttributes<HTMLInputElement> {
	label: string;
	error?: string;
}

export const fieldClass = cn(inputClass, "mt-2 h-11");

// Field is a labelled input with its validation error.
export default function Field({ label, error, ...input }: Props) {
	return (
		<label className="block">
			<span className="text-sm font-semibold">{label}</span>
			<input aria-invalid={error ? true : undefined} {...input} className={fieldClass} />
			{error && <p className="mt-1.5 text-sm text-danger">{error}</p>}
		</label>
	);
}
