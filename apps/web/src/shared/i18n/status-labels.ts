import { translate, type Locale } from "./index";

export function getStatusLabel(status: string, locale?: Locale) {
  return translate(`status.${status}`, { locale, fallback: status });
}
