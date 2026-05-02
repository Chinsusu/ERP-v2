import { translate, type Locale } from "./index";

export function getValidationLabel(rule: string, locale?: Locale) {
  return translate(`validation.${rule}`, { locale, fallback: rule });
}
