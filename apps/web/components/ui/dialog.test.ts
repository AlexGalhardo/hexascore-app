import { describe, expect, test } from "bun:test";
import { Glob } from "bun";

// Project rule: every modal is centred (components/ui/dialog.ts). A Base UI popup or backdrop styled by hand
// would drift back to "stuck at the top" or "under the navbar", so the source is checked here.
describe("dialog layout rule", () => {
	test("every Dialog/AlertDialog popup and backdrop uses the shared classes", async () => {
		const offenders: string[] = [];
		for await (const file of new Glob("{app,components}/**/*.tsx").scan({ cwd: `${import.meta.dir}/../..` })) {
			const source = await Bun.file(`${import.meta.dir}/../../${file}`).text();
			for (const match of source.matchAll(/<(?:Alert)?Dialog\.(Popup|Backdrop)\b([^>]*)>/g)) {
				const shared = match[1] === "Popup" ? "dialogPopup(" : "DIALOG_BACKDROP";
				if (!match[2]?.includes(shared)) offenders.push(`${file}: Dialog.${match[1]}`);
			}
		}
		expect(offenders).toEqual([]);
	});
});
