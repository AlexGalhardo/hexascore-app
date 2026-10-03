import { notFound } from "next/navigation";

// Unknown paths under a locale render the localized not-found page instead of the root one.
export default function CatchAll(): never {
	notFound();
}
