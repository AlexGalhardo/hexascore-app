// Client side of the vukem catalogues (https://vukem.dev/docs/frontend/inertia-i18n):
// fills {name} placeholders and picks plural forms like the server's i18n.T.
export type Message = string | Record<string, string>;

export interface I18nProps {
  locale: string;
  locales: string[];
  messages: Record<string, Message>;
  [key: string]: unknown;
}

const fill = (text: string, params: Record<string, string | number>): string =>
  text.replace(/\{(\w+)\}/g, (match: string, name: string) =>
    name in params ? String(params[name]) : match,
  );

export function translate(
  messages: Record<string, Message>,
  locale: string,
  key: string,
  params: Record<string, string | number> = {},
): string {
  const message = messages[key];
  if (message === undefined) return key;
  if (typeof message === "string") return fill(message, params);

  const count = params.count;
  if (typeof count !== "number") return fill(message.other ?? key, params);
  // An explicit zero form wins for 0 in every language, as on the server.
  if (count === 0 && message.zero !== undefined) return fill(message.zero, params);
  const category = new Intl.PluralRules(locale).select(count);
  return fill(message[category] ?? message.other ?? key, params);
}
