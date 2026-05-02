import { apiGet, apiGetRaw, apiPatch, apiPost } from "../../../shared/api/client";
import { shouldUsePrototypeFallback } from "../../../shared/api/prototypeFallback";
import type { components, operations } from "../../../shared/api/generated/schema";
import type {
  WarehouseLocationMasterDataInput,
  WarehouseLocationMasterDataItem,
  WarehouseLocationMasterDataQuery,
  WarehouseLocationMasterDataSummary,
  WarehouseLocationStatus,
  WarehouseLocationType,
  WarehouseMasterDataInput,
  WarehouseMasterDataItem,
  WarehouseMasterDataQuery,
  WarehouseStatus,
  WarehouseType
} from "../types";

type WarehouseApiItem = components["schemas"]["WarehouseListItem"];
type WarehouseLocationApiItem = components["schemas"]["WarehouseLocationListItem"];
type WarehouseApiQuery = operations["listWarehouses"]["parameters"]["query"];
type WarehouseLocationApiQuery = operations["listWarehouseLocations"]["parameters"]["query"];
type WarehouseApiCreateRequest = components["schemas"]["CreateWarehouseRequest"];
type WarehouseApiUpdateRequest = components["schemas"]["UpdateWarehouseRequest"];
type WarehouseApiStatusRequest = components["schemas"]["ChangeWarehouseStatusRequest"];
type WarehouseLocationApiCreateRequest = components["schemas"]["CreateWarehouseLocationRequest"];
type WarehouseLocationApiUpdateRequest = components["schemas"]["UpdateWarehouseLocationRequest"];
type WarehouseLocationApiStatusRequest = components["schemas"]["ChangeWarehouseLocationStatusRequest"];

const defaultAccessToken = "local-dev-access-token";

export const warehouseTypeOptions: { label: string; value: WarehouseType }[] = [
  { label: "Finished good", value: "finished_good" },
  { label: "Raw material", value: "raw_material" },
  { label: "Packaging", value: "packaging" },
  { label: "Semi finished", value: "semi_finished" },
  { label: "Quarantine", value: "quarantine" },
  { label: "Sample", value: "sample" },
  { label: "Defect", value: "defect" },
  { label: "Retail store", value: "retail_store" }
];

export const warehouseStatusOptions: { label: string; value: WarehouseStatus }[] = [
  { label: "Active", value: "active" },
  { label: "Inactive", value: "inactive" }
];

export const locationTypeOptions: { label: string; value: WarehouseLocationType }[] = [
  { label: "Receiving", value: "receiving" },
  { label: "QC hold", value: "qc_hold" },
  { label: "Storage", value: "storage" },
  { label: "Pick", value: "pick" },
  { label: "Pack", value: "pack" },
  { label: "Handover", value: "handover" },
  { label: "Return", value: "return" },
  { label: "Lab", value: "lab" },
  { label: "Scrap", value: "scrap" }
];

export const locationStatusOptions: { label: string; value: WarehouseLocationStatus }[] = [
  { label: "Active", value: "active" },
  { label: "Inactive", value: "inactive" }
];

export const emptyWarehouseInput: WarehouseMasterDataInput = {
  warehouseCode: "",
  warehouseName: "",
  warehouseType: "finished_good",
  siteCode: "HCM",
  address: "",
  allowSaleIssue: true,
  allowProdIssue: false,
  allowQuarantine: false,
  status: "active"
};

export const emptyLocationInput: WarehouseLocationMasterDataInput = {
  warehouseId: "",
  locationCode: "",
  locationName: "",
  locationType: "storage",
  zoneCode: "",
  allowReceive: false,
  allowPick: false,
  allowStore: true,
  isDefault: false,
  status: "active"
};

