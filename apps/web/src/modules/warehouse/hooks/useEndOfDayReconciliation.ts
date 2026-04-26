import { useEffect, useState } from "react";
import { getEndOfDayReconciliations } from "../services/warehouseDailyBoardService";
import type { EndOfDayReconciliation, EndOfDayReconciliationQuery } from "../types";

export function useEndOfDayReconciliation(query: EndOfDayReconciliationQuery = {}) {
  const [reconciliations, setReconciliations] = useState<EndOfDayReconciliation[]>([]);
  const [loading, setLoading] = useState(true);
  const queryKey = JSON.stringify(query);

  useEffect(() => {
    let active = true;
    const currentQuery = query;

    setLoading(true);
    getEndOfDayReconciliations(currentQuery)
      .then((data) => {
        if (active) {
          setReconciliations(data);
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

  return { reconciliations, loading };
}
