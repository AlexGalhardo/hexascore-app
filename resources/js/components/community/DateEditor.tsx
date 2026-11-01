import { useId } from "react";
import type { DateValue } from "./fields";

interface Props {
	label: string;
	current: DateValue | null;
	value: DateValue;
	lockReason?: string;
	lockedLabel: string;
	precisionLabel: string;
	precisionName: (p: string) => string;
	onChange: (value: DateValue) => void;
}

const inputClass =
	"mt-1 block rounded-md border border-zinc-300 bg-transparent px-3 py-2 disabled:opacity-60 dark:border-zinc-700";

// DateEditor edits a release date and its precision; the day input holds the full ISO date for every precision.
export default function DateEditor({
	label,
	current,
	value,
	lockReason,
	lockedLabel,
	precisionLabel,
	precisionName,
	onChange,
}: Props) {
	const reasonId = useId();
	const inputId = useId();
	const disabled = lockReason !== undefined;
	const describedBy = disabled ? reasonId : undefined;

	return (
		<div className="block">
			<label htmlFor={inputId} className="font-medium">
				{label}
			</label>
			<div className="flex gap-2">
				<input
					id={inputId}
					type="date"
					value={value.value}
					disabled={disabled}
					aria-describedby={describedBy}
					onChange={(e) => onChange({ ...value, value: e.target.value })}
					className={inputClass}
				/>
				<select
					aria-label={`${precisionLabel} ${label}`}
					value={value.precision}
					disabled={disabled}
					aria-describedby={describedBy}
					onChange={(e) => onChange({ ...value, precision: e.target.value })}
					className={inputClass}
				>
					{["day", "month", "year"].map((p) => (
						<option key={p} value={p}>
							{precisionName(p)}
						</option>
					))}
				</select>
			</div>
			{current && current.value !== value.value && (
				<span className="mt-1 block text-xs text-muted">{current.value}</span>
			)}
			{disabled && (
				<span id={reasonId} className="mt-1 block text-sm text-red-700 dark:text-red-400">
					{lockedLabel} {lockReason}
				</span>
			)}
		</div>
	);
}
