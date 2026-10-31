import { useId } from "react";

export interface RelationDraft {
	add: string;
	remove: string;
	role: string;
}

interface Props {
	label: string;
	roles: string[];
	value: RelationDraft;
	lockReason?: string;
	lockedLabel: string;
	addLabel: string;
	removeLabel: string;
	roleLabel: string;
	onChange: (value: RelationDraft) => void;
}

const inputClass =
	"mt-1 block w-full rounded-md border border-zinc-300 bg-transparent px-3 py-2 disabled:opacity-60 dark:border-zinc-700";

// RelationEditor collects entity ids to add and remove, one per line (no entity search endpoint exists yet).
export default function RelationEditor({
	label,
	roles,
	value,
	lockReason,
	lockedLabel,
	addLabel,
	removeLabel,
	roleLabel,
	onChange,
}: Props) {
	const reasonId = useId();
	const disabled = lockReason !== undefined;
	const common = {
		disabled,
		"aria-describedby": disabled ? reasonId : undefined,
		className: inputClass,
	};

	return (
		<fieldset className="block">
			<legend className="font-medium">{label}</legend>
			{roles.length > 0 && (
				<select
					aria-label={roleLabel}
					value={value.role || roles[0]}
					onChange={(e) => onChange({ ...value, role: e.target.value })}
					{...common}
				>
					{roles.map((r) => (
						<option key={r} value={r}>
							{r}
						</option>
					))}
				</select>
			)}
			<textarea
				aria-label={addLabel}
				rows={2}
				value={value.add}
				onChange={(e) => onChange({ ...value, add: e.target.value })}
				{...common}
			/>
			<textarea
				aria-label={removeLabel}
				rows={2}
				value={value.remove}
				onChange={(e) => onChange({ ...value, remove: e.target.value })}
				{...common}
			/>
			{disabled && (
				<span id={reasonId} className="mt-1 block text-sm text-red-700 dark:text-red-400">
					{lockedLabel} {lockReason}
				</span>
			)}
		</fieldset>
	);
}
