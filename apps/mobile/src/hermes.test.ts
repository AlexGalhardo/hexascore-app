import config from "../app.config";

test("HermesInternal is defined", () => {
	const runtime = globalThis as { HermesInternal?: object };
	expect(runtime.HermesInternal).toBeDefined();
	expect(config.android?.package).toBe("app.hexascore");
	expect(config.ios).toBeUndefined();
	expect(config.experiments?.typedRoutes).toBe(true);
});
