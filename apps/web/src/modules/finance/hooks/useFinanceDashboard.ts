"use client";

import { useEffect, useState } from "react";
import { getFinanceDashboard } from "../services/financeDashboardService";
import type { FinanceDashboard, FinanceDashboardQuery } from "../types";

export function useFinanceDashboard(query: FinanceDashboardQuery = {}) {
  const [dashboard, setDashboard] = useState<FinanceDashboard | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    let active = true;
    setLoading(true);
    setError(null);

    getFinanceDashboard(query)
      .then((metrics) => {
        if (active) {
          setDashboard(metrics);
        }
      })
      .catch((reason) => {
        if (active) {
          setError(reason instanceof Error ? reason : new Error("Finance dashboard could not be loaded"));
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
  }, [query.businessDate]);

  return { dashboard, loading, error };
}
