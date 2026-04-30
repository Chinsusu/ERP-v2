"use client";

import { useEffect, useState } from "react";
import { getFinanceSummaryReport } from "../services/financeSummaryReportService";
import type { FinanceSummaryQuery, FinanceSummaryReport } from "../types";

export function useFinanceSummaryReport(query: FinanceSummaryQuery = {}) {
  const [report, setReport] = useState<FinanceSummaryReport | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);
  const queryKey = JSON.stringify(query);

  useEffect(() => {
    let active = true;
    setLoading(true);
    setError(null);

    getFinanceSummaryReport(query)
      .then((nextReport) => {
        if (active) {
          setReport(nextReport);
        }
      })
      .catch((reason) => {
        if (active) {
          setError(reason instanceof Error ? reason : new Error("Finance summary report could not be loaded"));
        }
      })
      .finally(() => {
        if (active) {
          setLoading(false);
        }
      });

    return () => {
      active = false;
    };
  }, [queryKey]);

  return { report, loading, error };
}
