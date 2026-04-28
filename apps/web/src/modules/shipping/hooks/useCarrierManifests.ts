"use client";

import { useEffect, useState } from "react";
import { getCarrierManifests } from "../services/carrierManifestService";
import type { CarrierManifest, CarrierManifestQuery } from "../types";

export function useCarrierManifests(query: CarrierManifestQuery = {}) {
  const [manifests, setManifests] = useState<CarrierManifest[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    let active = true;
    setLoading(true);
    setError(null);

    getCarrierManifests(query)
      .then((rows) => {
        if (active) {
          setManifests(rows);
        }
      })
      .catch((cause: unknown) => {
        if (active) {
          setError(cause instanceof Error ? cause : new Error("Carrier manifests could not be loaded"));
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

  return { manifests, loading, error };
}
