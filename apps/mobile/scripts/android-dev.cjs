// One-command local Android debug run (docs/runbooks/mobile-local-android.md).
// The emulator reaches the host loopback at 10.0.2.2; the dev backend (`vukem serve`) is not
// production, so CanonicalHost does not redirect it. Debug builds allow cleartext http; release does not.
const { spawnSync } = require("node:child_process");
const { cpSync, existsSync, mkdirSync, readFileSync, writeFileSync } = require("node:fs");
const path = require("node:path");

const mobileDir = path.resolve(__dirname, "..");
const env = {
	...process.env,
	EXPO_PUBLIC_API_BASE_URL: process.env.EXPO_PUBLIC_API_BASE_URL || "http://10.0.2.2:3000",
	// Gradle 9 can deadlock resolving artifacts in parallel on this project.
	GRADLE_OPTS: process.env.GRADLE_OPTS || "-Dorg.gradle.parallel=false",
};

const sdk = process.env.ANDROID_HOME || process.env.ANDROID_SDK_ROOT;
if (sdk) env.PATH = `${path.join(sdk, "platform-tools")}${path.delimiter}${env.PATH}`;

const run = (cmd, args, cwd) => {
	const r = spawnSync(cmd, args, { stdio: "inherit", shell: true, cwd, env });
	if (r.status !== 0) process.exit(r.status ?? 1);
};

if (process.platform !== "win32") {
	run("bunx", ["expo", "run:android", "--variant", "debug", ...process.argv.slice(2)], mobileDir);
	process.exit(0);
}

// Windows: ninja (SDK CMake) cannot stat the 260+ character paths of bun's isolated node_modules.
// The native build therefore runs in a short-path mirror with a hoisted install (C:\h by default, as short as possible);
// only the native config is mirrored. JS is served by Metro from the real checkout, so edits need no sync.
const mirror = process.env.HEXASCORE_ANDROID_DIR || path.join(path.parse(mobileDir).root, "h");
mkdirSync(mirror, { recursive: true });
for (const item of ["app.config.ts", "assets", "package.json", "tsconfig.json"]) {
	cpSync(path.join(mobileDir, item), path.join(mirror, item), { recursive: true });
}
const stamp = path.join(mirror, ".installed-package.json");
const pkg = readFileSync(path.join(mobileDir, "package.json"), "utf8");
if (
	!existsSync(path.join(mirror, "node_modules")) ||
	!existsSync(stamp) ||
	readFileSync(stamp, "utf8") !== pkg
) {
	run("bun", ["install", "--linker", "hoisted"], mirror);
	writeFileSync(stamp, pkg);
}
if (!existsSync(path.join(mirror, "android"))) {
	run("bunx", ["expo", "prebuild", "--platform", "android", "--no-install"], mirror);
}
run(
	"bunx",
	["expo", "run:android", "--variant", "debug", "--no-bundler", ...process.argv.slice(2)],
	mirror,
);

run("adb", ["reverse", "tcp:8081", "tcp:8081"], mobileDir);
run("bunx", ["expo", "start"], mobileDir);
