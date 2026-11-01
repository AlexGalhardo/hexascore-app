import type { InputHTMLAttributes } from "react";

interface Props extends InputHTMLAttributes<HTMLInputElement> {
	label: string;
	error?: string;
}

export const fieldClass =
	"mt-2 block min-h-12 w-full rounded-none border border-field bg-field-bg px-3.5 text-base text-fg placeholder:text-muted focus-visible:outline-2 focus-visible:outline-offset-0 focus-visible:outline-primary aria-invalid:border-danger read-only:bg-surface read-only:text-muted";

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
