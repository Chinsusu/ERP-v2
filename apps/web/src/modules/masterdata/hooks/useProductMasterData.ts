import { useCallback, useEffect, useMemo, useState } from "react";
import {
  changeProductStatus,
  createProduct,
  getProduct,
  getProducts,
  summarizeProducts,
  updateProduct
} from "../services/productMasterDataService";
import type { ProductMasterDataInput, ProductMasterDataItem, ProductMasterDataQuery, ProductStatus } from "../types";

export function useProductMasterData(query: ProductMasterDataQuery = {}) {
  const [items, setItems] = useState<ProductMasterDataItem[]>([]);
  const [selectedItem, setSelectedItem] = useState<ProductMasterDataItem | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | undefined>();
  const queryKey = JSON.stringify(query);

  const load = useCallback(async () => {
    setLoading(true);
    setError(undefined);
    try {
      const data = await getProducts(query);
      setItems(data);
    } catch (loadError) {
      setError(errorMessage(loadError));
    } finally {
      setLoading(false);
    }
  }, [queryKey]);

  useEffect(() => {
    let active = true;

    setLoading(true);
    setError(undefined);
    getProducts(query)
      .then((data) => {
        if (active) {
          setItems(data);
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
  }, [queryKey]);

  const saveNewProduct = useCallback(
    async (input: ProductMasterDataInput) => {
      setSaving(true);
      setError(undefined);
      try {
        const item = await createProduct(input);
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

  const saveProduct = useCallback(
    async (productId: string, input: ProductMasterDataInput) => {
      setSaving(true);
      setError(undefined);
      try {
        const item = await updateProduct(productId, input);
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

  const saveProductStatus = useCallback(
    async (productId: string, status: ProductStatus) => {
      setSaving(true);
      setError(undefined);
      try {
        const item = await changeProductStatus(productId, status);
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

  const loadProductDetail = useCallback(async (productId: string) => {
    setError(undefined);
    try {
      const item = await getProduct(productId);
      setSelectedItem(item);
      return item;
    } catch (detailError) {
      const message = errorMessage(detailError);
      setError(message);
      throw new Error(message);
    }
  }, []);

  const summary = useMemo(() => summarizeProducts(items), [items]);

  return {
    items,
    selectedItem,
    loading,
    saving,
    error,
    summary,
    clearError: () => setError(undefined),
    clearSelectedItem: () => setSelectedItem(null),
    loadProductDetail,
    refresh: load,
    saveNewProduct,
    saveProduct,
    saveProductStatus
  };
}

function errorMessage(error: unknown) {
  return error instanceof Error ? error.message : "Product master data request failed";
}
