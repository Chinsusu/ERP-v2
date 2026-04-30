"use client";

import { useEffect, useState } from "react";
import { getInventorySnapshotReport } from "../services/inventorySnapshotReportService";
import type { InventorySnapshotQuery, InventorySnapshotReport } from "../types";

export function useInventorySnapshotReport(query: InventorySnapshotQuery = {}) {
  const [report, setReport] = useState<InventorySnapshotReport | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);
  const queryKey = JSON.stringify(query);

  useEffect(() => {
    let active = true;
    setLoading(true);
    setError(null);

    getInventorySnapshotReport(query)
      .then((nextReport) => {
        if (active) {
          setReport(nextReport);
        }
      })
      .catch((reason) => {
        if (active) {
          setError(reason instanceof Error ? reason : new Error("Inventory snapshot report could not be loaded"));
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
