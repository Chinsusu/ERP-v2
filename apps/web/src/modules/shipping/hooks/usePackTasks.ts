"use client";

import { useEffect, useState } from "react";
import { getPackTasks } from "../services/packTaskService";
import type { PackTask, PackTaskQuery } from "../types";

export function usePackTasks(query: PackTaskQuery = {}) {
  const [packTasks, setPackTasks] = useState<PackTask[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    let active = true;
    setLoading(true);
    setError(null);

    getPackTasks(query)
      .then((rows) => {
        if (active) {
          setPackTasks(rows);
        }
      })
      .catch((cause: unknown) => {
        if (active) {
          setError(cause instanceof Error ? cause : new Error("Pack tasks could not be loaded"));
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

  return { packTasks, loading, error };
}
