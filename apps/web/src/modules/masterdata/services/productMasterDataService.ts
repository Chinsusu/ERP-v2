import { apiGet, apiGetRaw, apiPatch, apiPost } from "../../../shared/api/client";
import { shouldUsePrototypeFallback } from "../../../shared/api/prototypeFallback";
import type { components, operations } from "../../../shared/api/generated/schema";
import { decimalScales, isNegativeDecimal, normalizeDecimalInput } from "../../../shared/format/numberFormat";
import importedProductMasterData from "../data/importedProductMasterData.json";
import type {
  ProductMasterDataInput,
  ProductMasterDataItem,
  ProductMasterDataQuery,
  ProductMasterDataSummary,
  ProductStatus,
  ProductType
} from "../types";

type ProductApiItem = components["schemas"]["ProductListItem"];
type ProductApiQuery = operations["listProducts"]["parameters"]["query"];
type ProductApiCreateRequest = components["schemas"]["CreateProductRequest"];
type ProductApiUpdateRequest = components["schemas"]["UpdateProductRequest"];
type ProductApiStatusRequest = components["schemas"]["ChangeProductStatusRequest"];

const defaultAccessToken = "local-dev-access-token";
const productApiPageSize = 100;
export const finishedProductTypes: ProductType[] = ["finished_good", "semi_finished"];
export const materialProductTypes: ProductType[] = ["raw_material", "packaging", "service"];

export const productTypeOptions: { label: string; value: ProductType }[] = [
  { label: "Finished good", value: "finished_good" },
  { label: "Raw material", value: "raw_material" },
  { label: "Packaging", value: "packaging" },
  { label: "Semi finished", value: "semi_finished" },
  { label: "Service", value: "service" }
];

export const productStatusOptions: { label: string; value: ProductStatus }[] = [
  { label: "Draft", value: "draft" },
  { label: "Active", value: "active" },
  { label: "Inactive", value: "inactive" },
  { label: "Obsolete", value: "obsolete" }
];

export const productUomOptions: { value: string }[] = [
  { value: "MG" },
  { value: "G" },
  { value: "KG" },
  { value: "ML" },
  { value: "L" },
  { value: "PCS" },
  { value: "BOTTLE" },
  { value: "JAR" },
  { value: "TUBE" },
  { value: "BOX" },
  { value: "CARTON" },
  { value: "SET" },
  { value: "BAG" },
  { value: "PACK" },
  { value: "ROLL" },
  { value: "CM" },
  { value: "SERVICE" }
];

export const emptyProductInput: ProductMasterDataInput = {
  itemCode: "",
  skuCode: "",
  name: "",
  itemType: "finished_good",
  itemGroup: "",
  brandCode: "MYH",
  uomBase: "PCS",
  uomPurchase: "PCS",
  uomIssue: "PCS",
  lotControlled: true,
  expiryControlled: true,
  shelfLifeDays: 365,
  qcRequired: true,
  status: "draft",
  standardCost: "0.000000",
  isSellable: true,
  isPurchasable: false,
  isProducible: true,
  specVersion: ""
};

export const prototypeProductMasterData: ProductMasterDataItem[] = importedProductMasterData as ProductMasterDataItem[];

let localProducts = cloneProducts(prototypeProductMasterData);

export async function getProducts(query: ProductMasterDataQuery = {}): Promise<ProductMasterDataItem[]> {
  try {
    const items = await getAllProductApiPages(query);

    return filterProducts(items.map(fromApiItem), query);
  } catch (reason) {
    if (!shouldUsePrototypeFallback(reason)) {
      throw reason;
    }

    return filterProducts(localProducts, query);
  }
}

async function getAllProductApiPages(query: ProductMasterDataQuery): Promise<ProductApiItem[]> {
  const items: ProductApiItem[] = [];
  let page = 1;

  while (true) {
    const pageItems = await apiGet("/products", {
      accessToken: defaultAccessToken,
      query: toApiQuery(query, page)
    });
    items.push(...pageItems);
    if (pageItems.length < productApiPageSize) {
      break;
    }
    page += 1;
  }

  return items;
}

export async function getProduct(productId: string): Promise<ProductMasterDataItem> {
  try {
    const item = await apiGetRaw<ProductApiItem>(`/products/${encodeURIComponent(productId)}`, {
      accessToken: defaultAccessToken
    });

    return fromApiItem(item);
  } catch (error) {
    if (!shouldUsePrototypeFallback(error)) {
      throw error;
    }

    const item = localProducts.find((candidate) => candidate.id === productId);
    if (!item) {
      throw new Error("Product master data was not found");
    }

    return { ...item };
  }
}