export const prototypeWarehouses: WarehouseMasterDataItem[] = [
  {
    id: "wh-hcm-fg",
    warehouseCode: "WH-HCM-FG",
    warehouseName: "Finished Goods Warehouse HCM",
    warehouseType: "finished_good",
    siteCode: "HCM",
    address: "Ho Chi Minh distribution center",
    allowSaleIssue: true,
    allowProdIssue: false,
    allowQuarantine: false,
    status: "active",
    createdAt: "2026-04-26T09:00:00Z",
    updatedAt: "2026-04-26T09:00:00Z"
  },
  {
    id: "wh-hcm-rm",
    warehouseCode: "WH-HCM-RM",
    warehouseName: "Raw Material Warehouse HCM",
    warehouseType: "raw_material",
    siteCode: "HCM",
    address: "Ho Chi Minh production site",
    allowSaleIssue: false,
    allowProdIssue: true,
    allowQuarantine: false,
    status: "active",
    createdAt: "2026-04-26T09:10:00Z",
    updatedAt: "2026-04-26T09:10:00Z"
  },
  {
    id: "wh-hcm-qh",
    warehouseCode: "WH-HCM-QH",
    warehouseName: "QC Hold Warehouse HCM",
    warehouseType: "quarantine",
    siteCode: "HCM",
    address: "QC quarantine area",
    allowSaleIssue: false,
    allowProdIssue: false,
    allowQuarantine: true,
    status: "active",
    createdAt: "2026-04-26T09:20:00Z",
    updatedAt: "2026-04-26T09:20:00Z"
  },
  {
    id: "wh-hcm-def",
    warehouseCode: "WH-HCM-DEF",
    warehouseName: "Defect Warehouse HCM",
    warehouseType: "defect",
    siteCode: "HCM",
    address: "Defect and scrap area",
    allowSaleIssue: false,
    allowProdIssue: false,
    allowQuarantine: false,
    status: "inactive",
    createdAt: "2026-04-26T09:30:00Z",
    updatedAt: "2026-04-26T09:30:00Z"
  }
];

export const prototypeLocations: WarehouseLocationMasterDataItem[] = [
  {
    id: "loc-hcm-fg-recv-01",
    warehouseId: "wh-hcm-fg",
    warehouseCode: "WH-HCM-FG",
    locationCode: "FG-RECV-01",
    locationName: "Finished Goods Receiving Dock",
    locationType: "receiving",
    zoneCode: "RECV",
    allowReceive: true,
    allowPick: false,
    allowStore: true,
    isDefault: true,
    status: "active",
    createdAt: "2026-04-26T10:00:00Z",
    updatedAt: "2026-04-26T10:00:00Z"
  },
  {
    id: "loc-hcm-fg-pick-a01",
    warehouseId: "wh-hcm-fg",
    warehouseCode: "WH-HCM-FG",
    locationCode: "FG-PICK-A01",
    locationName: "Finished Goods Pick A01",
    locationType: "pick",
    zoneCode: "PICK",
    allowReceive: false,
    allowPick: true,
    allowStore: true,
    isDefault: false,
    status: "active",
    createdAt: "2026-04-26T10:10:00Z",
    updatedAt: "2026-04-26T10:10:00Z"
  },
  {
    id: "loc-hcm-rm-recv-01",
    warehouseId: "wh-hcm-rm",
    warehouseCode: "WH-HCM-RM",
    locationCode: "RM-RECV-01",
    locationName: "Raw Material Receiving Dock",
    locationType: "receiving",
    zoneCode: "RECV",
    allowReceive: true,
    allowPick: false,
    allowStore: true,
    isDefault: true,
    status: "active",
    createdAt: "2026-04-26T10:20:00Z",
    updatedAt: "2026-04-26T10:20:00Z"
  },
  {
    id: "loc-hcm-qh-hold-01",
    warehouseId: "wh-hcm-qh",
    warehouseCode: "WH-HCM-QH",
    locationCode: "QH-HOLD-01",
    locationName: "QC Hold Bay 01",
    locationType: "qc_hold",
    zoneCode: "QC",
    allowReceive: true,
    allowPick: false,
    allowStore: true,
    isDefault: true,
    status: "active",
    createdAt: "2026-04-26T10:30:00Z",
    updatedAt: "2026-04-26T10:30:00Z"
  },
  {
    id: "loc-hcm-def-scrap-01",
    warehouseId: "wh-hcm-def",
    warehouseCode: "WH-HCM-DEF",
    locationCode: "DEF-SCRAP-01",
    locationName: "Defect Scrap Bay",
    locationType: "scrap",
    zoneCode: "SCRAP",
    allowReceive: true,
    allowPick: false,
    allowStore: true,
    isDefault: false,
    status: "inactive",
    createdAt: "2026-04-26T10:40:00Z",
    updatedAt: "2026-04-26T10:40:00Z"
  }
];

