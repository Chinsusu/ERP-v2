import { defaultLocale, isSupportedLocale, type Locale } from "./config";

export const localeStorageKey = "erp-ui-locale";

let activeLocale: Locale = defaultLocale;
const listeners = new Set<() => void>();

export function getActiveLocale() {
  return activeLocale;
}

export function setActiveLocale(locale: Locale) {
  activeLocale = locale;
  syncDocumentLocale(locale);
  persistLocale(locale);
  notifyLocaleListeners();
}

export function hydrateActiveLocale() {
  if (typeof window === "undefined") {
    return activeLocale;
  }

  const storedLocale = window.localStorage.getItem(localeStorageKey);
  const nextLocale = storedLocale && isSupportedLocale(storedLocale) ? storedLocale : defaultLocale;
  setActiveLocale(nextLocale);

  return nextLocale;
}

export function subscribeActiveLocale(listener: () => void) {
  listeners.add(listener);

  return () => {
    listeners.delete(listener);
  };
}

function persistLocale(locale: Locale) {
  if (typeof window !== "undefined") {
    window.localStorage.setItem(localeStorageKey, locale);
  }
}

function syncDocumentLocale(locale: Locale) {
  if (typeof document !== "undefined") {
    document.documentElement.lang = locale;
  }
}

function notifyLocaleListeners() {
  listeners.forEach((listener) => listener());
}
