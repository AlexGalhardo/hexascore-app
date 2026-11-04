import { CircleAlert } from "lucide-react";
import type { ReactElement } from "react";

import { Alert } from "./ui/alert";

/** The message of a refused admin action, announced to assistive tech; renders nothing without one. */
export function ActionError({ message }: { message: string | null }): ReactElement | null {
	if (!message) return null;
	return (
		<Alert variant="destructive">
			<CircleAlert aria-hidden="true" className="mt-0.5 size-4 shrink-0" />
			<span>{message}</span>
		</Alert>
	);
}