let localWarehouses = cloneWarehouses(prototypeWarehouses);
let localLocations = cloneLocations(prototypeLocations);

export async function getWarehouses(query: WarehouseMasterDataQuery = {}): Promise<WarehouseMasterDataItem[]> {
  try {
    const items = await apiGet("/warehouses", {
      accessToken: defaultAccessToken,
      query: toWarehouseApiQuery(query)
    });

    return items.map(fromWarehouseApiItem);
  } catch (error) {
    if (!shouldUsePrototypeFallback(error)) {
      throw error;
    }

    return filterWarehouses(localWarehouses, query);
  }
}

export async function getWarehouse(warehouseId: string): Promise<WarehouseMasterDataItem> {
  try {
    const item = await apiGetRaw<WarehouseApiItem>(`/warehouses/${encodeURIComponent(warehouseId)}`, {
      accessToken: defaultAccessToken
    });

    return fromWarehouseApiItem(item);
  } catch (error) {
    if (!shouldUsePrototypeFallback(error)) {
      throw error;
    }

    const item = localWarehouses.find((candidate) => candidate.id === warehouseId);
    if (!item) {
      throw new Error("Warehouse master data was not found");
    }

    return { ...item };
  }
}

export async function createWarehouse(input: WarehouseMasterDataInput): Promise<WarehouseMasterDataItem> {
  const normalized = normalizeWarehouseInput(input);
  validateWarehouseInput(normalized);

  try {
    const item = await apiPost<WarehouseApiItem, WarehouseApiCreateRequest>("/warehouses", toWarehouseApiRequest(normalized), {
      accessToken: defaultAccessToken
    });

    return fromWarehouseApiItem(item);
  } catch (error) {
    if (!shouldUsePrototypeFallback(error)) {
      throw error;
    }

    ensureUniqueWarehouse(normalized);
    const now = new Date().toISOString();
    const item: WarehouseMasterDataItem = {
      ...normalized,
      id: `wh-${normalized.warehouseCode.toLowerCase().replaceAll("-", "_")}-${Date.now()}`,
      createdAt: now,
      updatedAt: now,
      auditLogId: `audit-local-warehouse-create-${Date.now()}`
    };
    localWarehouses = sortWarehouses([...localWarehouses, item]);

    return { ...item };
  }
}

export async function updateWarehouse(warehouseId: string, input: WarehouseMasterDataInput): Promise<WarehouseMasterDataItem> {
  const normalized = normalizeWarehouseInput(input);
  validateWarehouseInput(normalized);

  try {
    const item = await apiPatch<WarehouseApiItem, WarehouseApiUpdateRequest>(
      `/warehouses/${encodeURIComponent(warehouseId)}`,
      toWarehouseApiRequest(normalized),
      {
        accessToken: defaultAccessToken
      }
    );

    return fromWarehouseApiItem(item);
  } catch (error) {
    if (!shouldUsePrototypeFallback(error)) {
      throw error;
    }

    ensureUniqueWarehouse(normalized, warehouseId);
    const current = localWarehouses.find((candidate) => candidate.id === warehouseId);
    if (!current) {
      throw new Error("Warehouse master data was not found");
    }
    const item: WarehouseMasterDataItem = {
      ...current,
      ...normalized,
      updatedAt: new Date().toISOString(),
      auditLogId: `audit-local-warehouse-update-${Date.now()}`
    };
    localWarehouses = sortWarehouses(localWarehouses.map((candidate) => (candidate.id === warehouseId ? item : candidate)));
    localLocations = localLocations.map((location) =>
      location.warehouseId === warehouseId ? { ...location, warehouseCode: item.warehouseCode } : location
    );

    return { ...item };
  }
}

