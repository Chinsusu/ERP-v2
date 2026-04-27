import { useCallback, useEffect, useMemo, useState } from "react";
import {
  changeLocationStatus,
  changeWarehouseStatus,
  createLocation,
  createWarehouse,
  getLocation,
  getLocations,
  getWarehouse,
  getWarehouses,
  summarizeWarehouseLocations,
  updateLocation,
  updateWarehouse
} from "../services/warehouseMasterDataService";
import type {
  WarehouseLocationMasterDataInput,
  WarehouseLocationMasterDataItem,
  WarehouseLocationMasterDataQuery,
  WarehouseLocationStatus,
  WarehouseMasterDataInput,
  WarehouseMasterDataItem,
  WarehouseMasterDataQuery,
  WarehouseStatus
} from "../types";

export function useWarehouseMasterData(
  warehouseQuery: WarehouseMasterDataQuery = {},
  locationQuery: WarehouseLocationMasterDataQuery = {}
) {
  const [warehouses, setWarehouses] = useState<WarehouseMasterDataItem[]>([]);
  const [locations, setLocations] = useState<WarehouseLocationMasterDataItem[]>([]);
  const [selectedWarehouse, setSelectedWarehouse] = useState<WarehouseMasterDataItem | null>(null);
  const [selectedLocation, setSelectedLocation] = useState<WarehouseLocationMasterDataItem | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | undefined>();
  const warehouseQueryKey = JSON.stringify(warehouseQuery);
  const locationQueryKey = JSON.stringify(locationQuery);

  const load = useCallback(async () => {
    setLoading(true);
    setError(undefined);
    try {
      const [warehouseRows, locationRows] = await Promise.all([getWarehouses(warehouseQuery), getLocations(locationQuery)]);
      setWarehouses(warehouseRows);
      setLocations(locationRows);
    } catch (loadError) {
      setError(errorMessage(loadError));
    } finally {
      setLoading(false);
    }
  }, [warehouseQueryKey, locationQueryKey]);

  useEffect(() => {
    let active = true;

    setLoading(true);
    setError(undefined);
    Promise.all([getWarehouses(warehouseQuery), getLocations(locationQuery)])
      .then(([warehouseRows, locationRows]) => {
        if (active) {
          setWarehouses(warehouseRows);
          setLocations(locationRows);
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
  }, [warehouseQueryKey, locationQueryKey]);

  const saveNewWarehouse = useCallback(
    async (input: WarehouseMasterDataInput) => {
      setSaving(true);
      setError(undefined);
      try {
        const item = await createWarehouse(input);
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

  const saveWarehouse = useCallback(
    async (warehouseId: string, input: WarehouseMasterDataInput) => {
      setSaving(true);
      setError(undefined);
      try {
        const item = await updateWarehouse(warehouseId, input);
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

  const saveWarehouseStatus = useCallback(
    async (warehouseId: string, status: WarehouseStatus) => {
      setSaving(true);
      setError(undefined);
      try {
        const item = await changeWarehouseStatus(warehouseId, status);
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

  const saveNewLocation = useCallback(
    async (input: WarehouseLocationMasterDataInput) => {
      setSaving(true);
      setError(undefined);
      try {
        const item = await createLocation(input);
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

  const saveLocation = useCallback(
    async (locationId: string, input: WarehouseLocationMasterDataInput) => {
      setSaving(true);
      setError(undefined);
      try {
        const item = await updateLocation(locationId, input);
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

  const saveLocationStatus = useCallback(
    async (locationId: string, status: WarehouseLocationStatus) => {
      setSaving(true);
      setError(undefined);
      try {
        const item = await changeLocationStatus(locationId, status);
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

  const loadWarehouseDetail = useCallback(async (warehouseId: string) => {
    setError(undefined);
    try {
      const item = await getWarehouse(warehouseId);
      setSelectedWarehouse(item);
      setSelectedLocation(null);
      return item;
    } catch (detailError) {
      const message = errorMessage(detailError);
      setError(message);
      throw new Error(message);
    }
  }, []);

  const loadLocationDetail = useCallback(async (locationId: string) => {
    setError(undefined);
    try {
      const item = await getLocation(locationId);
      setSelectedLocation(item);
      setSelectedWarehouse(null);
      return item;
    } catch (detailError) {
      const message = errorMessage(detailError);
      setError(message);
      throw new Error(message);
    }
  }, []);

  const summary = useMemo(() => summarizeWarehouseLocations(warehouses, locations), [warehouses, locations]);

  return {
    warehouses,
    locations,
    selectedWarehouse,
    selectedLocation,
    loading,
    saving,
    error,
    summary,
    clearError: () => setError(undefined),
    clearSelectedWarehouse: () => setSelectedWarehouse(null),
    clearSelectedLocation: () => setSelectedLocation(null),
    loadWarehouseDetail,
    loadLocationDetail,
    refresh: load,
    saveNewWarehouse,
    saveWarehouse,
    saveWarehouseStatus,
    saveNewLocation,
    saveLocation,
    saveLocationStatus
  };
}

function errorMessage(error: unknown) {
  return error instanceof Error ? error.message : "Warehouse master data request failed";
}
