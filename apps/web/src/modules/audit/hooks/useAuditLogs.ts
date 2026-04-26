import { useEffect, useMemo, useState } from "react";
import { getAuditLogs, summarizeAuditLogs } from "../services/auditLogService";
import type { AuditLogItem, AuditLogQuery } from "../types";

export function useAuditLogs(query: AuditLogQuery = {}) {
  const [items, setItems] = useState<AuditLogItem[]>([]);
  const [loading, setLoading] = useState(true);
  const queryKey = JSON.stringify(query);

  useEffect(() => {
    let active = true;
    const currentQuery = query;

    setLoading(true);
    getAuditLogs(currentQuery)
      .then((data) => {
        if (active) {
          setItems(data);
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
  }, [queryKey]);

  const summary = useMemo(() => summarizeAuditLogs(items), [items]);

  return { items, loading, summary };
}
