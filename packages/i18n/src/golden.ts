// Prints translate() results for the cases on stdin; app/i18n/plural_test.go
// compares them with the Go side (TestPluralGolden).
import { readFileSync } from "node:fs";

import { type Message, translate } from "../../../resources/js/lib/i18n";

interface Case {
  locale: string;
  key: string;
  count: number;
}

type Tree = { [k: string]: string | Tree };
const PLURAL = new Set(["zero", "one", "two", "few", "many", "other"]);

function flatten(tree: Tree, prefix: string, out: Record<string, Message>): Record<string, Message> {
  for (const [k, v] of Object.entries(tree)) {
    const key = prefix ? `${prefix}.${k}` : k;
    if (typeof v === "string") out[key] = v;
    else if (Object.keys(v).every((c) => PLURAL.has(c))) out[key] = v as Record<string, string>;
    else flatten(v, key, out);
  }
  return out;
}

const cases = JSON.parse(readFileSync(0, "utf8")) as Case[];
const catalogs = new Map<string, Record<string, Message>>();
const results = cases.map((c) => {
  if (!catalogs.has(c.locale)) {
    catalogs.set(c.locale, flatten(JSON.parse(readFileSync(`resources/lang/${c.locale}.json`, "utf8")) as Tree, "", {}));
  }
  return translate(catalogs.get(c.locale) ?? {}, c.locale, c.key, { count: c.count });
});
console.log(JSON.stringify(results));
