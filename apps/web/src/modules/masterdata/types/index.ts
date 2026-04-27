import type { components } from "@/shared/api/generated/schema";

export type ProductStatus = components["schemas"]["MasterDataStatus"];
export type ProductType = components["schemas"]["ItemType"];
export type WarehouseStatus = components["schemas"]["WarehouseStatus"];
export type WarehouseType = components["schemas"]["WarehouseType"];
export type WarehouseLocationStatus = components["schemas"]["LocationStatus"];
export type WarehouseLocationType = components["schemas"]["LocationType"];
export type SupplierStatus = components["schemas"]["SupplierStatus"];
export type SupplierGroup = components["schemas"]["SupplierGroup"];
export type CustomerStatus = components["schemas"]["CustomerStatus"];
export type CustomerType = components["schemas"]["CustomerType"];

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
  standardCost?: string;
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
  standardCost: string;
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

export type SupplierMasterDataItem = {
  id: string;
  supplierCode: string;
  supplierName: string;
  supplierGroup: SupplierGroup;
  contactName?: string;
  phone?: string;
  email?: string;
  taxCode?: string;
  address?: string;
  paymentTerms?: string;
  leadTimeDays?: number;
  moq?: string;
  qualityScore?: string;
  deliveryScore?: string;
  status: SupplierStatus;
  createdAt: string;
  updatedAt: string;
  auditLogId?: string;
};

export type SupplierMasterDataInput = {
  supplierCode: string;
  supplierName: string;
  supplierGroup: SupplierGroup;
  contactName: string;
  phone: string;
  email: string;
  taxCode: string;
  address: string;
  paymentTerms: string;
  leadTimeDays: number;
  moq: string;
  qualityScore: string;
  deliveryScore: string;
  status: SupplierStatus;
};

export type SupplierMasterDataQuery = {
  search?: string;
  status?: SupplierStatus | "";
  supplierGroup?: SupplierGroup | "";
};

export type CustomerMasterDataItem = {
  id: string;
  customerCode: string;
  customerName: string;
  customerType: CustomerType;
  channelCode?: string;
  priceListCode?: string;
  discountGroup?: string;
  creditLimit?: string;
  paymentTerms?: string;
  contactName?: string;
  phone?: string;
  email?: string;
  taxCode?: string;
  address?: string;
  status: CustomerStatus;
  createdAt: string;
  updatedAt: string;
  auditLogId?: string;
};

export type CustomerMasterDataInput = {
  customerCode: string;
  customerName: string;
  customerType: CustomerType;
  channelCode: string;
  priceListCode: string;
  discountGroup: string;
  creditLimit: string;
  paymentTerms: string;
  contactName: string;
  phone: string;
  email: string;
  taxCode: string;
  address: string;
  status: CustomerStatus;
};

export type CustomerMasterDataQuery = {
  search?: string;
  status?: CustomerStatus | "";
  customerType?: CustomerType | "";
};

export type PartyMasterDataSummary = {
  suppliers: number;
  activeSuppliers: number;
  customers: number;
  activeCustomers: number;
};
