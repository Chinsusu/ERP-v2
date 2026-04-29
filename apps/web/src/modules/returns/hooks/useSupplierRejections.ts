"use client";

import { useEffect, useState } from "react";
import { getSupplierRejections } from "../services/supplierRejectionService";
import type { SupplierRejection, SupplierRejectionQuery } from "../types";

export function useSupplierRejections(query: SupplierRejectionQuery = {}) {
  const [rejections, setRejections] = useState<SupplierRejection[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    let active = true;
    setLoading(true);
    setError(null);

    getSupplierRejections(query)
      .then((items) => {
        if (active) {
          setRejections(items);
        }
      })
      .catch((reason) => {
        if (active) {
          setError(reason instanceof Error ? reason : new Error("Supplier rejections could not be loaded"));
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
  }, [query.supplierId, query.status, query.warehouseId]);

  return { rejections, loading, error };
}
