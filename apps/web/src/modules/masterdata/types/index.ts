import type { components } from "@/shared/api/generated/schema";

export type ProductStatus = components["schemas"]["MasterDataStatus"];
export type ProductType = components["schemas"]["ItemType"];
export type WarehouseStatus = components["schemas"]["WarehouseStatus"];
export type WarehouseType = components["schemas"]["WarehouseType"];
export type WarehouseLocationStatus = components["schemas"]["LocationStatus"];
export type WarehouseLocationType = components["schemas"]["LocationType"];

export type ProductMasterDataItem = {
  id: string;
  itemCode: string;
  skuCode: string;
  name: string;
  itemType: ProductType;
  itemGroup?: string;
  brandCode?: string;
  uomBase: string;
  uomPurchase?: string;
  uomIssue?: string;
  lotControlled: boolean;
  expiryControlled: boolean;
  shelfLifeDays?: number;
  qcRequired: boolean;
  status: ProductStatus;
  standardCost?: number;
  isSellable: boolean;
  isPurchasable: boolean;
  isProducible: boolean;
  specVersion?: string;
  createdAt: string;
  updatedAt: string;
  auditLogId?: string;
};

export type ProductMasterDataInput = {
  itemCode: string;
  skuCode: string;
  name: string;
  itemType: ProductType;
  itemGroup: string;
  brandCode: string;
  uomBase: string;
  uomPurchase: string;
  uomIssue: string;
  lotControlled: boolean;
  expiryControlled: boolean;
  shelfLifeDays: number;
  qcRequired: boolean;
  status: ProductStatus;
  standardCost: number;
  isSellable: boolean;
  isPurchasable: boolean;
  isProducible: boolean;
  specVersion: string;
};

export type ProductMasterDataQuery = {
  search?: string;
  status?: ProductStatus | "";
  itemType?: ProductType | "";
};

export type ProductMasterDataSummary = {
  total: number;
  active: number;
  draft: number;
  controlled: number;
};

export type WarehouseMasterDataItem = {
  id: string;
  warehouseCode: string;
  warehouseName: string;
  warehouseType: WarehouseType;
  siteCode: string;
  address?: string;
  allowSaleIssue: boolean;
  allowProdIssue: boolean;
  allowQuarantine: boolean;
  status: WarehouseStatus;
  createdAt: string;
  updatedAt: string;
  auditLogId?: string;
};

export type WarehouseMasterDataInput = {
  warehouseCode: string;
  warehouseName: string;
  warehouseType: WarehouseType;
  siteCode: string;
  address: string;
  allowSaleIssue: boolean;
  allowProdIssue: boolean;
  allowQuarantine: boolean;
  status: WarehouseStatus;
};

export type WarehouseMasterDataQuery = {
  search?: string;
  status?: WarehouseStatus | "";
  warehouseType?: WarehouseType | "";
};

export type WarehouseLocationMasterDataItem = {
  id: string;
  warehouseId: string;
  warehouseCode: string;
  locationCode: string;
  locationName: string;
  locationType: WarehouseLocationType;
  zoneCode?: string;
  allowReceive: boolean;
  allowPick: boolean;
  allowStore: boolean;
  isDefault: boolean;
  status: WarehouseLocationStatus;
  createdAt: string;
  updatedAt: string;
  auditLogId?: string;
};

export type WarehouseLocationMasterDataInput = {
  warehouseId: string;
  locationCode: string;
  locationName: string;
  locationType: WarehouseLocationType;
  zoneCode: string;
  allowReceive: boolean;
  allowPick: boolean;
  allowStore: boolean;
  isDefault: boolean;
  status: WarehouseLocationStatus;
};

export type WarehouseLocationMasterDataQuery = {
  search?: string;
  warehouseId?: string;
  status?: WarehouseLocationStatus | "";
  locationType?: WarehouseLocationType | "";
};

export type WarehouseLocationMasterDataSummary = {
  warehouses: number;
  activeWarehouses: number;
  activeLocations: number;
  receivingLocations: number;
};
