const { statSync } = require("node:fs");

// expo export:embed has been seen to exit 0 after printing a resolution error,
// so the bundle file itself is the proof: it must exist and be non-empty.
const file = process.argv[2];
let size = 0;
try {
	size = statSync(file).size;
} catch {}
if (size === 0) {
	console.error(`bundle check failed: ${file} is missing or empty`);
	process.exit(1);
}
console.log(`bundle check ok: ${file} (${size} bytes)`);
