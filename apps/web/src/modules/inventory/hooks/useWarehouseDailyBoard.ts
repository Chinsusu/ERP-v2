import { useEffect, useState } from "react";
import { getWarehouseDailyBoard } from "../services/warehouseDailyBoardService";
import type { WarehouseDailyBoardItem } from "../types";

export function useWarehouseDailyBoard() {
  const [items, setItems] = useState<WarehouseDailyBoardItem[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let active = true;

    getWarehouseDailyBoard()
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
  }, []);

  return { items, loading };
}
