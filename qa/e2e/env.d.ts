// shortcut: minimal Node typing for the Playwright config; replace with @types/node if it is ever approved (D-22).
declare const process: { env: Record<string, string | undefined> };

declare module "node:child_process" {
	export function spawnSync(
		command: string,
		args: string[],
		options: { stdio?: "ignore"; encoding?: "utf8"; maxBuffer?: number },
	): { status: number | null; stdout: string; stderr: string };
}
declare module "node:fs" {
	export function readFileSync(path: string, encoding: "utf8"): string;
}
declare module "node:url" {
	export function fileURLToPath(url: URL): string;
}
declare module "node:crypto" {
	export function createHmac(
		algorithm: "sha1",
		key: Uint8Array,
	): { update(data: Uint8Array): { digest(): Uint8Array } };
}
