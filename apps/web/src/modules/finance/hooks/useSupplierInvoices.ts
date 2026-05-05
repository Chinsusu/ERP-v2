"use client";

import { useEffect, useState } from "react";
import { getSupplierInvoices } from "../services/supplierInvoiceService";
import type { SupplierInvoice, SupplierInvoiceQuery } from "../types";

export function useSupplierInvoices(query: SupplierInvoiceQuery = {}) {
  const [invoices, setInvoices] = useState<SupplierInvoice[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    let active = true;
    setLoading(true);
    setError(null);

    getSupplierInvoices(query)
      .then((items) => {
        if (active) {
          setInvoices(items);
        }
      })
      .catch((reason) => {
        if (active) {
          setError(reason instanceof Error ? reason : new Error("Supplier invoices could not be loaded"));
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
  }, [query.payableId, query.search, query.status, query.supplierId]);

  return { invoices, loading, error };
}
