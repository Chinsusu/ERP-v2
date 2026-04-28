"use client";

import { useEffect, useState } from "react";
import { getReturnReceipts } from "../services/returnReceivingService";
import type { ReturnReceipt, ReturnReceiptQuery } from "../types";

export function useReturnReceipts(query: ReturnReceiptQuery = {}) {
  const [receipts, setReceipts] = useState<ReturnReceipt[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    let active = true;
    setLoading(true);
    setError(null);

    getReturnReceipts(query)
      .then((rows) => {
        if (active) {
          setReceipts(rows);
        }
      })
      .catch((cause: unknown) => {
        if (active) {
          setError(cause instanceof Error ? cause : new Error("Return receipts could not be loaded"));
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

  return { receipts, loading, error };
}
