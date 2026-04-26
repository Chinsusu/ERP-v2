"use client";

import { useEffect, useState } from "react";
import { getCarrierManifests } from "../services/carrierManifestService";
import type { CarrierManifest, CarrierManifestQuery } from "../types";

export function useCarrierManifests(query: CarrierManifestQuery = {}) {
  const [manifests, setManifests] = useState<CarrierManifest[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let active = true;
    setLoading(true);

    getCarrierManifests(query)
      .then((rows) => {
        if (active) {
          setManifests(rows);
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

  return { manifests, loading };
}
