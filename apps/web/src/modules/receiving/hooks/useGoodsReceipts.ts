"use client";

import { useEffect, useState } from "react";
import { getGoodsReceipts } from "../services/warehouseReceivingService";
import type { GoodsReceipt, GoodsReceiptQuery } from "../types";

export function useGoodsReceipts(query: GoodsReceiptQuery = {}) {
  const [receipts, setReceipts] = useState<GoodsReceipt[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    let active = true;
    setLoading(true);
    setError(null);

    getGoodsReceipts(query)
      .then((items) => {
        if (active) {
          setReceipts(items);
        }
      })
      .catch((reason) => {
        if (active) {
          setError(reason instanceof Error ? reason : new Error("Goods receipts could not be loaded"));
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
  }, [query.status, query.warehouseId]);

  return { receipts, loading, error };
}
