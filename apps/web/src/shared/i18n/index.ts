import { defaultLocale, fallbackLocale, isSupportedLocale, type Locale } from "./config";
import { dictionaries, dictionaryNamespaces, type DictionaryNamespace, type DictionaryTree } from "./dictionaries";
import { getActiveLocale } from "./runtime";

export type TranslationValues = Record<string, string | number>;

export type TranslationOptions = {
  locale?: Locale;
  fallback?: string;
  values?: TranslationValues;
};

export function translate(key: string, options: TranslationOptions = {}) {
  const locale = options.locale ?? getActiveLocale();
  const value =
    getTranslationValue(locale, key) ??
    getTranslationValue(fallbackLocale, key) ??
    options.fallback ??
    key;

  return interpolate(String(value), options.values);
}

export const t = translate;

export function resolveLocale(value: string | null | undefined): Locale {
  return value && isSupportedLocale(value) ? value : defaultLocale;
}

export function getDictionary(locale: Locale = defaultLocale) {
  return dictionaries[locale];
}

function getTranslationValue(locale: Locale, key: string) {
  const [namespace, ...path] = key.split(".");
  if (!isDictionaryNamespace(namespace) || path.length === 0) {
    return null;
  }

  return getNestedValue(dictionaries[locale][namespace], path);
}

function isDictionaryNamespace(value: string): value is DictionaryNamespace {
  return dictionaryNamespaces.includes(value as DictionaryNamespace);
}

function getNestedValue(tree: DictionaryTree, path: string[]) {
  let current: unknown = tree;
  for (const segment of path) {
    if (!isRecord(current) || !(segment in current)) {
      return null;
    }
    current = current[segment];
  }

  return typeof current === "string" || typeof current === "number" ? current : null;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function interpolate(value: string, values: TranslationValues | undefined) {
  if (!values) {
    return value;
  }

  return value.replace(/\{([a-zA-Z0-9_]+)\}/g, (match, name: string) =>
    values[name] === undefined ? match : String(values[name])
  );
}

export type { Locale } from "./config";
