export const supportedLocales = ["vi", "en"] as const;

export type Locale = (typeof supportedLocales)[number];

export const defaultLocale: Locale = "vi";
export const fallbackLocale: Locale = "en";
export const erpDisplayLocale = "vi-VN";
export const erpDisplayTimezone = "Asia/Ho_Chi_Minh";
export const erpDisplayCurrency = "VND";

export function isSupportedLocale(value: string): value is Locale {
  return supportedLocales.includes(value as Locale);
}