export async function changeWarehouseStatus(warehouseId: string, status: WarehouseStatus): Promise<WarehouseMasterDataItem> {
  try {
    const item = await apiPatch<WarehouseApiItem, WarehouseApiStatusRequest>(
      `/warehouses/${encodeURIComponent(warehouseId)}/status`,
      { status },
      {
        accessToken: defaultAccessToken
      }
    );

    return fromWarehouseApiItem(item);
  } catch (error) {
    if (!shouldUsePrototypeFallback(error)) {
      throw error;
    }

    const current = localWarehouses.find((candidate) => candidate.id === warehouseId);
    if (!current) {
      throw new Error("Warehouse master data was not found");
    }
    const item: WarehouseMasterDataItem = {
      ...current,
      status,
      updatedAt: new Date().toISOString(),
      auditLogId: `audit-local-warehouse-status-${Date.now()}`
    };
    localWarehouses = sortWarehouses(localWarehouses.map((candidate) => (candidate.id === warehouseId ? item : candidate)));

    return { ...item };
  }
}

export async function getLocations(query: WarehouseLocationMasterDataQuery = {}): Promise<WarehouseLocationMasterDataItem[]> {
  try {
    const items = await apiGet("/warehouse-locations", {
      accessToken: defaultAccessToken,
      query: toLocationApiQuery(query)
    });

    return items.map(fromLocationApiItem);
  } catch (error) {
    if (!shouldUsePrototypeFallback(error)) {
      throw error;
    }

    return filterLocations(localLocations, query);
  }
}

export async function getLocation(locationId: string): Promise<WarehouseLocationMasterDataItem> {
  try {
    const item = await apiGetRaw<WarehouseLocationApiItem>(`/warehouse-locations/${encodeURIComponent(locationId)}`, {
      accessToken: defaultAccessToken
    });

    return fromLocationApiItem(item);
  } catch (error) {
    if (!shouldUsePrototypeFallback(error)) {
      throw error;
    }

    const item = localLocations.find((candidate) => candidate.id === locationId);
    if (!item) {
      throw new Error("Warehouse location was not found");
    }

    return { ...item };
  }
}

export async function createLocation(input: WarehouseLocationMasterDataInput): Promise<WarehouseLocationMasterDataItem> {
  const normalized = normalizeLocationInput(input);
  validateLocationInput(normalized);

  try {
    const item = await apiPost<WarehouseLocationApiItem, WarehouseLocationApiCreateRequest>(
      "/warehouse-locations",
      toLocationApiRequest(normalized),
      {
        accessToken: defaultAccessToken
      }
    );

    return fromLocationApiItem(item);
  } catch (error) {
    if (!shouldUsePrototypeFallback(error)) {
      throw error;
    }

    const warehouse = findWarehouse(normalized.warehouseId);
    ensureUniqueLocation(normalized);
    const now = new Date().toISOString();
    const item: WarehouseLocationMasterDataItem = {
      ...normalized,
      warehouseCode: warehouse.warehouseCode,
      id: `loc-${warehouse.warehouseCode.toLowerCase().replaceAll("-", "_")}-${normalized.locationCode.toLowerCase().replaceAll("-", "_")}-${Date.now()}`,
      createdAt: now,
      updatedAt: now,
      auditLogId: `audit-local-location-create-${Date.now()}`
    };
    localLocations = sortLocations([...localLocations, item]);

    return { ...item };
  }
}

