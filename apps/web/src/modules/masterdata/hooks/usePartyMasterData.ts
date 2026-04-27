import { useCallback, useEffect, useMemo, useState } from "react";
import {
  changeCustomerStatus,
  changeSupplierStatus,
  createCustomer,
  createSupplier,
  getCustomer,
  getCustomers,
  getSupplier,
  getSuppliers,
  summarizeParties,
  updateCustomer,
  updateSupplier
} from "../services/partyMasterDataService";
import type {
  CustomerMasterDataInput,
  CustomerMasterDataItem,
  CustomerMasterDataQuery,
  CustomerStatus,
  SupplierMasterDataInput,
  SupplierMasterDataItem,
  SupplierMasterDataQuery,
  SupplierStatus
} from "../types";

export function usePartyMasterData(
  supplierQuery: SupplierMasterDataQuery = {},
  customerQuery: CustomerMasterDataQuery = {}
) {
  const [suppliers, setSuppliers] = useState<SupplierMasterDataItem[]>([]);
  const [customers, setCustomers] = useState<CustomerMasterDataItem[]>([]);
  const [selectedSupplier, setSelectedSupplier] = useState<SupplierMasterDataItem | null>(null);
  const [selectedCustomer, setSelectedCustomer] = useState<CustomerMasterDataItem | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | undefined>();
  const supplierQueryKey = JSON.stringify(supplierQuery);
  const customerQueryKey = JSON.stringify(customerQuery);

  const load = useCallback(async () => {
    setLoading(true);
    setError(undefined);
    try {
      const [supplierRows, customerRows] = await Promise.all([getSuppliers(supplierQuery), getCustomers(customerQuery)]);
      setSuppliers(supplierRows);
      setCustomers(customerRows);
    } catch (loadError) {
      setError(errorMessage(loadError));
    } finally {
      setLoading(false);
    }
  }, [supplierQueryKey, customerQueryKey]);

  useEffect(() => {
    let active = true;

    setLoading(true);
    setError(undefined);
    Promise.all([getSuppliers(supplierQuery), getCustomers(customerQuery)])
      .then(([supplierRows, customerRows]) => {
        if (active) {
          setSuppliers(supplierRows);
          setCustomers(customerRows);
        }
      })
      .catch((loadError) => {
        if (active) {
          setError(errorMessage(loadError));
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
  }, [supplierQueryKey, customerQueryKey]);

  const saveNewSupplier = useCallback(
    async (input: SupplierMasterDataInput) => {
      setSaving(true);
      setError(undefined);
      try {
        const item = await createSupplier(input);
        await load();
        return item;
      } catch (saveError) {
        const message = errorMessage(saveError);
        setError(message);
        throw new Error(message);
      } finally {
        setSaving(false);
      }
    },
    [load]
  );

  const saveSupplier = useCallback(
    async (supplierId: string, input: SupplierMasterDataInput) => {
      setSaving(true);
      setError(undefined);
      try {
        const item = await updateSupplier(supplierId, input);
        await load();
        return item;
      } catch (saveError) {
        const message = errorMessage(saveError);
        setError(message);
        throw new Error(message);
      } finally {
        setSaving(false);
      }
    },
    [load]
  );

  const saveSupplierStatus = useCallback(
    async (supplierId: string, status: SupplierStatus) => {
      setSaving(true);
      setError(undefined);
      try {
        const item = await changeSupplierStatus(supplierId, status);
        await load();
        return item;
      } catch (saveError) {
        const message = errorMessage(saveError);
        setError(message);
        throw new Error(message);
      } finally {
        setSaving(false);
      }
    },
    [load]
  );

  const saveNewCustomer = useCallback(
    async (input: CustomerMasterDataInput) => {
      setSaving(true);
      setError(undefined);
      try {
        const item = await createCustomer(input);
        await load();
        return item;
      } catch (saveError) {
        const message = errorMessage(saveError);
        setError(message);
        throw new Error(message);
      } finally {
        setSaving(false);
      }
    },
    [load]
  );

  const saveCustomer = useCallback(
    async (customerId: string, input: CustomerMasterDataInput) => {
      setSaving(true);
      setError(undefined);
      try {
        const item = await updateCustomer(customerId, input);
        await load();
        return item;
      } catch (saveError) {
        const message = errorMessage(saveError);
        setError(message);
        throw new Error(message);
      } finally {
        setSaving(false);
      }
    },
    [load]
  );

  const saveCustomerStatus = useCallback(
    async (customerId: string, status: CustomerStatus) => {
      setSaving(true);
      setError(undefined);
      try {
        const item = await changeCustomerStatus(customerId, status);
        await load();
        return item;
      } catch (saveError) {
        const message = errorMessage(saveError);
        setError(message);
        throw new Error(message);
      } finally {
        setSaving(false);
      }
    },
    [load]
  );

  const loadSupplierDetail = useCallback(async (supplierId: string) => {
    setError(undefined);
    try {
      const item = await getSupplier(supplierId);
      setSelectedSupplier(item);
      setSelectedCustomer(null);
      return item;
    } catch (detailError) {
      const message = errorMessage(detailError);
      setError(message);
      throw new Error(message);
    }
  }, []);

  const loadCustomerDetail = useCallback(async (customerId: string) => {
    setError(undefined);
    try {
      const item = await getCustomer(customerId);
      setSelectedCustomer(item);
      setSelectedSupplier(null);
      return item;
    } catch (detailError) {
      const message = errorMessage(detailError);
      setError(message);
      throw new Error(message);
    }
  }, []);

  const summary = useMemo(() => summarizeParties(suppliers, customers), [suppliers, customers]);

  return {
    suppliers,
    customers,
    selectedSupplier,
    selectedCustomer,
    loading,
    saving,
    error,
    summary,
    clearError: () => setError(undefined),
    clearSelectedSupplier: () => setSelectedSupplier(null),
    clearSelectedCustomer: () => setSelectedCustomer(null),
    loadSupplierDetail,
    loadCustomerDetail,
    refresh: load,
    saveNewSupplier,
    saveSupplier,
    saveSupplierStatus,
    saveNewCustomer,
    saveCustomer,
    saveCustomerStatus
  };
}

function errorMessage(error: unknown) {
  return error instanceof Error ? error.message : "Party master data request failed";
}
