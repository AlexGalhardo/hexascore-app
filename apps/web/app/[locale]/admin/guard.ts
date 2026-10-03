import { notFound } from "next/navigation";
import { isAdmin } from "@/lib/member";
import { getSessionUser } from "@/lib/session";

export async function requireAdmin(): Promise<void> {
	if (!isAdmin(await getSessionUser())) notFound();
}

export const cell = "border-b border-border px-2 py-1.5 align-top";
export const head = "border-b border-border px-2 py-1.5 text-left text-xs font-semibold text-muted";
export const input = "min-h-9 rounded border border-border bg-bg px-2 text-sm";
export const button = "min-h-9 rounded border border-border px-2 text-sm hover:bg-surface";