export async function updateLocation(locationId: string, input: WarehouseLocationMasterDataInput): Promise<WarehouseLocationMasterDataItem> {
  const normalized = normalizeLocationInput(input);
  validateLocationInput(normalized);

  try {
    const item = await apiPatch<WarehouseLocationApiItem, WarehouseLocationApiUpdateRequest>(
      `/warehouse-locations/${encodeURIComponent(locationId)}`,
      toLocationApiRequest(normalized),
      {
        accessToken: defaultAccessToken
      }
    );

    return fromLocationApiItem(item);
  } catch (error) {
    if (!shouldUsePrototypeFallback(error)) {
      throw error;
    }

    const current = localLocations.find((candidate) => candidate.id === locationId);
    if (!current) {
      throw new Error("Warehouse location was not found");
    }
    if (current.status !== "active" && normalized.status !== "active") {
      throw new Error("Inactive warehouse location cannot be edited except by reactivating it");
    }
    const warehouse = findWarehouse(normalized.warehouseId);
    ensureUniqueLocation(normalized, locationId);
    const item: WarehouseLocationMasterDataItem = {
      ...current,
      ...normalized,
      warehouseCode: warehouse.warehouseCode,
      updatedAt: new Date().toISOString(),
      auditLogId: `audit-local-location-update-${Date.now()}`
    };
    localLocations = sortLocations(localLocations.map((candidate) => (candidate.id === locationId ? item : candidate)));

    return { ...item };
  }
}

export async function changeLocationStatus(locationId: string, status: WarehouseLocationStatus): Promise<WarehouseLocationMasterDataItem> {
  try {
    const item = await apiPatch<WarehouseLocationApiItem, WarehouseLocationApiStatusRequest>(
      `/warehouse-locations/${encodeURIComponent(locationId)}/status`,
      { status },
      {
        accessToken: defaultAccessToken
      }
    );

    return fromLocationApiItem(item);
  } catch (error) {
    if (!shouldUsePrototypeFallback(error)) {
      throw error;
    }

    const current = localLocations.find((candidate) => candidate.id === locationId);
    if (!current) {
      throw new Error("Warehouse location was not found");
    }
    const item: WarehouseLocationMasterDataItem = {
      ...current,
      status,
      updatedAt: new Date().toISOString(),
      auditLogId: `audit-local-location-status-${Date.now()}`
    };
    localLocations = sortLocations(localLocations.map((candidate) => (candidate.id === locationId ? item : candidate)));

    return { ...item };
  }
}

export function summarizeWarehouseLocations(
  warehouses: WarehouseMasterDataItem[],
  locations: WarehouseLocationMasterDataItem[]
): WarehouseLocationMasterDataSummary {
  return {
    warehouses: warehouses.length,
    activeWarehouses: warehouses.filter((warehouse) => warehouse.status === "active").length,
    activeLocations: locations.filter((location) => location.status === "active").length,
    receivingLocations: locations.filter((location) => location.status === "active" && location.allowReceive).length
  };
}

export function warehouseStatusTone(status: WarehouseStatus | WarehouseLocationStatus): "normal" | "success" | "warning" | "danger" | "info" {
  return status === "active" ? "success" : "warning";
}

export function warehouseTypeLabel(type: WarehouseType) {
  return warehouseTypeOptions.find((option) => option.value === type)?.label ?? type;
}

export function locationTypeLabel(type: WarehouseLocationType) {
  return locationTypeOptions.find((option) => option.value === type)?.label ?? type;
}

export function warehouseStatusLabel(status: WarehouseStatus) {
  return warehouseStatusOptions.find((option) => option.value === status)?.label ?? status;
}

export function locationStatusLabel(status: WarehouseLocationStatus) {
  return locationStatusOptions.find((option) => option.value === status)?.label ?? status;
}

export function toWarehouseInput(item: WarehouseMasterDataItem): WarehouseMasterDataInput {
  return {
    warehouseCode: item.warehouseCode,
    warehouseName: item.warehouseName,
    warehouseType: item.warehouseType,
    siteCode: item.siteCode,
    address: item.address ?? "",
    allowSaleIssue: item.allowSaleIssue,
    allowProdIssue: item.allowProdIssue,
    allowQuarantine: item.allowQuarantine,
    status: item.status
  };
}

export function toLocationInput(item: WarehouseLocationMasterDataItem): WarehouseLocationMasterDataInput {
  return {
    warehouseId: item.warehouseId,
    locationCode: item.locationCode,
    locationName: item.locationName,
    locationType: item.locationType,
    zoneCode: item.zoneCode ?? "",
    allowReceive: item.allowReceive,
    allowPick: item.allowPick,
    allowStore: item.allowStore,
    isDefault: item.isDefault,
    status: item.status
  };
}

