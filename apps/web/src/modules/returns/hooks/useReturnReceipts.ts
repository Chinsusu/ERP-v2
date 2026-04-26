"use client";

import { useEffect, useState } from "react";
import { getReturnReceipts } from "../services/returnReceivingService";
import type { ReturnReceipt, ReturnReceiptQuery } from "../types";

export function useReturnReceipts(query: ReturnReceiptQuery = {}) {
  const [receipts, setReceipts] = useState<ReturnReceipt[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let active = true;
    setLoading(true);

    getReturnReceipts(query)
      .then((rows) => {
        if (active) {
          setReceipts(rows);
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
  }, [query]);

  return { receipts, loading };
}
