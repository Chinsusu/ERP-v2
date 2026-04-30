"use client";

import { useEffect, useState } from "react";
import { getCODRemittances } from "../services/codRemittanceService";
import type { CODRemittance, CODRemittanceQuery } from "../types";

export function useCODRemittances(query: CODRemittanceQuery = {}) {
  const [remittances, setRemittances] = useState<CODRemittance[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    let active = true;
    setLoading(true);
    setError(null);

    getCODRemittances(query)
      .then((items) => {
        if (active) {
          setRemittances(items);
        }
      })
      .catch((reason) => {
        if (active) {
          setError(reason instanceof Error ? reason : new Error("COD remittances could not be loaded"));
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
  }, [query.carrierId, query.search, query.status]);

  return { remittances, loading, error };
}
