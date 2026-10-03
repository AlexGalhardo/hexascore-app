import { type Locale, routing } from "@/i18n/routing";

/** Public path of the OpenAPI document for a locale. English keeps `/openapi.json` (the SDKs generate from it). */
export function openApiPath(locale: Locale): string {
	return locale === routing.defaultLocale ? "/openapi.json" : `/openapi.${locale}.json`;
}
