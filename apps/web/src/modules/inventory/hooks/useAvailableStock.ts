import { useEffect, useMemo, useState } from "react";
import { getAvailableStock, summarizeAvailableStock } from "../services/stockAvailabilityService";
import type { AvailableStockItem, AvailableStockQuery } from "../types";

export function useAvailableStock(query: AvailableStockQuery = {}) {
  const [items, setItems] = useState<AvailableStockItem[]>([]);
  const [loading, setLoading] = useState(true);
  const queryKey = JSON.stringify(query);

  useEffect(() => {
    let active = true;
    const currentQuery = query;

    setLoading(true);
    getAvailableStock(currentQuery)
      .then((data) => {
        if (active) {
          setItems(data);
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

  const summary = useMemo(() => summarizeAvailableStock(items), [items]);

  return { items, loading, summary };
}
