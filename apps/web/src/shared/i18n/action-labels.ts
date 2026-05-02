import { translate, type Locale } from "./index";

export function getActionLabel(label: string, locale?: Locale) {
  return translate(`actions.${label}`, { locale, fallback: label });
}
