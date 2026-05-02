import { translate, type Locale } from "./index";

export function getErrorLabel(errorCode: string, locale?: Locale) {
  return translate(`errors.${errorCode}`, { locale, fallback: errorCode });
}
