import type { ButtonHTMLAttributes } from "react";

import { cn } from "../../lib/utils";
import { buttonVariants } from "./button";

export type BtnKind = "primary" | "secondary" | "ghost";

// The pages keep their Btn/btnClass call sites; both now render the shadcn button (components/ui/button.tsx).
const VARIANT = { primary: "default", secondary: "outline", ghost: "ghost" } as const;

/** Class names for a button or a link that looks like one. */
export function btnClass(kind: BtnKind = "primary", extra = ""): string {
	return cn(buttonVariants({ variant: VARIANT[kind], size: "lg" }), extra);
}

interface Props extends ButtonHTMLAttributes<HTMLButtonElement> {
	kind?: BtnKind;
}

export default function Btn({ kind = "primary", className = "", type = "button", ...rest }: Props) {
	return <button {...rest} type={type} className={btnClass(kind, className)} />;
}
