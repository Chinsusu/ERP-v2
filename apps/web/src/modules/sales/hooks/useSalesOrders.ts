"use client";

import { useEffect, useState } from "react";
import { getSalesOrders } from "../services/salesOrderService";
import type { SalesOrder, SalesOrderQuery } from "../types";

export function useSalesOrders(query: SalesOrderQuery = {}) {
  const [orders, setOrders] = useState<SalesOrder[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    let active = true;
    setLoading(true);
    setError(null);

    getSalesOrders(query)
      .then((items) => {
        if (active) {
          setOrders(items);
        }
      })
      .catch((reason) => {
        if (active) {
          setError(reason instanceof Error ? reason : new Error("Sales orders could not be loaded"));
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
  }, [query.channel, query.customerId, query.search, query.status, query.warehouseId]);

  return { orders, loading, error };
}
