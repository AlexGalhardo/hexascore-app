import { type ReactElement, type ReactNode, useId } from "react";

/** A visible label bound to its control, with an optional hint under it. The control gets its id from render. */
export function Field({
	label,
	hint,
	children,
	className,
}: {
	label: string;
	hint?: string;
	children: (id: string) => ReactNode;
	className?: string;
}): ReactElement {
	const id = useId();
	return (
		<div
			className={
				className ? `grid content-start gap-1.5 ${className}` : "grid content-start gap-1.5"
			}
		>
			<label htmlFor={id} className="text-sm font-medium">
				{label}
			</label>
			{children(id)}
			{hint && <p className="text-xs text-muted">{hint}</p>}
		</div>
	);
}