export function resetPrototypeWarehouseMasterData() {
  localWarehouses = cloneWarehouses(prototypeWarehouses);
  localLocations = cloneLocations(prototypeLocations);
}

function fromWarehouseApiItem(item: WarehouseApiItem): WarehouseMasterDataItem {
  return {
    id: item.id,
    warehouseCode: item.warehouse_code,
    warehouseName: item.warehouse_name,
    warehouseType: item.warehouse_type,
    siteCode: item.site_code,
    address: item.address,
    allowSaleIssue: item.allow_sale_issue,
    allowProdIssue: item.allow_prod_issue,
    allowQuarantine: item.allow_quarantine,
    status: item.status,
    createdAt: item.created_at,
    updatedAt: item.updated_at,
    auditLogId: item.audit_log_id
  };
}

function fromLocationApiItem(item: WarehouseLocationApiItem): WarehouseLocationMasterDataItem {
  return {
    id: item.id,
    warehouseId: item.warehouse_id,
    warehouseCode: item.warehouse_code,
    locationCode: item.location_code,
    locationName: item.location_name,
    locationType: item.location_type,
    zoneCode: item.zone_code,
    allowReceive: item.allow_receive,
    allowPick: item.allow_pick,
    allowStore: item.allow_store,
    isDefault: item.is_default,
    status: item.status,
    createdAt: item.created_at,
    updatedAt: item.updated_at,
    auditLogId: item.audit_log_id
  };
}

function toWarehouseApiQuery(query: WarehouseMasterDataQuery): WarehouseApiQuery {
  return {
    q: query.search,
    status: query.status || undefined,
    warehouse_type: query.warehouseType || undefined,
    page: 1,
    page_size: 100
  };
}

function toLocationApiQuery(query: WarehouseLocationMasterDataQuery): WarehouseLocationApiQuery {
  return {
    q: query.search,
    warehouse_id: query.warehouseId || undefined,
    status: query.status || undefined,
    location_type: query.locationType || undefined,
    page: 1,
    page_size: 100
  };
}

function toWarehouseApiRequest(input: WarehouseMasterDataInput): WarehouseApiCreateRequest {
  return {
    warehouse_code: input.warehouseCode,
    warehouse_name: input.warehouseName,
    warehouse_type: input.warehouseType,
    site_code: input.siteCode,
    address: input.address || undefined,
    allow_sale_issue: input.allowSaleIssue,
    allow_prod_issue: input.allowProdIssue,
    allow_quarantine: input.allowQuarantine,
    status: input.status
  };
}

function toLocationApiRequest(input: WarehouseLocationMasterDataInput): WarehouseLocationApiCreateRequest {
  return {
    warehouse_id: input.warehouseId,
    location_code: input.locationCode,
    location_name: input.locationName,
    location_type: input.locationType,
    zone_code: input.zoneCode || undefined,
    allow_receive: input.allowReceive,
    allow_pick: input.allowPick,
    allow_store: input.allowStore,
    is_default: input.isDefault,
    status: input.status
  };
}

function normalizeWarehouseInput(input: WarehouseMasterDataInput): WarehouseMasterDataInput {
  return {
    ...input,
    warehouseCode: input.warehouseCode.trim().toUpperCase(),
    warehouseName: input.warehouseName.trim(),
    siteCode: input.siteCode.trim().toUpperCase(),
    address: input.address.trim()
  };
}

function normalizeLocationInput(input: WarehouseLocationMasterDataInput): WarehouseLocationMasterDataInput {
  return {
    ...input,
    warehouseId: input.warehouseId.trim(),
    locationCode: input.locationCode.trim().toUpperCase(),
    locationName: input.locationName.trim(),
    zoneCode: input.zoneCode.trim().toUpperCase()
  };
}

