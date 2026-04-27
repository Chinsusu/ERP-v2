import type { components } from "@/shared/api/generated/schema";

export type ProductStatus = components["schemas"]["MasterDataStatus"];
export type ProductType = components["schemas"]["ItemType"];

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
