"use client";

import { useEffect, useState } from "react";
import { getPickTasks } from "../services/pickTaskService";
import type { PickTask, PickTaskQuery } from "../types";

export function usePickTasks(query: PickTaskQuery = {}) {
  const [pickTasks, setPickTasks] = useState<PickTask[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    let active = true;
    setLoading(true);
    setError(null);

    getPickTasks(query)
      .then((rows) => {
        if (active) {
          setPickTasks(rows);
        }
      })
      .catch((cause: unknown) => {
        if (active) {
          setError(cause instanceof Error ? cause : new Error("Pick tasks could not be loaded"));
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

  return { pickTasks, loading, error };
}
