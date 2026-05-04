import { createPurchaseOrder } from "../../purchase/services/purchaseOrderService";
import type { CreatePurchaseOrderInput, PurchaseOrder } from "../../purchase/types";
import { createSubcontractOrder } from "../../subcontract/services/subcontractOrderService";
import type { CreateSubcontractOrderInput, SubcontractOrder } from "../../subcontract/types";
import type { ProductionPlan } from "../types";

export type ProductionPlanPurchaseOrderInput = {
  supplierId: string;
  warehouseId: string;
  expectedDate: string;
  currencyCode?: string;
  unitPrice?: string;
};

export type ProductionPlanSubcontractOrderInput = {
  factoryId: string;
  expectedDeliveryDate: string;
  sampleRequired?: boolean;
  depositStatus?: CreateSubcontractOrderInput["depositStatus"];
  materialUnitCost?: string;
};

export function buildPurchaseOrderFromProductionPlan(
  plan: ProductionPlan,
  input: ProductionPlanPurchaseOrderInput
): CreatePurchaseOrderInput {
  const lines = plan.purchaseRequestDraft.lines;
  if (lines.length === 0) {
    throw new Error("Production plan has no purchase request draft lines");
  }

  return {
    supplierId: input.supplierId,
    warehouseId: input.warehouseId,
    expectedDate: input.expectedDate,
    currencyCode: input.currencyCode ?? "VND",
    note: `Tạo từ kế hoạch sản xuất ${plan.planNo}${
      plan.purchaseRequestDraft.requestNo ? ` / ${plan.purchaseRequestDraft.requestNo}` : ""
    }`,
    lines: lines.map((line) => {
      if (!line.itemId) {
        throw new Error(`Purchase draft line ${line.sku} has no item id`);
      }

      return {
        itemId: line.itemId,
        lineNo: line.lineNo,
        orderedQty: line.requestedQty,
        uomCode: line.uomCode,
        unitPrice: input.unitPrice ?? "0",
        currencyCode: input.currencyCode ?? "VND",
        expectedDate: input.expectedDate,
        note: `Từ ${plan.planNo} dòng ${line.lineNo}: ${line.sku}`
      };
    })
  };
}

export async function createPurchaseOrderFromProductionPlan(
  plan: ProductionPlan,
  input: ProductionPlanPurchaseOrderInput
): Promise<PurchaseOrder> {
  return createPurchaseOrder(buildPurchaseOrderFromProductionPlan(plan, input));
}

export function buildSubcontractOrderFromProductionPlan(
  plan: ProductionPlan,
  input: ProductionPlanSubcontractOrderInput
): CreateSubcontractOrderInput {
  if (plan.lines.some((line) => line.needsPurchase || Number(line.shortageQty) > 0)) {
    throw new Error("Production plan still has material shortages");
  }

  const materialLines = plan.lines
    .filter((line) => line.isStockManaged)
    .map((line) => {
      if (!line.componentItemId) {
        throw new Error(`Production plan material line ${line.componentSku} has no item id`);
      }

      return {
        itemId: line.componentItemId,
        skuCode: line.componentSku,
        itemName: line.componentName,
        plannedQty: line.requiredStockBaseQty,
        uomCode: line.stockBaseUomCode,
        unitCost: input.materialUnitCost ?? "0",
        currencyCode: "VND",
        lotTraceRequired: true,
        note: `Từ ${plan.planNo} dòng ${line.lineNo}`
      };
    });

  const [firstLine] = materialLines;
  if (!firstLine) {
    throw new Error("Production plan has no stock-managed material lines");
  }

  return {
    factoryId: input.factoryId,
    productId: plan.outputItemId,
    productName: plan.outputItemName,
    quantity: Number(plan.plannedQty),
    uomCode: plan.uomCode,
    specVersion: `${plan.formulaCode} ${plan.formulaVersion} / ${plan.planNo}`,
    sampleRequired: input.sampleRequired ?? true,
    expectedDeliveryDate: input.expectedDeliveryDate,
    depositStatus: input.depositStatus ?? "pending",
    materialItemId: firstLine.itemId,
    materialQty: firstLine.plannedQty,
    materialUnitCost: firstLine.unitCost,
    materialLines
  };
}

export async function createSubcontractOrderFromProductionPlan(
  plan: ProductionPlan,
  input: ProductionPlanSubcontractOrderInput
): Promise<SubcontractOrder> {
  return createSubcontractOrder(buildSubcontractOrderFromProductionPlan(plan, input));
}
