"use client";

import { useCallback } from "react";
import { usePathname, useRouter, useSearchParams } from "next/navigation";

type ReportTab = "inventory" | "operations" | "finance";
type SearchParamsLike = {
  get(key: string): string | null;
};

export function useReportUrlState() {
  const pathname = usePathname();
  const router = useRouter();
  const searchParams = useSearchParams();
  const searchKey = searchParams.toString();

  const replaceReportUrlParams = useCallback(
    (report: ReportTab, values: Record<string, string | undefined>) => {
      const params = new URLSearchParams(searchKey);
      params.set("report", report);
      for (const [key, value] of Object.entries(values)) {
        const normalized = value?.trim();
        if (normalized) {
          params.set(key, normalized);
        } else {
          params.delete(key);
        }
      }

      const next = params.toString();
      if (next !== searchKey) {
        router.replace(`${pathname}?${next}`, { scroll: false });
      }
    },
    [pathname, router, searchKey]
  );

  return { searchParams, replaceReportUrlParams };
}

export function urlParam(searchParams: SearchParamsLike, key: string) {
  return searchParams.get(key)?.trim() ?? "";
}

export function urlDateParam(
  searchParams: SearchParamsLike,
  key: string,
  fallback: string
) {
  const value = urlParam(searchParams, key);
  return isDateInputValue(value) ? value : fallback;
}

export function urlOptionParam<T extends string>(
  searchParams: SearchParamsLike,
  key: string,
  allowedValues: readonly T[],
  fallback: T
) {
  const value = urlParam(searchParams, key) as T;
  return allowedValues.includes(value) ? value : fallback;
}

function isDateInputValue(value: string) {
  return /^\d{4}-\d{2}-\d{2}$/.test(value);
}
