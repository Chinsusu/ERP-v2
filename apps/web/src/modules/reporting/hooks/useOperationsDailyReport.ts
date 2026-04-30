"use client";

import { useEffect, useState } from "react";
import { getOperationsDailyReport } from "../services/operationsDailyReportService";
import type { OperationsDailyQuery, OperationsDailyReport } from "../types";

export function useOperationsDailyReport(query: OperationsDailyQuery = {}) {
  const [report, setReport] = useState<OperationsDailyReport | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);
  const queryKey = JSON.stringify(query);

  useEffect(() => {
    let active = true;
    setLoading(true);
    setError(null);

    getOperationsDailyReport(query)
      .then((nextReport) => {
        if (active) {
          setReport(nextReport);
        }
      })
      .catch((reason) => {
        if (active) {
          setError(reason instanceof Error ? reason : new Error("Operations daily report could not be loaded"));
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
