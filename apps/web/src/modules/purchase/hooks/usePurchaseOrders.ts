"use client";

import { useEffect, useState } from "react";
import { getPurchaseOrders } from "../services/purchaseOrderService";
import type { PurchaseOrder, PurchaseOrderQuery } from "../types";

export function usePurchaseOrders(query: PurchaseOrderQuery = {}) {
  const [orders, setOrders] = useState<PurchaseOrder[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    let active = true;
    setLoading(true);
    setError(null);

    getPurchaseOrders(query)
      .then((items) => {
        if (active) {
          setOrders(items);
        }
      })
      .catch((reason) => {
        if (active) {
          setError(reason instanceof Error ? reason : new Error("Purchase orders could not be loaded"));
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
  }, [query.expectedFrom, query.expectedTo, query.search, query.status, query.supplierId, query.warehouseId]);

  return { orders, loading, error };
}
