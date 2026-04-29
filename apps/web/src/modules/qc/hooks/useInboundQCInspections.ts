"use client";

import { useEffect, useState } from "react";
import { getInboundQCInspections } from "../services/inboundQCService";
import type { InboundQCInspection, InboundQCInspectionQuery } from "../types";

export function useInboundQCInspections(query: InboundQCInspectionQuery = {}) {
  const [inspections, setInspections] = useState<InboundQCInspection[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    let active = true;
    setLoading(true);
    setError(null);

    getInboundQCInspections(query)
      .then((items) => {
        if (active) {
          setInspections(items);
        }
      })
      .catch((reason) => {
        if (active) {
          setError(reason instanceof Error ? reason : new Error("Inbound QC inspections could not be loaded"));
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
  }, [query.goodsReceiptId, query.goodsReceiptLineId, query.status, query.warehouseId]);

  return { inspections, loading, error };
}
