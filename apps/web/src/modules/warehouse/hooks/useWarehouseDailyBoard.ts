import { useEffect, useState } from "react";
import { getWarehouseDailyBoard } from "../services/warehouseDailyBoardService";
import type { WarehouseDailyBoardData, WarehouseDailyBoardQuery } from "../types";

export function useWarehouseDailyBoard(query: WarehouseDailyBoardQuery = {}) {
  const [board, setBoard] = useState<WarehouseDailyBoardData | null>(null);
  const [loading, setLoading] = useState(true);
  const queryKey = JSON.stringify(query);

  useEffect(() => {
    let active = true;
    const currentQuery = query;

    setLoading(true);
    getWarehouseDailyBoard(currentQuery)
      .then((data) => {
        if (active) {
          setBoard(data);
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

  return { board, loading };
}
