import { usePage } from "@inertiajs/react";

import { useT } from "../../lib/useT";
import type { SharedProps } from "../../types";

export interface Coverage {
	measuredAt: string | null;
	rows: { type: string; field: string; items: number; pct: number }[];
}

const TYPES = ["game", "movie", "series", "book"];

// Percentages from resources/data/api_coverage.json (qa/sql/api_coverage.sql), shown on /pricing and /docs/coverage.
export default function CoverageTable({ coverage }: { coverage: Coverage }) {
	const t = useT();
	const { locale } = usePage<SharedProps>().props;

	if (!coverage.measuredAt || coverage.rows.length === 0) {
		return <p className="mt-4 text-muted">{t("coverage.notMeasured")}</p>;
	}

	const percent = new Intl.NumberFormat(locale, { style: "percent", maximumFractionDigits: 1 });
	const types = TYPES.filter((type) => coverage.rows.some((r) => r.type === type));
	const fields = [...new Set(coverage.rows.map((r) => r.field))];
	const pct = (type: string, field: string) =>
		coverage.rows.find((r) => r.type === type && r.field === field)?.pct;

	return (
		<div className="mt-4 max-w-3xl overflow-x-auto">
			<table className="w-full border-collapse text-sm">
				<caption className="mb-2 text-left text-muted">
					{t("coverage.caption", { date: coverage.measuredAt })}
				</caption>
				<thead>
					<tr>
						<th scope="col" className="border-b-2 border-fg py-2 pr-4 text-left">
							{t("coverage.field")}
						</th>
						{types.map((type) => (
							<th
								key={type}
								scope="col"
								className="border-b-2 border-fg py-2 pr-4 text-right"
							>
								{t(`catalog.title.${type}`)}
							</th>
						))}
					</tr>
				</thead>
				<tbody>
					{fields.map((field) => (
						<tr key={field}>
							<th
								scope="row"
								className="border-b border-border py-2 pr-4 text-left font-normal"
							>
								<code className="font-mono">{field}</code>
							</th>
							{types.map((type) => {
								const value = pct(type, field);
								return (
									<td
										key={type}
										className="border-b border-border py-2 pr-4 text-right tabular-nums"
									>
										{value === undefined ? "–" : percent.format(value / 100)}
									</td>
								);
							})}
						</tr>
					))}
				</tbody>
			</table>
		</div>
	);
}
