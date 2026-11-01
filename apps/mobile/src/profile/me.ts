import { ApiError, type Client } from "../../../../packages/app-client/src/api";
import type { ReadClient } from "../api/client";
import { type AppLocale, isAppLocale } from "../i18n/locale";

export type Me = {
	name: string;
	email: string;
	locale: AppLocale;
	profilePrivate: boolean;
	personalizationOptOut: boolean;
	/** Null until set; the server stores it only once. */
	username: string | null;
};

export type Device = { id: string; name: string; lastUsedAt: string | null; current: boolean };

export type ProfileApi = {
	read: Pick<ReadClient, "meShow" | "tokensIndex">;
	write<T>(call: (client: Pick<Client, "meUpdate" | "tokensRevoke">) => Promise<T>): Promise<T>;
};

export const meKey = ["meShow"] as const;
export const devicesKey = ["tokensIndex"] as const;

type Fields = Record<string, unknown>;
const isObject = (value: unknown): value is Fields => typeof value === "object" && value !== null;
const str = (value: unknown): string => (typeof value === "string" ? value : "");

// The generated client types GET /me and GET /tokens as unknown; narrow at the boundary.
export function parseMe(body: unknown): Me | null {
	if (!isObject(body) || typeof body.name !== "string") return null;
	const locale = str(body.locale);
	return {
		name: body.name,
		email: str(body.email),
		locale: isAppLocale(locale) ? locale : "en",
		profilePrivate: body.profilePrivate === true,
		personalizationOptOut: body.personalizationOptOut === true,
		username: typeof body.username === "string" ? body.username : null,
	};
}

export function parseDevices(body: unknown): Device[] {
	if (!Array.isArray(body)) return [];
	return body.filter(isObject).map((row) => ({
		id: str(row.id),
		name: str(row.name),
		lastUsedAt: typeof row.lastUsedAt === "string" ? row.lastUsedAt : null,
		current: row.current === true,
	}));
}

export type MeChange = Partial<Pick<Me, "name" | "locale" | "personalizationOptOut" | "username">>;

/** PATCH /me requires every field, so the change is merged over the loaded profile. */
export function patchBody(me: Me, change: MeChange): Fields {
	const next = { ...me, ...change };
	return {
		name: next.name,
		locale: next.locale,
		profilePrivate: next.profilePrivate,
		personalizationOptOut: next.personalizationOptOut,
		...(next.username ? { username: next.username } : {}),
	};
}

export function isNicknameTaken(error: unknown): boolean {
	if (!(error instanceof ApiError) || error.status !== 422) return false;
	return (
		(error.problem as { code?: unknown } | undefined)?.code === "nickname-taken" ||
		(error.errors.username ?? []).includes("nickname-taken")
	);
}