export async function createProduct(input: ProductMasterDataInput): Promise<ProductMasterDataItem> {
  const normalized = normalizeInput(input);
  validateProductInput(normalized);

  try {
    const item = await apiPost<ProductApiItem, ProductApiCreateRequest>("/products", toApiRequest(normalized), {
      accessToken: defaultAccessToken
    });

    return fromApiItem(item);
  } catch (error) {
    if (!shouldUsePrototypeFallback(error)) {
      throw error;
    }

    ensureUniqueProduct(normalized);
    const now = new Date().toISOString();
    const item: ProductMasterDataItem = {
      ...normalized,
      id: `item-${normalized.skuCode.toLowerCase().replaceAll("-", "_")}-${Date.now()}`,
      createdAt: now,
      updatedAt: now,
      auditLogId: `audit-local-create-${Date.now()}`
    };
    localProducts = sortProducts([...localProducts, item]);

    return { ...item };
  }
}

export async function updateProduct(productId: string, input: ProductMasterDataInput): Promise<ProductMasterDataItem> {
  const normalized = normalizeInput(input);
  validateProductInput(normalized);

  try {
    const item = await apiPatch<ProductApiItem, ProductApiUpdateRequest>(
      `/products/${encodeURIComponent(productId)}`,
      toApiRequest(normalized),
      {
        accessToken: defaultAccessToken
      }
    );

    return fromApiItem(item);
  } catch (error) {
    if (!shouldUsePrototypeFallback(error)) {
      throw error;
    }

    ensureUniqueProduct(normalized, productId);
    const current = localProducts.find((candidate) => candidate.id === productId);
    if (!current) {
      throw new Error("Product master data was not found");
    }
    const item: ProductMasterDataItem = {
      ...current,
      ...normalized,
      updatedAt: new Date().toISOString(),
      auditLogId: `audit-local-update-${Date.now()}`
    };
    localProducts = sortProducts(localProducts.map((candidate) => (candidate.id === productId ? item : candidate)));

    return { ...item };
  }
}

export async function changeProductStatus(productId: string, status: ProductStatus): Promise<ProductMasterDataItem> {
  try {
    const item = await apiPatch<ProductApiItem, ProductApiStatusRequest>(
      `/products/${encodeURIComponent(productId)}/status`,
      { status },
      {
        accessToken: defaultAccessToken
      }
    );

    return fromApiItem(item);
  } catch (error) {
    if (!shouldUsePrototypeFallback(error)) {
      throw error;
    }

    const current = localProducts.find((candidate) => candidate.id === productId);
    if (!current) {
      throw new Error("Product master data was not found");
    }
    const item: ProductMasterDataItem = {
      ...current,
      status,
      updatedAt: new Date().toISOString(),
      auditLogId: `audit-local-status-${Date.now()}`
    };
    localProducts = sortProducts(localProducts.map((candidate) => (candidate.id === productId ? item : candidate)));

    return { ...item };
  }
}

export function summarizeProducts(items: ProductMasterDataItem[]): ProductMasterDataSummary {
  return {
    total: items.length,
    active: items.filter((item) => item.status === "active").length,
    draft: items.filter((item) => item.status === "draft").length,
    controlled: items.filter((item) => item.lotControlled || item.expiryControlled || item.qcRequired).length
  };
}

export function productStatusTone(status: ProductStatus): "normal" | "success" | "warning" | "danger" | "info" {
  switch (status) {
    case "active":
      return "success";
    case "draft":
      return "info";
    case "inactive":
      return "warning";
    case "obsolete":
      return "danger";
    default:
      return "normal";
  }
}

export function productTypeLabel(type: ProductType) {
  return productTypeOptions.find((option) => option.value === type)?.label ?? type;
}

export function statusLabel(status: ProductStatus) {
  return productStatusOptions.find((option) => option.value === status)?.label ?? status;
}

export function toProductInput(item: ProductMasterDataItem): ProductMasterDataInput {
  return {
    itemCode: item.itemCode,
    skuCode: item.skuCode,
    name: item.name,
    itemType: item.itemType,
    itemGroup: item.itemGroup ?? "",
    brandCode: item.brandCode ?? "",
    uomBase: item.uomBase,
    uomPurchase: item.uomPurchase ?? item.uomBase,
    uomIssue: item.uomIssue ?? item.uomBase,
    lotControlled: item.lotControlled,
    expiryControlled: item.expiryControlled,
    shelfLifeDays: item.shelfLifeDays ?? 0,
    qcRequired: item.qcRequired,
    status: item.status,
    standardCost: item.standardCost ?? "0.000000",
    isSellable: item.isSellable,
    isPurchasable: item.isPurchasable,
    isProducible: item.isProducible,
    specVersion: item.specVersion ?? ""
  };
}

export function resetPrototypeProductMasterData() {
  localProducts = cloneProducts(prototypeProductMasterData);
}

