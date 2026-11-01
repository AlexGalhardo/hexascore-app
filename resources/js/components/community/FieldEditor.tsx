import { useId } from "react";

interface Props {
	label: string;
	current: string;
	value: string;
	multiline?: boolean;
	lockReason?: string;
	lockedLabel: string;
	onChange: (value: string) => void;
}

// FieldEditor edits one suggested value; a locked field stays visible but disabled, with its reason linked.
export default function FieldEditor({
	label,
	current,
	value,
	multiline,
	lockReason,
	lockedLabel,
	onChange,
}: Props) {
	const reasonId = useId();
	const inputId = useId();
	const common = {
		id: inputId,
		value,
		disabled: lockReason !== undefined,
		"aria-describedby": lockReason !== undefined ? reasonId : undefined,
		onChange: (e: { target: { value: string } }) => onChange(e.target.value),
		className:
			"mt-1 block w-full rounded-md border border-zinc-300 bg-transparent px-3 py-2 disabled:opacity-60 dark:border-zinc-700",
	};

	return (
		<div className="block">
			<label htmlFor={inputId} className="font-medium">
				{label}
			</label>
			{multiline ? <textarea rows={4} {...common} /> : <input type="text" {...common} />}
			{current !== "" && current !== value && (
				<span className="mt-1 block text-xs text-muted">{current}</span>
			)}
			{lockReason !== undefined && (
				<span id={reasonId} className="mt-1 block text-sm text-red-700 dark:text-red-400">
					{lockedLabel} {lockReason}
				</span>
			)}
		</div>
	);
}
