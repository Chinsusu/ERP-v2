"use client";

import { useEffect, useState } from "react";
import { getCustomerReceivables } from "../services/customerReceivableService";
import type { CustomerReceivable, CustomerReceivableQuery } from "../types";

export function useCustomerReceivables(query: CustomerReceivableQuery = {}) {
  const [receivables, setReceivables] = useState<CustomerReceivable[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    let active = true;
    setLoading(true);
    setError(null);

    getCustomerReceivables(query)
      .then((items) => {
        if (active) {
          setReceivables(items);
        }
      })
      .catch((reason) => {
        if (active) {
          setError(reason instanceof Error ? reason : new Error("Customer receivables could not be loaded"));
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
  }, [query.customerId, query.search, query.status]);

  return { receivables, loading, error };
}
