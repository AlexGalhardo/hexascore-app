import SuggestBox from "../search/SuggestBox";

interface Props {
	label: string;
	domain: string;
	id: string;
	title: string;
	changeLabel: string;
	onPick: (id: string, title: string) => void;
}

// ItemPicker chooses one catalog record by search; a draft restored with only an id shows the id.
export default function ItemPicker({ label, domain, id, title, changeLabel, onPick }: Props) {
	return (
		<div className="block">
			<span className="font-medium">{label}</span>
			{id === "" ? (
				<div className="mt-1">
					<SuggestBox type={domain} label={label} onPick={(s) => onPick(s.id, s.title)} />
				</div>
			) : (
				<p className="mt-1 flex items-center gap-2">
					<span>{title || id}</span>
					<button
						type="button"
						onClick={() => onPick("", "")}
						className="rounded-md border border-border px-2 py-1 text-sm"
					>
						{changeLabel}
					</button>
				</p>
			)}
		</div>
	);
}