function fromApiItem(item: ProductApiItem): ProductMasterDataItem {
  return {
    id: item.id,
    itemCode: item.item_code,
    skuCode: item.sku_code,
    name: item.name,
    itemType: item.item_type,
    itemGroup: item.item_group,
    brandCode: item.brand_code,
    uomBase: item.uom_base,
    uomPurchase: item.uom_purchase,
    uomIssue: item.uom_issue,
    lotControlled: item.lot_controlled,
    expiryControlled: item.expiry_controlled,
    shelfLifeDays: item.shelf_life_days,
    qcRequired: item.qc_required,
    status: item.status,
    standardCost: item.standard_cost,
    isSellable: item.is_sellable,
    isPurchasable: item.is_purchasable,
    isProducible: item.is_producible,
    specVersion: item.spec_version,
    createdAt: item.created_at,
    updatedAt: item.updated_at,
    auditLogId: item.audit_log_id
  };
}

function toApiQuery(query: ProductMasterDataQuery, page = 1): ProductApiQuery {
  return {
    q: query.search,
    status: query.status || undefined,
    item_type: query.itemType || undefined,
    page,
    page_size: productApiPageSize
  };
}

function toApiRequest(input: ProductMasterDataInput): ProductApiCreateRequest {
  return {
    item_code: input.itemCode,
    sku_code: input.skuCode,
    name: input.name,
    item_type: input.itemType,
    item_group: input.itemGroup || undefined,
    brand_code: input.brandCode || undefined,
    uom_base: input.uomBase,
    uom_purchase: input.uomPurchase || undefined,
    uom_issue: input.uomIssue || undefined,
    lot_controlled: input.lotControlled,
    expiry_controlled: input.expiryControlled,
    shelf_life_days: input.shelfLifeDays,
    qc_required: input.qcRequired,
    status: input.status,
    standard_cost: input.standardCost || undefined,
    is_sellable: input.isSellable,
    is_purchasable: input.isPurchasable,
    is_producible: input.isProducible,
    spec_version: input.specVersion || undefined
  };
}

function normalizeInput(input: ProductMasterDataInput): ProductMasterDataInput {
  return {
    ...input,
    itemCode: input.itemCode.trim().toUpperCase(),
    skuCode: input.skuCode.trim().toUpperCase(),
    name: input.name.trim(),
    itemGroup: input.itemGroup.trim().toLowerCase(),
    brandCode: input.brandCode.trim().toUpperCase(),
    uomBase: input.uomBase.trim().toUpperCase(),
    uomPurchase: input.uomPurchase.trim().toUpperCase(),
    uomIssue: input.uomIssue.trim().toUpperCase(),
    shelfLifeDays: Number.isFinite(input.shelfLifeDays) ? input.shelfLifeDays : 0,
    standardCost: normalizeDecimalInput(input.standardCost, decimalScales.unitCost),
    specVersion: input.specVersion.trim()
  };
}

function validateProductInput(input: ProductMasterDataInput) {
  const missing = [
    ["item code", input.itemCode],
    ["SKU code", input.skuCode],
    ["name", input.name],
    ["base UOM", input.uomBase]
  ].filter(([, value]) => !String(value).trim());

  if (missing.length > 0) {
    throw new Error(`Missing required fields: ${missing.map(([label]) => label).join(", ")}`);
  }
  if (input.expiryControlled && input.shelfLifeDays <= 0) {
    throw new Error("Shelf life days is required when expiry control is enabled");
  }
  if (isNegativeDecimal(input.standardCost)) {
    throw new Error("Standard cost cannot be negative");
  }
}

function ensureUniqueProduct(input: ProductMasterDataInput, currentId?: string) {
  const duplicate = localProducts.find((item) => {
    if (currentId && item.id === currentId) {
      return false;
    }

    return item.itemCode === input.itemCode || item.skuCode === input.skuCode;
  });
  if (!duplicate) {
    return;
  }
  if (duplicate.itemCode === input.itemCode) {
    throw new Error("Item code already exists");
  }

  throw new Error("SKU code already exists");
}

function filterProducts(items: ProductMasterDataItem[], query: ProductMasterDataQuery) {
  const search = query.search?.trim().toLowerCase();
  return sortProducts(
    items.filter((item) => {
      if (query.status && item.status !== query.status) {
        return false;
      }
      if (query.itemType && item.itemType !== query.itemType) {
        return false;
      }
      if (!query.itemType && query.itemTypes && query.itemTypes.length > 0 && !query.itemTypes.includes(item.itemType)) {
        return false;
      }
      if (!search) {
        return true;
      }

      return [item.itemCode, item.skuCode, item.name, item.itemGroup ?? "", item.brandCode ?? ""].some((value) =>
        value.toLowerCase().includes(search)
      );
    })
  );
}

function sortProducts(items: ProductMasterDataItem[]) {
  return [...items].sort((left, right) => {
    const statusDelta = statusRank(left.status) - statusRank(right.status);
    if (statusDelta !== 0) {
      return statusDelta;
    }

    return left.skuCode.localeCompare(right.skuCode);
  });
}

function statusRank(status: ProductStatus) {
  switch (status) {
    case "active":
      return 0;
    case "draft":
      return 1;
    case "inactive":
      return 2;
    case "obsolete":
      return 3;
    default:
      return 4;
  }
}

function cloneProducts(items: ProductMasterDataItem[]) {
  return items.map((item) => ({ ...item }));
}
