// Base URL comes from E2E_BASE_URL, the variable Playwright reads (default http://127.0.0.1:3000).
const base = (process.env.E2E_BASE_URL ?? "http://127.0.0.1:3000").replace(/\/$/, "");
const paths = ["/", "/games", "/games/adipiscing-consectetur", "/books/consectetur-2004", "/people/ana-oliveira"];

module.exports = {
	"ci": {
		"collect": {
			"url": paths.map((p) => `${base}${p}`),
			"numberOfRuns": 3,
			"settings": {
				"formFactor": "mobile",
				"screenEmulation": {
					"mobile": true,
					"width": 412,
					"height": 823,
					"deviceScaleFactor": 1.75
				},
				"throttlingMethod": "simulate"
			}
		},
		"assert": {
			"aggregationMethod": "median",
			"assertions": {
				"categories:performance": [
					"error",
					{
						"minScore": 0.9
					}
				],
				"categories:accessibility": [
					"error",
					{
						"minScore": 1
					}
				],
				"categories:seo": [
					"error",
					{
						"minScore": 1
					}
				],
				"categories:best-practices": [
					"error",
					{
						"minScore": 0.95
					}
				],
				"largest-contentful-paint": [
					"error",
					{
						"maxNumericValue": 2000
					}
				],
				"total-blocking-time": [
					"error",
					{
						"maxNumericValue": 200
					}
				],
				"cumulative-layout-shift": [
					"error",
					{
						"maxNumericValue": 0.05
					}
				],
				"server-response-time": [
					"error",
					{
						"maxNumericValue": 400
					}
				],
				"resource-summary:script:size": [
					"error",
					{
						"maxNumericValue": 153600
					}
				],
				"resource-summary:stylesheet:size": [
					"error",
					{
						"maxNumericValue": 30720
					}
				],
				"resource-summary:font:count": [
					"error",
					{
						"maxNumericValue": 3
					}
				],
				"resource-summary:font:size": [
					"error",
					{
						"maxNumericValue": 102400
					}
				]
			}
		},
		"upload": {
			"target": "filesystem",
			"outputDir": ".lighthouseci"
		}
	}
};
