import { Link } from "@inertiajs/react";
import { useT } from "../../lib/useT";

// Entry point to the create-request form; the page itself shows the closed notice while the flag is off.
export default function SuggestNewLink({ domain }: { domain: string }) {
	const t = useT();

	return (
		<Link href={`/suggest/new/${domain}`} className="underline">
			{t("community.create.link", { type: t(`itemTypes.${domain}`) })}
		</Link>
	);
}
