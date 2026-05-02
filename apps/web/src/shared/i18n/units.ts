import { normalizeUOMCode } from "../format/numberFormat";
import { translate, type Locale } from "./index";

export function getUnitLabel(uomCode: string, locale?: Locale) {
  const code = normalizeUOMCode(uomCode);
  return translate(`units.${code}`, { locale, fallback: code });
}
