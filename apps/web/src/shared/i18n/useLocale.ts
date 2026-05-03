"use client";

import { useEffect, useSyncExternalStore } from "react";
import type { Locale } from "./config";
import { getActiveLocale, hydrateActiveLocale, setActiveLocale, subscribeActiveLocale } from "./runtime";

export function useLocale() {
  const locale = useSyncExternalStore(subscribeActiveLocale, getActiveLocale, getActiveLocale);

  useEffect(() => {
    hydrateActiveLocale();
  }, []);

  return [locale, setActiveLocale] as const satisfies readonly [Locale, (locale: Locale) => void];
}