function validateWarehouseInput(input: WarehouseMasterDataInput) {
  const missing = [
    ["warehouse code", input.warehouseCode],
    ["warehouse name", input.warehouseName],
    ["site code", input.siteCode]
  ].filter(([, value]) => !String(value).trim());

  if (missing.length > 0) {
    throw new Error(`Missing required fields: ${missing.map(([label]) => label).join(", ")}`);
  }
}

function validateLocationInput(input: WarehouseLocationMasterDataInput) {
  const missing = [
    ["warehouse", input.warehouseId],
    ["location code", input.locationCode],
    ["location name", input.locationName]
  ].filter(([, value]) => !String(value).trim());

  if (missing.length > 0) {
    throw new Error(`Missing required fields: ${missing.map(([label]) => label).join(", ")}`);
  }
  findWarehouse(input.warehouseId);
}

function ensureUniqueWarehouse(input: WarehouseMasterDataInput, currentId = "") {
  if (localWarehouses.some((item) => item.id !== currentId && item.warehouseCode === input.warehouseCode)) {
    throw new Error("Warehouse code already exists");
  }
}

function ensureUniqueLocation(input: WarehouseLocationMasterDataInput, currentId = "") {
  if (
    localLocations.some(
      (item) => item.id !== currentId && item.warehouseId === input.warehouseId && item.locationCode === input.locationCode
    )
  ) {
    throw new Error("Location code already exists for this warehouse");
  }
}

function findWarehouse(warehouseId: string) {
  const warehouse = localWarehouses.find((item) => item.id === warehouseId);
  if (!warehouse) {
    throw new Error("Warehouse location references an invalid warehouse");
  }

  return warehouse;
}

function filterWarehouses(items: WarehouseMasterDataItem[], query: WarehouseMasterDataQuery) {
  const search = query.search?.trim().toLowerCase();
  return sortWarehouses(
    items.filter((item) => {
      if (query.status && item.status !== query.status) {
        return false;
      }
      if (query.warehouseType && item.warehouseType !== query.warehouseType) {
        return false;
      }
      if (!search) {
        return true;
      }

      return [item.warehouseCode, item.warehouseName, item.siteCode, item.address ?? ""].some((value) =>
        value.toLowerCase().includes(search)
      );
    })
  );
}

function filterLocations(items: WarehouseLocationMasterDataItem[], query: WarehouseLocationMasterDataQuery) {
  const search = query.search?.trim().toLowerCase();
  return sortLocations(
    items.filter((item) => {
      if (query.warehouseId && item.warehouseId !== query.warehouseId) {
        return false;
      }
      if (query.status && item.status !== query.status) {
        return false;
      }
      if (query.locationType && item.locationType !== query.locationType) {
        return false;
      }
      if (!search) {
        return true;
      }

      return [item.warehouseCode, item.locationCode, item.locationName, item.zoneCode ?? ""].some((value) =>
        value.toLowerCase().includes(search)
      );
    })
  );
}

function sortWarehouses(items: WarehouseMasterDataItem[]) {
  return [...items].sort((left, right) => {
    if (left.status !== right.status) {
      return left.status === "active" ? -1 : 1;
    }
    if (left.siteCode !== right.siteCode) {
      return left.siteCode.localeCompare(right.siteCode);
    }

    return left.warehouseCode.localeCompare(right.warehouseCode);
  });
}

function sortLocations(items: WarehouseLocationMasterDataItem[]) {
  return [...items].sort((left, right) => {
    if (left.status !== right.status) {
      return left.status === "active" ? -1 : 1;
    }
    if (left.warehouseCode !== right.warehouseCode) {
      return left.warehouseCode.localeCompare(right.warehouseCode);
    }
    if ((left.zoneCode ?? "") !== (right.zoneCode ?? "")) {
      return (left.zoneCode ?? "").localeCompare(right.zoneCode ?? "");
    }

    return left.locationCode.localeCompare(right.locationCode);
  });
}

function cloneWarehouses(items: WarehouseMasterDataItem[]) {
  return items.map((item) => ({ ...item }));
}

function cloneLocations(items: WarehouseLocationMasterDataItem[]) {
  return items.map((item) => ({ ...item }));
}
