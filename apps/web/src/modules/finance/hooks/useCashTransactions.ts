"use client";

import { useEffect, useState } from "react";
import { getCashTransactions } from "../services/cashTransactionService";
import type { CashTransaction, CashTransactionQuery } from "../types";

export function useCashTransactions(query: CashTransactionQuery = {}) {
  const [transactions, setTransactions] = useState<CashTransaction[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    let active = true;
    setLoading(true);
    setError(null);

    getCashTransactions(query)
      .then((items) => {
        if (active) {
          setTransactions(items);
        }
      })
      .catch((reason) => {
        if (active) {
          setError(reason instanceof Error ? reason : new Error("Cash transactions could not be loaded"));
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
  }, [query.counterpartyId, query.direction, query.search, query.status]);

  return { transactions, loading, error };
}
