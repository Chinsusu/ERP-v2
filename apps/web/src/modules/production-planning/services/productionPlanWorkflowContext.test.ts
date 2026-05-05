import { describe, expect, it } from "vitest";
import { buildProductionPlanWorkflowContext, productionPlanWorkflowSteps } from "./productionPlanWorkflowContext";
import type { ProductionPlan } from "../types";

describe("productionPlanWorkflowContext", () => {
  it("builds selected-plan context for the one-page purchase order flow", () => {
    const context = buildProductionPlanWorkflowContext(shortagePlan);

    expect(productionPlanWorkflowSteps.map((step) => `Bước ${step.number}: ${step.label}`)).toEqual([
      "Bước 1: Chọn kế hoạch sản xuất",
      "Bước 2: Tính nhu cầu vật tư",
      "Bước 3: Tạo PO",
      "Bước 4: Tạo lệnh gia công"
    ]);
    expect(context.planLabel).toBe("PP-260504-0001 - XFF - 162 PCS");
    expect(context.outputLabel).toBe("XFF - Tinh chat buoi Fast & Furious 150ML");
    expect(context.formulaLabel).toBe("XFF-150ML - v1");
    expect(context.materialStatusLabel).toBe("Thiếu 1 dòng vật tư");
    expect(context.materialStatusTone).toBe("warning");
    expect(context.purchaseTitle).toBe("Tạo PO từ PP-260504-0001");
    expect(context.purchaseSummary).toBe("1 dòng vật tư cần mua cho XFF - 162 PCS.");
    expect(context.purchaseButtonLabel).toBe("Tạo PO từ kế hoạch này");
    expect(context.subcontractTitle).toBe("Tạo lệnh gia công từ PP-260504-0001");
    expect(context.subcontractSummary).toBe("Còn 1 dòng thiếu vật tư, cần xử lý mua hàng trước.");
  });

  it("marks a selected plan as ready for subcontract when materials are available", () => {
    const context = buildProductionPlanWorkflowContext(availablePlan);

    expect(context.materialStatusLabel).toBe("Đủ vật tư");
    expect(context.materialStatusTone).toBe("success");
    expect(context.purchaseSummary).toBe("Kế hoạch này không có dòng đề nghị mua.");
    expect(context.subcontractSummary).toBe("Đủ vật tư để tạo lệnh gia công từ PP-260504-0002.");
  });
});

const basePlan = {
  id: "plan-001",
  orgId: "org-my-pham",
  planNo: "PP-260504-0001",
  outputItemId: "item-xff-150",
  outputSku: "XFF",
  outputItemName: "Tinh chat buoi Fast & Furious 150ML",
  outputItemType: "finished_good",
  plannedQty: "162.000000",
  uomCode: "PCS",
  formulaId: "formula-xff-v1",
  formulaCode: "XFF-150ML",
  formulaVersion: "v1",
  formulaBatchQty: "1.000000",
  formulaBatchUomCode: "PCS",
  plannedStartDate: "2026-05-10",
  plannedEndDate: "2026-05-12",
  status: "purchase_request_draft_created",
  auditLogId: "audit-production-plan-001",
  createdAt: "2026-05-04T03:00:00Z",
  createdBy: "user-production",
  updatedAt: "2026-05-04T03:00:00Z",
  updatedBy: "user-production",
  version: 1
} satisfies Omit<ProductionPlan, "lines" | "purchaseRequestDraft">;

const shortagePlan: ProductionPlan = {
  ...basePlan,
  lines: [
    {
      id: "pp-line-plan-001-001",
      formulaLineId: "formula-line-001",
      lineNo: 1,
      componentItemId: "item-act-baicapil",
      componentSku: "ACT_BAICAPIL",
      componentName: "BAICAPIL",
      componentType: "raw_material",
      formulaQty: "1.000000",
      formulaUomCode: "G",
      requiredQty: "162.000000",
      requiredUomCode: "G",
      requiredStockBaseQty: "0.162000",
      stockBaseUomCode: "KG",
      availableQty: "0.000500",
      shortageQty: "0.161500",
      purchaseDraftQty: "0.161500",
      purchaseDraftUomCode: "KG",
      isStockManaged: true,
      needsPurchase: true
    }
  ],
  purchaseRequestDraft: {
    id: "pr-draft-001",
    requestNo: "PR-DRAFT-260504-0001",
    sourceProductionPlanId: "plan-001",
    sourceProductionPlanNo: "PP-260504-0001",
    status: "draft",
    lines: [
      {
        id: "pr-line-001",
        lineNo: 1,
        sourceProductionPlanLineId: "pp-line-plan-001-001",
        itemId: "item-act-baicapil",
        sku: "ACT_BAICAPIL",
        itemName: "BAICAPIL",
        requestedQty: "0.161500",
        uomCode: "KG"
      }
    ]
  }
};

const availablePlan: ProductionPlan = {
  ...basePlan,
  id: "plan-002",
  planNo: "PP-260504-0002",
  status: "draft",
  lines: [
    {
      id: "pp-line-plan-002-001",
      formulaLineId: "formula-line-001",
      lineNo: 1,
      componentItemId: "item-act-baicapil",
      componentSku: "ACT_BAICAPIL",
      componentName: "BAICAPIL",
      componentType: "raw_material",
      formulaQty: "1.000000",
      formulaUomCode: "G",
      requiredQty: "162.000000",
      requiredUomCode: "G",
      requiredStockBaseQty: "0.162000",
      stockBaseUomCode: "KG",
      availableQty: "1.000000",
      shortageQty: "0.000000",
      purchaseDraftQty: "0.000000",
      purchaseDraftUomCode: "KG",
      isStockManaged: true,
      needsPurchase: false
    }
  ],
  purchaseRequestDraft: { lines: [] }
};
