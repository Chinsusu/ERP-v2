"use client";

import { useEffect, useState } from "react";
import { getSupplierPayables } from "../services/supplierPayableService";
import type { SupplierPayable, SupplierPayableQuery } from "../types";

export function useSupplierPayables(query: SupplierPayableQuery = {}) {
  const [payables, setPayables] = useState<SupplierPayable[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    let active = true;
    setLoading(true);
    setError(null);

    getSupplierPayables(query)
      .then((items) => {
        if (active) {
          setPayables(items);
        }
      })
      .catch((reason) => {
        if (active) {
          setError(reason instanceof Error ? reason : new Error("Supplier payables could not be loaded"));
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
  }, [query.search, query.status, query.supplierId]);

  return { payables, loading, error };
}
